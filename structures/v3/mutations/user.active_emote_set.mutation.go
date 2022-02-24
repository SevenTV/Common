package mutations

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (um *UserMutation) SetActiveEmoteSet(ctx context.Context, inst mongo.Instance, opt SetUserActiveEmoteSet) (*UserMutation, error) {
	if um.UserBuilder == nil || um.UserBuilder.User == nil {
		return um, errors.ErrInternalIncompleteMutation()
	}

	// Check for actor's permission to do this
	ub := um.UserBuilder
	actor := opt.Actor
	victim := ub.User
	if actor != nil && actor.ID != victim.ID {
		ok := actor.HasPermission(structures.RolePermissionManageUsers)
		for _, ed := range victim.Editors {
			if ed.ID != actor.ID {
				continue
			}
			// actor is editor
			// actor has permission to modify victim's emotes
			if ed.HasPermission(structures.UserEditorPermissionModifyEmotes) {
				ok = true
			}
		}
		if !ok {
			return um, errors.ErrInsufficientPrivilege().SetDetail("You are not an editor of this user")
		}
	}

	// Validate that the emote set exists and can be enabled
	if !opt.EmoteSetID.IsZero() {
		set := &structures.EmoteSet{}
		if err := inst.Collection(mongo.CollectionNameEmoteSets).FindOne(ctx, bson.M{
			"_id": opt.EmoteSetID,
		}).Decode(set); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.ErrUnknownEmoteSet()
			}
			return nil, errors.ErrInternalServerError().SetDetail(err.Error())
		}

		if !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) && set.OwnerID != actor.ID {
			return nil, errors.ErrInsufficientPrivilege().
				SetFields(errors.Fields{"owner_id": set.OwnerID.Hex()}).
				SetDetail("You do not own this emote set")
		}
	}

	// Get the connection
	conn, ok := ub.GetConnection(opt.Platform, opt.ConnectionID)
	if !ok {
		return um, errors.ErrUnknownUserConnection()
	}
	conn.SetActiveEmoteSet(opt.EmoteSetID)

	// Update document
	if err := inst.Collection(mongo.CollectionNameUsers).FindOneAndUpdate(
		ctx,
		bson.M{
			"_id":            victim.ID,
			"connections.id": opt.ConnectionID,
		},
		conn.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(victim); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrUnknownUser().SetDetail("Victim was not found and could not be updated")
		}
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	return um, nil
}

type SetUserActiveEmoteSet struct {
	EmoteSetID   primitive.ObjectID
	Platform     structures.UserConnectionPlatform
	Actor        *structures.User
	ConnectionID string
}
