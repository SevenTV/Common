package mutations

import (
	"context"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (mm *MessageMutation) SendModRequestMessage(ctx context.Context, inst mongo.Instance) (*MessageMutation, error) {
	if mm.MessageBuilder == nil || mm.MessageBuilder.Message == nil {
		return nil, errors.ErrInternalIncompleteMutation()
	}

	// Get the message
	req := mm.MessageBuilder.DecodeModRequest()

	// Verify that the target item exists
	var target interface{}
	filter := bson.M{"_id": req.TargetID}
	switch req.TargetKind {
	case structures.ObjectKindEmote:
		filter = bson.M{"versions.id": req.TargetID}
	}
	coll := mongo.CollectionName(req.TargetKind.CollectionName())
	if err := inst.Collection(coll).FindOne(ctx, filter).Decode(&target); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrInvalidRequest().SetDetail("Target item doesn't exist")
		}
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Create the message
	result, err := inst.Collection(mongo.CollectionNameMessages).InsertOne(ctx, mm.MessageBuilder.Message)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"modrequest_target_kind": req.TargetKind,
			"modrequest_target_id":   req.TargetID,
		}).Error("mongo, failed to create mod request message")
		return nil, err
	}
	msgID := result.InsertedID.(primitive.ObjectID)
	mm.MessageBuilder.Message.ID = msgID

	// Create a read state
	inst.Collection(mongo.CollectionNameMessagesRead).InsertOne(ctx, &structures.MessageRead{
		MessageID: msgID,
		Kind:      structures.MessageKindModRequest,
		Timestamp: time.Now(),
	})

	return mm, nil
}
