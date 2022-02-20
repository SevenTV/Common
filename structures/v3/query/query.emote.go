package query

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (q *Query) Emotes(ctx context.Context, filter bson.M) ([]*structures.Emote, error) {
	items := []*structures.Emote{}
	cur, err := q.mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, mongo.Pipeline{
		{{
			Key:   "$match",
			Value: filter,
		}},
		{{
			Key: "$group",
			Value: bson.M{
				"_id": nil,
				"emotes": bson.M{
					"$push": "$$ROOT",
				},
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
		logrus.WithError(err).Error("query, failed to spawn aggregation")
		return nil, err
	}

	// Get roles
	roles, _ := q.Roles(ctx, bson.M{})
	roleMap := make(map[primitive.ObjectID]*structures.Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	cur.Next(ctx)
	v := &aggregatedEmotesResult{}
	if err = cur.Decode(v); err != nil {
		return items, err
	}

	// Map all objects
	emoteMap := make(map[primitive.ObjectID]*structures.Emote)
	ownerMap := make(map[primitive.ObjectID]*structures.User)
	entRoleMap := make(map[primitive.ObjectID][]primitive.ObjectID)
	for _, emote := range v.Emotes {
		emoteMap[emote.ID] = emote
	}
	for _, ent := range v.RoleEntitlements {
		ref := ent.GetData().ReadRole()
		if ref == nil {
			continue
		}
		entRoleMap[ent.UserID] = append(entRoleMap[ent.UserID], ref.ObjectReference)
	}
	for _, user := range v.EmoteOwners {
		user.RoleIDs = append(user.RoleIDs, entRoleMap[user.ID]...)
		ownerMap[user.ID] = user
	}

	var ok bool
	for _, e := range v.Emotes { // iterate over emotes
		// add owner
		if e.Owner, ok = ownerMap[e.OwnerID]; ok && !e.Owner.ID.IsZero() {
			// add owner's roles
			for _, roleID := range e.Owner.RoleIDs {
				role, roleOK := roleMap[roleID]
				if !roleOK {
					continue
				}
				e.Owner.Roles = append(e.Owner.Roles, role)
			}
		}

		items = append(items, e)
	}
	if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
		logrus.WithError(err).Error("query, failed to close the cursor")
	}

	return items, nil
}

type aggregatedEmotesResult struct {
	Emotes           []*structures.Emote       `bson:"emotes"`
	EmoteOwners      []*structures.User        `bson:"emote_owners"`
	RoleEntitlements []*structures.Entitlement `bson:"role_entitlements"`
}
