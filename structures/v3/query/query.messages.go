package query

import (
	"context"
	"io"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func (q *Query) ModRequestMessages(ctx context.Context, opt ModRequestMessagesQueryOptions) ([]*structures.Message, error) {
	actor := opt.Actor
	// sign-in is required
	targets := opt.Targets
	if !opt.SkipPermissionCheck {
		if actor == nil {
			return nil, errors.ErrUnauthorized()
		}

		// check permissions for targets
		if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
			targets[structures.ObjectKindEmote] = false
		}
		if !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
			targets[structures.ObjectKindEmoteSet] = false
		}
		if !actor.HasPermission(structures.RolePermissionReportCreate) {
			targets[structures.ObjectKindReport] = false
		}
	}
	targetsAry := []structures.ObjectKind{}
	for k, ok := range targets {
		if ok {
			targetsAry = append(targetsAry, k)
		}
	}
	return q.messages(ctx, bson.M{
		"kind": structures.MessageKindModRequest,
		"read": false,
	}, messageQueryOptions{
		Actor: actor,
		Limit: 100,
		SubFilter: bson.M{
			"kind": structures.MessageKindModRequest,
			"data.target_kind": bson.M{
				"$in": targetsAry,
			},
		},
	})
}

func (q *Query) messages(ctx context.Context, filter bson.M, opt messageQueryOptions) ([]*structures.Message, error) {
	items := []*structures.Message{}

	// Set limit?
	limit := mongo.Pipeline{}
	if opt.Limit != 0 {
		limit = append(limit, bson.D{{Key: "$limit", Value: opt.Limit}})
	}
	// Set sub-filter
	// this is an additional match operation onto message objects, rather than readstates
	// allowing filter on message data and top-level message properties
	subFilter := bson.M{"$expr": bson.M{
		"$eq": bson.A{"$$msg_id", "$_id"},
	}}
	if len(opt.SubFilter) > 0 {
		for k, v := range opt.SubFilter {
			subFilter[k] = v
		}
	}

	// Create the pipeline
	cur, err := q.mongo.Collection(mongo.CollectionNameMessagesRead).Aggregate(ctx, aggregations.Combine(
		// Search message read states
		mongo.Pipeline{
			{{Key: "$sort", Value: bson.M{"_id": -1}}},
			{{Key: "$match", Value: filter}},
		},
		limit,
		mongo.Pipeline{
			{{
				Key: "$lookup",
				Value: mongo.LookupWithPipeline{
					From: mongo.CollectionNameMessages,
					Let:  bson.M{"msg_id": "$message_id", "read": "$read"},
					Pipeline: aggregations.CombinePtr(
						mongo.Pipeline{{{
							Key:   "$match",
							Value: subFilter,
						}}},
						opt.SubPipeline,
					),

					As: "messages",
				},
			}},
			{{ // Collect message author users
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameUsers,
					LocalField:   "messages.author_id",
					ForeignField: "_id",
					As:           "authors",
				},
			}},
			{{
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameEntitlements,
					LocalField:   "authors._id",
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
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	v := &aggregatedMessagesResult{}
	cur.Next(ctx)
	if err := cur.Decode(v); err != nil {
		if err == io.EOF {
			return nil, errors.ErrNoItems().SetDetail("No messages")
		}
		logrus.WithError(err).Error("mongo, failed to decode aggregated result of mod requests query")
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	qb := &QueryBinder{ctx, q}
	userMap := qb.mapUsers(v.Authors, v.RoleEntitlements...)

	for _, msg := range v.Messages {
		msg.Author = userMap[msg.AuthorID]
		items = append(items, msg)
	}

	return items, nil
}

type ModRequestMessagesQueryOptions struct {
	Actor               *structures.User
	Targets             map[structures.ObjectKind]bool
	SkipPermissionCheck bool
}

type messageQueryOptions struct {
	Actor       *structures.User
	Limit       int
	SubFilter   bson.M
	SubPipeline mongo.Pipeline
}

type aggregatedMessagesResult struct {
	Messages         []*structures.Message     `bson:"messages"`
	Authors          []*structures.User        `bson:"authors"`
	RoleEntitlements []*structures.Entitlement `bson:"role_entitlements"`
}
