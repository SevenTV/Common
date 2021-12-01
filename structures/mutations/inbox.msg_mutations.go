package mutations

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (mm *MessageMutation) SendInboxMessage(ctx context.Context, inst mongo.Instance, opt SendInboxMessageOptions) (*MessageMutation, error) {
	if mm.MessageBuilder == nil || mm.MessageBuilder.Message == nil {
		return nil, structures.ErrIncompleteMutation
	}

	// Check actor permissions
	actor := opt.Actor
	if actor == nil || actor.ID.IsZero() || !actor.HasPermission(structures.RolePermissionSendMessages) {
		return nil, structures.ErrInsufficientPrivilege
	}

	// Find recipients
	recipients := []*structures.User{}
	cur, err := inst.Collection(mongo.CollectionNameUsers).Find(ctx, bson.M{
		"$and": func() bson.A {
			a := bson.A{bson.M{"_id": bson.M{"$in": opt.Recipients}}}
			if opt.ConsiderBlockedUsers { // omit blocked users from recipients?
				a = append(a, bson.M{"blocked_user_ids": bson.M{"$not": bson.M{"$eq": actor.ID}}})
			}

			return a
		}(),
	})
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}
	if err = cur.All(ctx, &recipients); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Write message to DB
	result, err := inst.Collection(mongo.CollectionNameMessages).InsertOne(ctx, mm.MessageBuilder.Message)
	if err != nil {
		logrus.WithError(err).WithField("actor_id", actor.ID).Error("mongo, failed to create message")
		return nil, err
	}
	msgID := result.InsertedID.(primitive.ObjectID)

	// Create read states for the recipients
	w := make([]mongo.WriteModel, len(recipients))
	for i, u := range recipients {
		w[i] = &mongo.InsertOneModel{
			Document: &structures.MessageRead{
				MessageID:   msgID,
				RecipientID: u.ID,
				Read:        false,
			},
		}
	}
	if _, err = inst.Collection(mongo.CollectionNameMessagesRead).BulkWrite(ctx, w); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"message_id":      result.InsertedID,
			"recipient_count": len(recipients),
			"recipient_ids":   opt.Recipients,
		}).Error("mongo, couldn't create a read state for message")
	}

	mm.MessageBuilder.Message.ID = msgID
	return mm, nil
}

type SendInboxMessageOptions struct {
	Actor                *structures.User
	Recipients           []primitive.ObjectID
	ConsiderBlockedUsers bool
}
