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

	bans := q.Bans(ctx, BanQueryOptions{
		Filter: bson.M{"effects": bson.M{"$bitsAnySet": structures.BanEffectNoOwnership | structures.BanEffectMemoryHole}},
	})
	cur, err := q.mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, mongo.Pipeline{
		{{
			Key:   "$match",
			Value: bson.M{"owner_id": bson.M{"$not": bson.M{"$in": bans.NoOwnership.KeySlice()}}},
		}},
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

	cur.Next(ctx)
	v := &aggregatedEmotesResult{}
	if err = cur.Decode(v); err != nil {
		return items, err
	}

	// Map all objects
	qb := &QueryBinder{ctx, q}
	ownerMap := qb.mapUsers(v.EmoteOwners, v.RoleEntitlements...)

	for _, e := range v.Emotes { // iterate over emotes
		// add owner
		if _, banned := bans.MemoryHole[e.OwnerID]; banned {
			e.OwnerID = primitive.NilObjectID
		} else {
			e.Owner = ownerMap[e.OwnerID]
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
