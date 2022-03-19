package mutations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (m *Mutate) CreateBan(ctx context.Context, bb *structures.BanBuilder, opt CreateBanOptions) error {
	if bb == nil || bb.Ban == nil {
		return structures.ErrIncompleteMutation
	} else if bb.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}
	if opt.Victim == nil {
		return errors.ErrMissingRequiredField().SetDetail("Did not specify a victim")
	}

	// Check permissions
	// can the actor ban the victim?
	actor := opt.Actor
	victim := opt.Victim
	if actor != nil {
		if victim.ID == actor.ID {
			return errors.ErrDontBeSilly()
		}
		if victim.GetHighestRole().Position >= actor.GetHighestRole().Position {
			return errors.ErrInsufficientPrivilege().
				SetDetail("Victim has an equal or higher privilege level").
				SetFields(errors.Fields{
					"ACTOR_ROLE_POSITION":  actor.GetHighestRole().Position,
					"VICTIM_ROLE_POSITION": victim.GetHighestRole().Position,
				})
		}
	}

	// Write
	result, err := m.mongo.Collection(mongo.CollectionNameBans).InsertOne(ctx, bb.Ban)
	if err != nil {
		return errors.ErrInternalServerError().SetDetail(err.Error())
	}
	bb.Ban.ID = result.InsertedID.(primitive.ObjectID)

	// Get the newly created ban
	_ = m.mongo.Collection(mongo.CollectionNameBans).FindOne(ctx, bson.M{"_id": bb.Ban.ID}).Decode(bb.Ban)

	// Send a message to the victim
	mb := structures.NewMessageBuilder(nil).
		SetKind(structures.MessageKindInbox).
		SetAuthorID(actor.ID).
		SetTimestamp(time.Now()).
		SetAnonymous(opt.AnonymousActor).
		AsInbox(structures.MessageDataInbox{
			Subject:   "inbox.generic.client_banned.subject",
			Content:   "inbox.generic.client_banned.content",
			Important: true,
			Placeholders: func() map[string]string {
				m := map[string]string{
					"BAN_REASON":    bb.Ban.Reason,
					"BAN_EXPIRE_AT": utils.Ternary(bb.Ban.ExpireAt.IsZero(), "never", bb.Ban.ExpireAt.Format(time.RFC822)).(string),
				}
				for k, e := range structures.BanEffectMap {
					if bb.Ban.Effects.Has(e) {
						m[fmt.Sprintf("EFFECT_%s", k)] = fmt.Sprintf(
							"inbox.generic.client_banned.effect.%s", strings.ToLower(k),
						)
					}
				}
				return m
			}(),
		})
	if err := m.SendInboxMessage(ctx, mb, SendInboxMessageOptions{
		Actor:                actor,
		Recipients:           []primitive.ObjectID{victim.ID},
		ConsiderBlockedUsers: false,
	}); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"actor_id":  actor.ID.Hex(),
			"victim_id": victim.ID.Hex(),
			"ban_id":    bb.Ban.ID.Hex(),
		}).Error("failed to send inbox message to victim about created ban")
	}

	bb.MarkAsTainted()
	return nil
}

type CreateBanOptions struct {
	Actor          *structures.User
	AnonymousActor bool
	Victim         *structures.User
}
