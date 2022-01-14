package mutations

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetChannelEmote: add, update or remove a channel emote for the user
func (um *UserMutation) SetChannelEmote(ctx context.Context, inst mongo.Instance, opt SetChannelEmoteOptions) (*UserMutation, error) {
	if um.UserBuilder == nil || um.UserBuilder.User == nil {
		return nil, structures.ErrIncompleteMutation
	}
	targetUser := um.UserBuilder.User
	actor := opt.Actor

	// Check for the permission
	if !targetUser.HasPermission(structures.RolePermissionSetChannelEmote) {
		return nil, structures.ErrInsufficientPrivilege
	}

	// Fetch the target user if they are not the actor
	if actor.ID != targetUser.ID {
		// Ensure that the actor has permission to edit the target
		ok := false
		for _, ed := range targetUser.Editors {
			if ed.ID != actor.ID { // Skip if not the actor
				continue
			}

			switch opt.Action {
			case ListItemActionAdd: // Check permissions for ADD
				ok = utils.BitField.HasBits(int64(ed.Permissions), int64(structures.UserEditorPermissionModifyChannelEmotes))
			case ListItemActionRemove: // Check permissions for REMOVE
				ok = utils.BitField.HasBits(int64(ed.Permissions), int64(structures.UserEditorPermissionModifyChannelEmotes))
			default:
				ok = true
			}
			break
		}
		if !ok {
			return nil, structures.ErrInsufficientPrivilege
		}
	}

	// Assign the emote
	switch opt.Action {
	case ListItemActionAdd: // Add Emote
		um.UserBuilder.Update.AddToSet("channel_emotes", &structures.UserEmote{
			ID: opt.EmoteID,
		})
	case ListItemActionUpdate: // Update Emote
		ind := -1
		emotes := um.UserBuilder.User.ChannelEmotes
		for i, em := range emotes {
			if em.ID != opt.EmoteID {
				continue
			}

			ind = i
			break
		}
		if ind == -1 {
			return nil, fmt.Errorf("emote not enabled")
		}

		um.UserBuilder.Update.Set(fmt.Sprintf("channel_emotes.%d", ind), &structures.UserEmote{
			ID: opt.EmoteID,
		})
	case ListItemActionRemove: // Remove Emote
		um.UserBuilder.Update.Pull("channel_emotes", bson.M{
			"id": opt.EmoteID,
		})
	}
	// Update the document
	if err := inst.Collection(structures.CollectionNameUsers).FindOneAndUpdate(
		ctx,
		bson.M{"_id": targetUser.ID},
		um.UserBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(targetUser); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, structures.ErrInternalError
	}

	return um, nil
}

type SetChannelEmoteOptions struct {
	Actor    *structures.User
	EmoteID  primitive.ObjectID
	Channels []primitive.ObjectID
	Alias    string
	Action   ListItemAction
}
