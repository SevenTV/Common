package mutations

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserMutation struct {
	UserBuilder *structures.UserBuilder
}

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
				ok = utils.BitField.HasBits(int64(ed.Permissions), int64(structures.UserEditorPermissionAddChannelEmotes))
			case ListItemActionRemove: // Check permissions for REMOVE
				ok = utils.BitField.HasBits(int64(ed.Permissions), int64(structures.UserEditorPermissionRemoveChannelEmotes))
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
	case ListItemActionAdd:
		um.UserBuilder.Update.AddToSet("channel_emotes", &structures.UserEmote{
			ID:    opt.EmoteID,
			Alias: opt.Alias,
		})
	case ListItemActionRemove:
		um.UserBuilder.Update.Pull("channel_emotes", bson.M{
			"id": opt.EmoteID,
		})
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
