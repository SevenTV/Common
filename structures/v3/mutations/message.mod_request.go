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

func (m *Mutate) SendModRequestMessage(ctx context.Context, mb *structures.MessageBuilder) error {
	if mb == nil || mb.Message == nil {
		return errors.ErrInternalIncompleteMutation()
	} else if mb.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}

	// Get the message
	req := mb.DecodeModRequest()

	// Verify that the target item exists
	var target interface{}
	filter := bson.M{"_id": req.TargetID}
	switch req.TargetKind {
	case structures.ObjectKindEmote:
		filter = bson.M{"versions.id": req.TargetID}
	}
	coll := mongo.CollectionName(req.TargetKind.CollectionName())
	if err := m.mongo.Collection(coll).FindOne(ctx, filter).Decode(&target); err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrInvalidRequest().SetDetail("Target item doesn't exist")
		}
		return errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Create the message
	result, err := m.mongo.Collection(mongo.CollectionNameMessages).InsertOne(ctx, mb.Message)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"modrequest_target_kind": req.TargetKind,
			"modrequest_target_id":   req.TargetID,
		}).Error("mongo, failed to create mod request message")
		return err
	}
	msgID := result.InsertedID.(primitive.ObjectID)
	mb.Message.ID = msgID

	// Create a read state
	m.mongo.Collection(mongo.CollectionNameMessagesRead).InsertOne(ctx, &structures.MessageRead{
		MessageID: msgID,
		Kind:      structures.MessageKindModRequest,
		Timestamp: time.Now(),
	})

	mb.MarkAsTainted()
	return nil
}
