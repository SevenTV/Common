package mutations

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Mutate) SetMessageReadStates(ctx context.Context, mb *structures.MessageBuilder, read bool, opt MessageReadStateOptions) (*MessageReadStateResponse, error) {
	if mb == nil || mb.Message == nil {
		return nil, errors.ErrInternalIncompleteMutation()
	} else if mb.IsTainted() {
		return nil, errors.ErrMutateTaintedObject()
	}

	// Check permissions
	actor := opt.Actor
	if !opt.SkipPermissionCheck && actor == nil {
		return nil, errors.ErrUnauthorized()
	}

	// Find the readstates
	filter := bson.M{}
	if len(opt.Filter) > 0 {
		filter = opt.Filter
	}
	filter["message_id"] = mb.Message.ID
	cur, err := m.mongo.Collection(mongo.CollectionNameMessagesRead).Find(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrUnknownMessage().SetDetail("Couldn't find any read states related to the message")
		}
		logrus.WithError(err).Error("mongo, ")
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	errorList := []error{}
	w := []mongo.WriteModel{}
	for cur.Next(ctx) {
		rs := &structures.MessageRead{}
		if err := cur.Decode(rs); err != nil {
			continue
		}
		// Can actor do this?
		if !opt.SkipPermissionCheck && rs.RecipientID != actor.ID {
			switch rs.Kind {
			// Check for a mod request
			case structures.MessageKindModRequest:
				d := mb.DecodeModRequest()
				errf := errors.Fields{
					"message_state_id": rs.ID,
					"msg_kind":         rs.Kind,
					"target_kind":      d.TargetKind,
				}

				if d.TargetKind == structures.ObjectKindEmote && !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
					errorList = append(errorList, errors.ErrInsufficientPrivilege().SetFields(errf))
					continue // target is emote but actor lacks "edit any emote" permission
				} else if d.TargetKind == structures.ObjectKindEmoteSet && !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
					errorList = append(errorList, errors.ErrInsufficientPrivilege().SetFields(errf))
					continue // target is emote set but actor lacks "edit any emote set" permission
				} else if d.TargetKind == structures.ObjectKindReport && !actor.HasPermission(structures.RolePermissionManageReports) {
					errorList = append(errorList, errors.ErrInsufficientPrivilege().SetFields(errf))
					continue // target is report but actor lacks "manage reports" permission
				}
			default:
				continue
			}
		}

		// Add as item to be written
		w = append(w, &mongo.UpdateOneModel{
			Filter: bson.M{"_id": rs.ID},
			Update: bson.M{"$set": bson.M{
				"read": read,
			}},
		})
	}

	updated := int64(0)
	if len(w) > 0 {
		result, err := m.mongo.Collection(mongo.CollectionNameMessagesRead).BulkWrite(ctx, w)
		if err != nil {
			logrus.WithError(err).WithField(
				"message_id", mb.Message.ID,
			).Error("mongo, failed to update message read states")
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}
		updated += result.ModifiedCount
	}

	mb.MarkAsTainted()
	return &MessageReadStateResponse{
		Updated: updated,
		Errors:  errorList,
	}, nil
}

type MessageReadStateOptions struct {
	Actor               *structures.User
	Filter              bson.M
	SkipPermissionCheck bool
}

type MessageReadStateResponse struct {
	Updated int64   `json:"changed"`
	Errors  []error `json:"errors"`
}
