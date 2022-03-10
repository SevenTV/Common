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

func (q *Query) EmoteSets(ctx context.Context, filter bson.M) ([]*structures.EmoteSet, error) {
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

	cur.Next(ctx)
	v := &aggregatedEmoteSets{}
	if err = cur.Decode(v); err != nil {
		logrus.WithError(err).Error("mongo, failed to decode aggregated emote sets")
		return items, err
	}

	emoteMap := make(map[primitive.ObjectID]*structures.Emote)
	ownerMap := make(map[primitive.ObjectID]*structures.User)
	for _, emote := range v.Emotes {
		for _, ver := range emote.Versions {
			emote := *emote
			emote.ID = ver.ID
			emoteMap[ver.ID] = &emote
		}
	}
	for _, user := range v.EmoteOwners {
		ownerMap[user.ID] = user
	}

	var ok bool
	for _, set := range v.Sets {
		for indEmotes, ae := range set.Emotes {
			if ae.Emote, ok = emoteMap[ae.ID]; ok {
				if ae.Emote.Owner, ok = ownerMap[ae.Emote.OwnerID]; ok {
					for _, roleID := range ae.Emote.Owner.RoleIDs {
						role, roleOK := roleMap[roleID]
						if !roleOK {
							continue
						}
						ae.Emote.Owner.Roles = append(ae.Emote.Owner.Roles, role)
					}
				}
			} else {
				set.Emotes[indEmotes].Emote = structures.DeletedEmote
			}
		}
		items = append(items, set)
	}

	return items, nil
}

type aggregatedEmoteSets struct {
	Sets        []*structures.EmoteSet `bson:"sets"`
	Emotes      []*structures.Emote    `bson:"emotes"`
	EmoteOwners []*structures.User     `bson:"emote_owners"`
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
		},
	))
	if err != nil {
		logrus.WithError(err).Error("mongo, failed to spawn aggregation")
	}

	// Get roles
	roles, _ := q.Roles(ctx, bson.M{})
	roleMap := make(map[primitive.ObjectID]*structures.Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	// Iterate over cursor
	for i := 0; cur.Next(ctx); i++ {
		v := &aggregatedUserEmoteSets{}
		if err = cur.Decode(v); err != nil {
			logrus.WithError(err).Error("mongo, couldn't decode user emote set item")
			continue
		}

		// Map emotes bound to the set
		emoteMap := make(map[primitive.ObjectID]*structures.Emote)
		ownerMap := make(map[primitive.ObjectID]*structures.User)
		for _, emote := range v.Emotes {
			for _, ver := range emote.Versions {
				emote := *emote
				emote.ID = ver.ID
				emoteMap[ver.ID] = &emote
			}
		}
		for _, user := range v.EmoteOwners {
			ownerMap[user.ID] = user
		}

		var ok bool
		for _, set := range v.Sets {
			for _, ae := range set.Emotes {
				if ae.Emote, ok = emoteMap[ae.ID]; ok {
					ae.Emote.ID = ae.ID
					if ae.Emote.Owner, ok = ownerMap[ae.Emote.OwnerID]; ok {
						for _, roleID := range ae.Emote.Owner.RoleIDs {
							role, roleOK := roleMap[roleID]
							if !roleOK {
								continue
							}
							ae.Emote.Owner.Roles = append(ae.Emote.Owner.Roles, role)
						}
					}
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
	UserID      primitive.ObjectID     `bson:"_id"`
	Sets        []*structures.EmoteSet `bson:"sets"`
	Emotes      []*structures.Emote    `bson:"emotes"`
	EmoteOwners []*structures.User     `bson:"emote_owners"`
}
