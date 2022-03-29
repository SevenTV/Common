package query

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (q *Query) EmoteSets(ctx context.Context, filter bson.M) *QueryResult[structures.EmoteSet] {
	qr := &QueryResult[structures.EmoteSet]{}
	items := []*structures.EmoteSet{}
	cur, err := q.mongo.Collection(mongo.CollectionNameEmoteSets).Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{
			Key: "$group",
			Value: bson.M{
				"_id":  nil,
				"sets": bson.M{"$push": "$$ROOT"},
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
			Key: "$lookup",
			Value: mongo.Lookup{
				From:         mongo.CollectionNameUsers,
				LocalField:   "emotes.owner_id",
				ForeignField: "_id",
				As:           "emote_owners",
			},
		}},
		{{
			Key: "$lookup",
			Value: mongo.Lookup{
				From:         mongo.CollectionNameEntitlements,
				LocalField:   "emote_owners._id",
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
	})
	if err != nil {
		logrus.WithError(err).Error("mongo, failed to spawn aggregation")
	}
	// Get roles (to assign to emote owners)
	roles, _ := q.Roles(ctx, bson.M{})
	roleMap := make(map[primitive.ObjectID]*structures.Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	if ok := cur.Next(ctx); !ok {
		return qr.setItems(items) // nothing found!
	}
	v := &aggregatedEmoteSets{}
	if err = cur.Decode(v); err != nil {
		logrus.WithError(err).Error("mongo, failed to decode aggregated emote sets")
		return qr.setItems(items).setError(err)
	}

	qb := &QueryBinder{ctx, q}
	ownerMap := qb.mapUsers(v.EmoteOwners, v.RoleEntitlements...)
	emoteMap := make(map[primitive.ObjectID]*structures.Emote)
	var ok bool
	for _, emote := range v.Emotes {
		emote.Owner = ownerMap[emote.OwnerID]
		for _, ver := range emote.Versions {
			emote := *emote
			emote.ID = ver.ID
			emoteMap[ver.ID] = &emote
		}
	}

	for _, set := range v.Sets {
		for indEmotes, ae := range set.Emotes {
			if ae.Emote, ok = emoteMap[ae.ID]; !ok {
				set.Emotes[indEmotes].Emote = structures.DeletedEmote
			}
		}
		items = append(items, set)
	}

	return qr.setItems(items)
}

type aggregatedEmoteSets struct {
	Sets             []*structures.EmoteSet    `bson:"sets"`
	Emotes           []*structures.Emote       `bson:"emotes"`
	EmoteOwners      []*structures.User        `bson:"emote_owners"`
	RoleEntitlements []*structures.Entitlement `bson:"role_entitlements"`
}

func (q *Query) UserEmoteSets(ctx context.Context, filter bson.M) (map[primitive.ObjectID][]*structures.EmoteSet, error) {
	items := make(map[primitive.ObjectID][]*structures.EmoteSet)
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
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameUsers,
					LocalField:   "emotes.owner_id",
					ForeignField: "_id",
					As:           "emote_owners",
				},
			}},
			{{
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameEntitlements,
					LocalField:   "emote_owners._id",
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
		logrus.WithError(err).Error("mongo, failed to spawn aggregation")
	}

	// Iterate over cursor
	bans := q.Bans(ctx, BanQueryOptions{ // remove emotes made by usersa who own nothing and are happy
		Filter: bson.M{"effects": bson.M{"$bitsAnySet": structures.BanEffectNoOwnership | structures.BanEffectMemoryHole}},
	})
	for i := 0; cur.Next(ctx); i++ {
		v := &aggregatedUserEmoteSets{}
		if err = cur.Decode(v); err != nil {
			logrus.WithError(err).Error("mongo, couldn't decode user emote set item")
			continue
		}

		// Map emotes bound to the set
		qb := &QueryBinder{ctx, q}
		ownerMap := qb.mapUsers(v.EmoteOwners, v.RoleEntitlements...)
		emoteMap := make(map[primitive.ObjectID]*structures.Emote)
		for _, emote := range v.Emotes {
			if _, ok := bans.NoOwnership[emote.OwnerID]; ok {
				continue
			}
			if _, ok := bans.MemoryHole[emote.OwnerID]; ok {
				emote.OwnerID = primitive.NilObjectID
			}
			for _, ver := range emote.Versions {
				emote := *emote
				emote.ID = ver.ID
				emote.Owner = ownerMap[emote.OwnerID]

				emoteMap[ver.ID] = &emote
			}
		}

		var ok bool
		for _, set := range v.Sets {
			for _, ae := range set.Emotes {
				if ae.Emote, ok = emoteMap[ae.ID]; ok {
					ae.Emote.ID = ae.ID
				}
			}
		}
		items[v.UserID] = v.Sets
	}
	if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
		logrus.WithError(err).Error("mongo, failed to close the cursor")
	}
	return items, nil
}

type aggregatedUserEmoteSets struct {
	UserID           primitive.ObjectID        `bson:"_id"`
	Sets             []*structures.EmoteSet    `bson:"sets"`
	Emotes           []*structures.Emote       `bson:"emotes"`
	EmoteOwners      []*structures.User        `bson:"emote_owners"`
	RoleEntitlements []*structures.Entitlement `bson:"role_entitlements"`
}
