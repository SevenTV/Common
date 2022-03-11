package query

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/redis"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/Common/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (q *Query) EmoteChannels(ctx context.Context, emoteID primitive.ObjectID, page int, limit int) ([]*structures.User, int64, error) {
	// Emote Sets that have this emote
	setIDs := []primitive.ObjectID{}

	// Ping redis for a cached value
	rKey := q.redis.ComposeKey("gql-v3", fmt.Sprintf("emote:%s:active_sets", emoteID.Hex()))
	v, err := q.redis.Get(ctx, rKey)
	if err == nil && v != "" {
		if err = json.Unmarshal(utils.S2B(v), &setIDs); err != nil {
			logrus.WithError(err).Error("couldn't decode emote's active set ids")
		}
	} else {
		cur, err := q.mongo.Collection(mongo.CollectionNameEmoteSets).Find(ctx, bson.M{"emotes.id": emoteID}, options.Find().SetProjection(bson.M{"owner_id": 1}))
		if err != nil {
			return nil, 0, err
		}
		for i := 0; cur.Next(ctx); i++ {
			v := &structures.EmoteSet{}
			if err = cur.Decode(v); err != nil {
				logrus.WithError(err).Error("mongo, couldn't decode into EmoteSet")
			}
			setIDs = append(setIDs, v.ID)
		}

		// Set in redis
		b, err := json.Marshal(setIDs)
		if err = multierror.Append(err, q.redis.SetEX(ctx, rKey, utils.B2S(b), time.Hour*6)).ErrorOrNil(); err != nil {
			logrus.WithError(err).Error("failed to cache set ids in redis")
		}
	}

	// Fetch users with this set active
	match := bson.M{
		"connections.emote_set_id": bson.M{
			"$in": setIDs,
		},
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	count := int64(0)
	go func() { // Get the total channel count
		defer wg.Done()
		k := q.redis.ComposeKey("gql-v3", fmt.Sprintf("emote:%s:channel_count", emoteID.Hex()))

		count, err = q.redis.RawClient().Get(ctx, k.String()).Int64()
		if err == redis.Nil { // query if not cached
			count, _ = q.mongo.Collection(mongo.CollectionNameUsers).CountDocuments(ctx, match)
			_ = q.redis.SetEX(ctx, k, count, time.Hour*6)
		}
	}()
	cur, err := q.mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, aggregations.Combine(
		mongo.Pipeline{
			{{
				Key:   "$match",
				Value: match,
			}},
			{{
				Key:   "$sort",
				Value: bson.D{{Key: "metadata.role_position", Value: -1}},
			}},
			{{Key: "$skip", Value: (page - 1) * limit}},
			{{
				Key:   "$limit",
				Value: limit,
			}},
			{{
				Key:   "$sort",
				Value: bson.D{{Key: "metadata.role_position", Value: -1}, {Key: "username", Value: 1}},
			}},
		},
		aggregations.UserRelationRoles,
	))
	if err != nil {
		return nil, count, err
	}
	users := []*structures.User{}
	if err = cur.All(ctx, &users); err != nil {
		return nil, count, err
	}
	wg.Wait()

	return users, count, nil
}
