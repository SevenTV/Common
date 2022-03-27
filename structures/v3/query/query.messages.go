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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (q *Query) InboxMessages(ctx context.Context, opt InboxMessagesQueryOptions) *QueryResult[structures.Message] {
	qr := &QueryResult[structures.Message]{}
	actor := opt.Actor
	user := opt.User
	if user == nil {
		return qr.setError(errors.ErrInternalServerError().SetDetail("no user passed to Inbox query"))
	}

	if !opt.SkipPermissionCheck {
		if actor == nil {
			return qr.setError(errors.ErrUnauthorized())
		}

		// Actor is not the target user
		if actor.ID != user.ID {
			ed, ok, _ := user.GetEditor(actor.ID)
			// Actor is not editor of target user
			if !ok {
				return qr.setError(errors.ErrInsufficientPrivilege().SetDetail("You are not an editor of this user"))
			}
			// Actor is an editor, but does not have the permission to do this
			if !ed.HasPermission(structures.UserEditorPermissionViewMessages) {
				return qr.setError(errors.ErrInsufficientPrivilege().SetDetail("You are not allowed to view the messages of this user"))
			}
		}
	}

	// Fetch message read states where target user is recipient
	cur, err := q.mongo.Collection(mongo.CollectionNameMessagesRead).Find(ctx, bson.M{
		"recipient_id": user.ID,
		"kind":         structures.MessageKindInbox,
	}, options.Find().SetProjection(bson.M{"message_id": 1}))
	if err != nil {
		logrus.WithError(err).WithField("user_id", user.ID.Hex()).Error("failed to find read states of inbox messages")
		return qr.setError(errors.ErrInternalServerError().SetDetail(err.Error()))
	}
	messageIDs := []primitive.ObjectID{}
	for cur.Next(ctx) {
		msg := &structures.MessageRead{}
		if err = cur.Decode(msg); err != nil {
			continue
		}
		messageIDs = append(messageIDs, msg.MessageID)
	}

	and := bson.A{bson.M{"_id": bson.M{"$in": messageIDs}}}
	if !opt.AfterID.IsZero() {
		and = append(and, bson.M{"_id": bson.M{"$gt": opt.AfterID}})
	}

	return q.Messages(ctx, bson.M{"$and": and}, MessageQueryOptions{
		Actor:            actor,
		Limit:            opt.Limit,
		ReturnUnread:     true,
		FilterRecipients: []primitive.ObjectID{user.ID},
	})
}

func (q *Query) ModRequestMessages(ctx context.Context, opt ModRequestMessagesQueryOptions) *QueryResult[structures.Message] {
	qr := &QueryResult[structures.Message]{}
	actor := opt.Actor
	targets := opt.Targets

	if !opt.SkipPermissionCheck {
		if actor == nil {
			return qr.setError(errors.ErrUnauthorized())
		}

		// check permissions for targets
		if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
			targets[structures.ObjectKindEmote] = false
		}
		if !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
			targets[structures.ObjectKindEmoteSet] = false
		}
		if !actor.HasPermission(structures.RolePermissionManageReports) {
			targets[structures.ObjectKindReport] = false
		}
	}
	targetsAry := []structures.ObjectKind{}
	for k, ok := range targets {
		if ok {
			targetsAry = append(targetsAry, k)
		}
	}

	f := bson.M{
		"kind": structures.MessageKindModRequest,
		"data.target_kind": bson.M{
			"$in": targetsAry,
		},
	}
	if len(opt.TargetIDs) > 0 {
		f["data.target_id"] = bson.M{"$in": opt.TargetIDs}
	}

	return q.Messages(ctx, f, MessageQueryOptions{
		Actor: actor,
		Limit: 100,
	})
}

func (q *Query) Messages(ctx context.Context, filter bson.M, opt MessageQueryOptions) *QueryResult[structures.Message] {
	qr := &QueryResult[structures.Message]{}
	items := []*structures.Message{}

	// Set limit?
	limit := mongo.Pipeline{}
	if opt.Limit != 0 {
		limit = append(limit, bson.D{{Key: "$limit", Value: opt.Limit}})
	}

	// Create the pipeline
	cur, err := q.mongo.Collection(mongo.CollectionNameMessages).Aggregate(ctx, aggregations.Combine(
		// Search message read states
		mongo.Pipeline{
			{{Key: "$sort", Value: bson.M{"_id": -1}}},
			{{Key: "$match", Value: filter}},
		},
		limit,
		mongo.Pipeline{
			{{
				Key: "$lookup",
				Value: mongo.Lookup{
					From:         mongo.CollectionNameMessagesRead,
					LocalField:   "_id",
					ForeignField: "message_id",
					As:           "read_states",
				},
			}},
			{{
				Key: "$set",
				Value: bson.M{
					"readers": bson.M{"$size": "$read_states"},
					"read": bson.M{"$getField": bson.M{
						"input": bson.M{"$first": bson.M{
							"$filter": bson.M{
								"input": "$read_states",
								"as":    "rs",
								"cond": bson.M{
									"$and": func() bson.A {
										a := bson.A{bson.M{"$eq": bson.A{"$$rs.read", true}}}
										if len(opt.FilterRecipients) > 0 {
											a = append(a, bson.M{"$in": bson.A{"$$rs.recipient_id", opt.FilterRecipients}})
										}

										return a
									}(),
								},
							},
						}},
						"field": "read",
					}},
				},
			}},
			{{
				Key: "$match",
				Value: func() bson.M {
					m := bson.M{"readers": bson.M{"$gt": 0}}
					if !opt.ReturnUnread {
						m["read"] = bson.M{"$not": bson.M{"$eq": true}}
					}

					return m
				}(),
			}},
			{{
				Key:   "$unset",
				Value: bson.A{"read_states"},
			}},
			{{
				Key: "$group",
				Value: bson.M{
					"_id": nil,
					"messages": bson.M{
						"$push": "$$ROOT",
					},
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
		return qr.setError(errors.ErrInternalServerError().SetDetail(err.Error()))
	}

	v := &aggregatedMessagesResult{}
	cur.Next(ctx)
	if err := cur.Decode(v); err != nil {
		if err == io.EOF {
			return qr.setError(errors.ErrNoItems().SetDetail("No messages"))
		}
		logrus.WithError(err).Error("mongo, failed to decode aggregated result of mod requests query")
		return qr.setError(errors.ErrInternalServerError().SetDetail(err.Error()))
	}

	qb := &QueryBinder{ctx, q}
	userMap := qb.mapUsers(v.Authors, v.RoleEntitlements...)

	for _, msg := range v.Messages {
		msg.Author = userMap[msg.AuthorID]
		items = append(items, msg)
	}

	return qr.setItems(items)
}

type InboxMessagesQueryOptions struct {
	Actor               *structures.User
	User                *structures.User // The user to fetch inbox messagesq from
	Limit               int
	AfterID             primitive.ObjectID
	SkipPermissionCheck bool
}

type ModRequestMessagesQueryOptions struct {
	Actor               *structures.User
	Targets             map[structures.ObjectKind]bool
	TargetIDs           []primitive.ObjectID
	Filter              bson.M
	SkipPermissionCheck bool
}

type MessageQueryOptions struct {
	Actor            *structures.User
	Limit            int
	ReturnUnread     bool
	FilterRecipients []primitive.ObjectID
}

type aggregatedMessagesResult struct {
	Messages         []*structures.Message     `bson:"messages"`
	Authors          []*structures.User        `bson:"authors"`
	RoleEntitlements []*structures.Entitlement `bson:"role_entitlements"`
}
