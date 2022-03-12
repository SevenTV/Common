package query

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
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
)

const EMOTES_QUERY_LIMIT = 300

func (q *Query) SearchEmotes(ctx context.Context, opt SearchEmotesOptions) ([]*structures.Emote, int, error) {
	// Define limit (how many emotes can be returned in a single query)
	limit := opt.Limit
	if limit > EMOTES_QUERY_LIMIT {
		limit = EMOTES_QUERY_LIMIT
	} else if limit < 1 {
		limit = 1
	}

	// Define page
	page := 1
	if opt.Page > page {
		page = opt.Page
	} else if opt.Page < 1 {
		page = 1
	}

	// Define default filter
	filter := opt.Filter
	if filter == nil {
		filter = &SearchEmotesFilter{
			CaseSensitive: utils.BoolPointer(false),
			ExactMatch:    utils.BoolPointer(false),
			IgnoreTags:    utils.BoolPointer(false),
			Document:      bson.M{},
		}
	}

	// Define the query string
	query := strings.Trim(opt.Query, " ")

	// Set up db query
	match := bson.D{{Key: "versions.0.state.lifecycle", Value: structures.EmoteLifecycleLive}}
	if len(filter.Document) > 0 {
		for k, v := range filter.Document {
			match = append(match, bson.E{Key: k, Value: v})
		}
	}

	// Apply permission checks
	// omit unlisted/private emotes
	privileged := int(1)
	if opt.Actor == nil || !opt.Actor.HasPermission(structures.RolePermissionEditAnyEmote) {
		privileged = 0
		match = append(match, bson.E{
			Key: "flags",
			Value: bson.M{
				"$bitsAllClear": structures.EmoteFlagsPrivate,
				"$bitsAllSet":   structures.EmoteFlagsListed,
			},
		})
	}

	// Define the pipeline
	pipeline := mongo.Pipeline{}

	// Apply name/tag query
	h := sha256.New()
	h.Write(utils.S2B(query))
	h.Write([]byte{byte(privileged)})
	if len(filter.Document) > 0 {
		optBytes, _ := json.Marshal(filter.Document)
		h.Write(optBytes)
	}

	queryKey := q.redis.ComposeKey("common", fmt.Sprintf("emote-search:%s", hex.EncodeToString((h.Sum(nil)))))
	cpargs := bson.A{}

	// Handle exact match
	if filter.ExactMatch != nil && *filter.ExactMatch {
		// For an exact mathc we will use the $text operator
		// rather than $indexOfCP because name/tags are indexed fields
		match = append(match, bson.E{Key: "$text", Value: bson.M{
			"$search":        query,
			"$caseSensitive": filter.CaseSensitive != nil && *filter.CaseSensitive,
		}})
		pipeline = append(pipeline, []bson.D{
			{{Key: "$match", Value: match}},
			{{Key: "$sort", Value: bson.M{"score": bson.M{"$meta": "textScore"}}}},
		}...)
	} else {
		or := bson.A{}
		if filter.CaseSensitive != nil && *filter.CaseSensitive {
			cpargs = append(cpargs, "$name", query)
		} else {
			cpargs = append(cpargs, bson.M{"$toLower": "$name"}, strings.ToLower(query))
		}

		or = append(or, bson.M{
			"$expr": bson.M{
				"$gt": bson.A{bson.M{"$indexOfCP": cpargs}, -1},
			},
		})

		// Add tag search
		if filter.IgnoreTags == nil || !*filter.IgnoreTags {
			or = append(or, bson.M{
				"$expr": bson.M{
					"$gt": bson.A{
						bson.M{"$indexOfCP": bson.A{bson.M{"$reduce": bson.M{
							"input":        "$tags",
							"initialValue": " ",
							"in":           bson.M{"$concat": bson.A{"$$value", "$$this"}},
						}}, strings.ToLower(query)}},
						-1,
					},
				},
			})
		}

		match = append(match, bson.E{Key: "$or", Value: or})
		if opt.Sort != nil && len(opt.Sort) > 0 {
			pipeline = append(pipeline, bson.D{
				{Key: "$sort", Value: opt.Sort},
			})
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// Complete the pipeline
	totalCount, countErr := q.redis.RawClient().Get(ctx, string(queryKey)).Int()
	wg := sync.WaitGroup{}
	wg.Add(1)
	if countErr == redis.Nil {
		go func() { // Run a separate pipeline to return the total count that could be paginated
			defer wg.Done()
			cur, err := q.mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, aggregations.Combine(
				pipeline,
				mongo.Pipeline{
					{{Key: "$count", Value: "count"}},
				}),
			)
			result := make(map[string]int, 1)
			if err == nil {
				cur.Next(ctx)
				if err = multierror.Append(cur.Decode(&result), cur.Close(ctx)).ErrorOrNil(); err != nil {
					logrus.WithError(err).Error("mongo, couldn't count")
				}
			}

			// Return total count & cache
			totalCount = result["count"]
			dur := utils.Ternary(query == "", time.Minute*10, time.Hour*1).(time.Duration)
			if err = q.redis.SetEX(ctx, queryKey, totalCount, dur); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"key":   queryKey,
					"count": totalCount,
				}).Error("redis, failed to save total list count of emotes() gql query")
			}
		}()
	} else {
		wg.Done()
	}

	// Paginate and fetch the relevant emotes
	result := []*structures.Emote{}
	cur, err := q.mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, aggregations.Combine(
		pipeline,
		mongo.Pipeline{
			{{Key: "$skip", Value: (page - 1) * limit}},
			{{Key: "$limit", Value: limit}},
		},
		aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{Roles: true}),
	))
	if err == nil {
		if err = cur.All(ctx, &result); err != nil {
			logrus.WithError(err).Error("mongo, failed to fetch emotes")
		}
	}
	wg.Wait() // wait for total count to finish

	return result, totalCount, nil
}

type SearchEmotesOptions struct {
	Query  string
	Page   int
	Limit  int
	Filter *SearchEmotesFilter
	Sort   bson.M
	Actor  *structures.User
}

type SearchEmotesFilter struct {
	CaseSensitive *bool  `json:"cs"`
	ExactMatch    *bool  `json:"exm"`
	IgnoreTags    *bool  `json:"ignt"`
	Document      bson.M `json:"doc"`
}
