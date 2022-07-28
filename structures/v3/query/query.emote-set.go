package query

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/seventv/common/mongo"
	"github.com/seventv/common/structures/v3"
	"github.com/seventv/common/structures/v3/aggregations"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (q *Query) EmoteSets(ctx context.Context, filter bson.M) *QueryResult[structures.EmoteSet] {
	qr := &QueryResult[structures.EmoteSet]{}
	items := []structures.EmoteSet{}

	cur, err := q.mongo.Collection(mongo.CollectionNameEmoteSets).Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{
			Key: "$group",
			Value: bson.M{
				"_id": "$_id",
				"set": bson.M{
					"$first": "$$ROOT",
				},
			},
		}},
		{{
			Key: "$lookup",
			Value: mongo.Lookup{
				From:         mongo.CollectionNameEmotes,
				LocalField:   "set.emotes.id",
				ForeignField: "versions.id",
				As:           "ext_emotes",
			},
		}},
		{{
			Key: "$set",
			Value: bson.M{
				"all_users": bson.M{
					"$setUnion": bson.A{bson.A{"$set.owner_id"}, "$set.emotes.actor_id", "$ext_emotes.owner_id"},
				},
			},
		}},
		{{
			Key: "$lookup",
			Value: mongo.Lookup{
				From:         mongo.CollectionNameUsers,
				LocalField:   "all_users",
				ForeignField: "_id",
				As:           "users",
			},
		}},
		{{
			Key: "$lookup",
			Value: mongo.Lookup{
				From:         mongo.CollectionNameEntitlements,
				LocalField:   "all_users",
				ForeignField: "user_id",
				As:           "role_entitlements",
			},
		}},
		{{Key: "$unset", Value: bson.A{"all_users"}}},
		{{
			Key: "$set",
			Value: bson.M{
				"role_entitlements": bson.M{
					"$filter": bson.M{
						"input": "$role_entitlements",
						"as":    "ent",
						"cond": bson.M{
							"$eq": bson.A{"$$ent.kind", structures.EntitlementKindRole},
						},
					},
				},
			},
		}},
	})
	if err != nil {
		return qr.setError(err)
	}

	// Get roles (to assign to emote owners)
	roles, _ := q.Roles(ctx, bson.M{})
	roleMap := make(map[primitive.ObjectID]structures.Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	for cur.Next(ctx) {
		v := &aggregatedEmoteSets{}
		if err = cur.Decode(v); err != nil {
			return qr.setItems(items).setError(err)
		}

		qb := &QueryBinder{ctx, q}
		userMap, err := qb.MapUsers(v.Users, v.RoleEntitlements...)
		if err != nil {
			return qr.setError(err)
		}

		emoteMap := make(map[primitive.ObjectID]structures.Emote)
		for _, emote := range v.Emotes {
			owner := userMap[emote.OwnerID]
			if !owner.ID.IsZero() {
				emote.Owner = &owner
			}
			for _, ver := range emote.Versions {
				emote.ID = ver.ID
				emoteMap[ver.ID] = emote
			}
		}

		owner := userMap[v.Set.OwnerID]
		if !owner.ID.IsZero() {
			v.Set.Owner = &owner
		}
		for indEmotes, ae := range v.Set.Emotes {
			if emote, ok := emoteMap[ae.ID]; !ok {
				v.Set.Emotes[indEmotes].Emote = &structures.DeletedEmote
			} else {
				v.Set.Emotes[indEmotes].Emote = &emote
			}

			// Apply actor user to active emote data?
			if ae.ActorID.IsZero() {
				continue
			}

			if actor, ok := userMap[ae.ActorID]; ok {
				v.Set.Emotes[indEmotes].Actor = &actor
			}
		}
		items = append(items, v.Set)
	}

	return qr.setItems(items)
}

type aggregatedEmoteSets struct {
	Set              structures.EmoteSet                `bson:"set"`
	Emotes           []structures.Emote                 `bson:"ext_emotes"`
	Users            []structures.User                  `bson:"users"`
	RoleEntitlements []structures.Entitlement[bson.Raw] `bson:"role_entitlements"`
}

func (q *Query) UserEmoteSets(ctx context.Context, filter bson.M) (map[primitive.ObjectID][]structures.EmoteSet, error) {
	items := make(map[primitive.ObjectID][]structures.EmoteSet)
	cur, err := q.mongo.Collection(mongo.CollectionNameEmoteSets).Aggregate(ctx, aggregations.Combine(
		mongo.Pipeline{
			{{
				Key:   "$match",
				Value: filter,
			}},
			{{
				Key: "$group",
				Value: bson.M{
					"_id": "$owner_id",
					"sets": bson.M{
						"$push": "$$ROOT",
					},
				},
			}},
			{{
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameEmotes,
					LocalField:   "sets.emotes.id",
					ForeignField: "versions.id",
					As:           "emotes",
				},
			}},
			{{
				Key: "$set",
				Value: bson.M{
					"all_users": bson.M{
						"$setUnion": bson.A{"$sets.owner_id", "$sets.emotes.actor_id", "$sets.emotes.owner_id"},
					},
				},
			}},
			{{
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameUsers,
					LocalField:   "all_users",
					ForeignField: "_id",
					As:           "users",
				},
			}},
			{{
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameEntitlements,
					LocalField:   "all_users",
					ForeignField: "user_id",
					As:           "role_entitlements",
				},
			}},
			{{
				Key: "$set",
				Value: bson.M{
					"role_entitlements": bson.M{
						"$filter": bson.M{
							"input": "$role_entitlements",
							"as":    "ent",
							"cond": bson.M{
								"$eq": bson.A{"$$ent.kind", structures.EntitlementKindRole},
							},
						},
					},
				},
			}},
		},
	))
	if err != nil {
		return nil, err
	}

	// Iterate over cursor
	bans, err := q.Bans(ctx, BanQueryOptions{ // remove emotes made by usersa who own nothing and are happy
		Filter: bson.M{"effects": bson.M{"$bitsAnySet": structures.BanEffectNoOwnership | structures.BanEffectMemoryHole}},
	})
	if err != nil {
		return nil, err
	}

	for i := 0; cur.Next(ctx); i++ {
		v := &aggregatedUserEmoteSets{}
		if err = cur.Decode(v); err != nil {
			continue
		}

		// Map emotes bound to the set
		qb := &QueryBinder{ctx, q}
		userMap, err := qb.MapUsers(v.Users, v.RoleEntitlements...)
		if err != nil {
			return nil, err
		}

		emoteMap := make(map[primitive.ObjectID]structures.Emote)
		for _, emote := range v.Emotes {
			if _, ok := bans.NoOwnership[emote.OwnerID]; ok {
				continue
			}
			if _, ok := bans.MemoryHole[emote.OwnerID]; ok {
				emote.OwnerID = primitive.NilObjectID
			}
			for _, ver := range emote.Versions {
				emote.ID = ver.ID

				owner := userMap[emote.OwnerID]
				if !owner.ID.IsZero() {
					emote.Owner = &owner
				}

				emoteMap[ver.ID] = emote
			}
		}

		for idx, set := range v.Sets {
			for idx, ae := range set.Emotes {
				if emote, ok := emoteMap[ae.ID]; ok {
					emote.ID = ae.ID
					ae.Emote = &emote
					set.Emotes[idx] = ae
				}

				// Apply actor user to active emote data?
				if ae.ActorID.IsZero() {
					continue
				}

				if actor, ok := userMap[ae.ActorID]; ok {
					set.Emotes[idx].Actor = &actor
				}
			}
			v.Sets[idx] = set
		}
		items[v.UserID] = v.Sets
	}
	return items, multierror.Append(err, cur.Close(ctx)).ErrorOrNil()
}

type aggregatedUserEmoteSets struct {
	UserID           primitive.ObjectID                 `bson:"_id"`
	Sets             []structures.EmoteSet              `bson:"sets"`
	Emotes           []structures.Emote                 `bson:"emotes"`
	Users            []structures.User                  `bson:"users"`
	RoleEntitlements []structures.Entitlement[bson.Raw] `bson:"role_entitlements"`
}
