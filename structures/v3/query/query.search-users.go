package query

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/redis"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (q *Query) SearchUsers(ctx context.Context, filter bson.M, opts ...UserSearchOptions) ([]structures.User, int, error) {
	mx := q.lock("SearchUsers")
	defer mx.Unlock()
	items := []structures.User{}

	paginate := mongo.Pipeline{}
	search := len(opts) > 0 && opts[0].Page != 0
	if search {
		opt := opts[0]
		sort := bson.M{"_id": -1}
		if len(opt.Sort) > 0 {
			sort = opt.Sort
		}
		paginate = append(paginate, []bson.D{
			{{Key: "$sort", Value: sort}},
			{{Key: "$skip", Value: (opt.Page - 1) * opt.Limit}},
			{{Key: "$limit", Value: opt.Limit}},
		}...)
		if opt.Query != "" {
			filter["$expr"] = bson.M{
				"$gt": bson.A{
					bson.M{"$indexOfCP": bson.A{
						bson.M{"$toLower": "$username"},
						strings.ToLower(opt.Query),
					}},
					-1,
				},
			}
		}
	}

	b, _ := bson.Marshal(filter)
	h := sha256.New()
	h.Write(b)
	queryKey := q.redis.ComposeKey("common", fmt.Sprintf("user-search:%s", hex.EncodeToString(h.Sum(nil))))

	bans := q.Bans(ctx, BanQueryOptions{ // remove emotes made by usersa who own nothing and are happy
		Filter: bson.M{"effects": bson.M{"$bitsAnySet": structures.BanEffectMemoryHole}},
	})
	cur, err := q.mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, aggregations.Combine(
		mongo.Pipeline{
			{{
				Key:   "$match",
				Value: filter,
			}},
			{{
				Key: "$set",
				Value: bson.M{ // Remove memory holed editors
					"editors": bson.M{"$filter": bson.M{
						"input": "$editors",
						"as":    "e",
						"cond":  bson.M{"$not": bson.M{"$in": bson.A{"$$e.id", bans.MemoryHole.KeySlice()}}},
					}},
				},
			}},
		},
		paginate,
		mongo.Pipeline{
			{{
				Key: "$group",
				Value: bson.M{
					"_id": nil,
					"users": bson.M{
						"$push": "$$ROOT",
					},
				},
			}},
			{{
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameEntitlements,
					LocalField:   "users._id",
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
		logrus.WithError(err).Error("query, failed to execute find")
		return items, 0, err
	}

	// Count the documents
	totalCount, countErr := q.redis.RawClient().Get(ctx, queryKey.String()).Int()
	if search {
		wg := sync.WaitGroup{}
		wg.Add(1)
		if countErr == redis.Nil {
			go func() {
				defer wg.Done()
				cur, err := q.mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, aggregations.Combine(
					mongo.Pipeline{
						{{Key: "$match", Value: filter}},
					},
					mongo.Pipeline{
						{{Key: "$count", Value: "count"}},
						{{Key: "$project", Value: bson.M{"count": "$count"}}},
					},
				))
				result := make(map[string]int, 1)
				if err == nil {
					if ok := cur.Next(ctx); ok {
						if err = cur.Decode(&result); err != nil {
							logrus.WithError(err).Error("mongo, couldn't count users")
						}
					}
					_ = cur.Close(ctx)
				}
				totalCount = result["count"]
				if err = q.redis.SetEX(ctx, queryKey, totalCount, time.Minute*1); err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"key":   queryKey,
						"count": totalCount,
					}).Error("redis, failed to save total list count of \"Search Users\" query")
				}
			}()
		} else {
			wg.Done()
		}
		wg.Wait()
	}

	// Get roles
	roles, _ := q.Roles(ctx, bson.M{})
	roleMap := make(map[primitive.ObjectID]structures.Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	// Map all objects
	if ok := cur.Next(ctx); !ok {
		return items, 0, nil // nothing found!
	}
	v := &aggregatedUsersResult{}
	if err = cur.Decode(v); err != nil {
		return items, 0, err
	}

	userMap := make(map[primitive.ObjectID]structures.User)
	entRoleMap := make(map[primitive.ObjectID][]primitive.ObjectID)
	for _, ent := range v.RoleEntitlements {
		ent, err := structures.ConvertEntitlement[structures.EntitlementDataRole](ent)
		if err != nil {
			return nil, 0, err
		}
		entRoleMap[ent.UserID] = append(entRoleMap[ent.UserID], ent.Data.ObjectReference)
	}
	for _, user := range v.Users {
		user.RoleIDs = append(user.RoleIDs, entRoleMap[user.ID]...)
		userMap[user.ID] = user
	}

	for _, u := range v.Users { // iterare over users
		// add user roles
		for _, roleID := range u.RoleIDs {
			role, rok := roleMap[roleID]
			if !rok {
				continue
			}
			u.Roles = append(u.Roles, role)
		}
		items = append(items, u)
	}

	if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
		logrus.WithError(err).Error("query, failed to close the cursor")
	}
	return items, totalCount, nil
}

type UserSearchOptions struct {
	Page  int
	Limit int
	Query string
	Sort  bson.M
}
type aggregatedUsersResult struct {
	Users            []structures.User                  `bson:"users"`
	RoleEntitlements []structures.Entitlement[bson.Raw] `bson:"role_entitlements"`
	TotalCount       int                                `bson:"total_count"`
}

func (q *Query) UserEditorOf(ctx context.Context, id primitive.ObjectID) ([]*structures.UserEditor, error) {
	cur, err := q.mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.M{
				"editors.id": id,
			},
		}},
		{{
			Key: "$project",
			Value: bson.M{
				"editor": bson.M{
					"$mergeObjects": bson.A{
						bson.M{"$first": bson.M{"$filter": bson.M{
							"input": "$editors",
							"as":    "ed",
							"cond": bson.M{
								"$eq": bson.A{"$$ed.id", id},
							},
						}}},
						bson.M{"id": "$_id"},
					},
				},
			},
		}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$editor"}}},
	})
	if err != nil {
		logrus.WithError(err).Error("query, failed to spawn aggregation")
		return nil, err
	}

	v := []*structures.UserEditor{}
	if err = cur.All(ctx, &v); err != nil {
		return nil, err
	}

	return v, nil
}
