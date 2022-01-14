package mutations

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Edit: edit the emote. Modify the EmoteBuilder beforehand!
//
// To account for editor permissions, the "editor_of" relation should be included in the actor's data
func (em *EmoteMutation) Edit(ctx context.Context, inst mongo.Instance, opt EmoteEditOptions) (*EmoteMutation, error) {
	if em.EmoteBuilder == nil || em.EmoteBuilder.Emote == nil {
		return nil, structures.ErrIncompleteMutation
	}
	actor := opt.Actor
	emote := em.EmoteBuilder.Emote

	// Check actor's permission
	if actor != nil {
		// User is not privileged
		if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
			if emote.OwnerID.IsZero() { // Deny when emote has no owner
				return nil, structures.ErrInsufficientPrivilege
			}

			// Check if actor is editor of the emote owner
			isPermittedEditor := false
			for _, ed := range actor.EditorOf {
				if ed.ID != emote.OwnerID {
					continue
				}
				// Allow if the actor has the "manage owned emotes" permission
				// as the editor of the emote owner
				if ed.HasPermission(structures.UserEditorPermissionManageOwnedEmotes) {
					isPermittedEditor = true
					break
				}
			}
			if emote.OwnerID != actor.ID && !isPermittedEditor { // Deny when not the owner or editor of the owner of the emote
				return nil, structures.ErrInsufficientPrivilege
			}
		}
	}

	// Update the emote
	if err := inst.Collection(structures.CollectionNameEmotes).FindOneAndUpdate(
		ctx,
		bson.M{"_id": emote.ID},
		em.EmoteBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(emote); err != nil {
		return nil, err
	}

	return em, nil
}

type EmoteEditOptions struct {
	Actor *structures.User
}
