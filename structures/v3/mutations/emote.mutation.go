package mutations

import (
	"context"
	"strconv"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Edit: edit the emote. Modify the EmoteBuilder beforehand!
//
// To account for editor permissions, the "editor_of" relation should be included in the actor's data
func (m *Mutate) EditEmote(ctx context.Context, eb *structures.EmoteBuilder, opt EmoteEditOptions) error {
	if eb == nil || eb.Emote == nil {
		return structures.ErrIncompleteMutation
	} else if eb.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}
	actor := opt.Actor
	emote := eb.Emote

	// Check actor's permission
	if actor != nil {
		// User is not privileged
		if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
			if emote.OwnerID.IsZero() { // Deny when emote has no owner
				return structures.ErrInsufficientPrivilege
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
				return structures.ErrInsufficientPrivilege
			}
		}
	}

	if !opt.SkipValidation {
		init := eb.Initial()
		validator := eb.Emote.Validator()
		// Change: Name
		if init.Name != emote.Name {
			if err := validator.Name(); err != nil {
				return err
			}
		}
		if init.OwnerID != emote.OwnerID {
			// Verify that the new emote exists
			if err := m.mongo.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
				"_id": emote.OwnerID,
			}).Err(); err != nil {
				if err == mongo.ErrNoDocuments {
					return errors.ErrUnknownUser()
				}
			}
		}
		if init.Flags != emote.Flags {
			f := emote.Flags

			// Validate privileged flags
			if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
				privilegedBits := []structures.EmoteFlag{
					structures.EmoteFlagsContentSexual,
					structures.EmoteFlagsContentEpilepsy,
					structures.EmoteFlagsContentEdgy,
					structures.EmoteFlagsContentTwitchDisallowed,
				}
				for _, flag := range privilegedBits {
					if f&flag != init.Flags&flag {
						return errors.ErrInsufficientPrivilege().SetDetail("Not allowed to modify flag %s", flag.String())
					}
				}
			}
		}
		// Change versions
		for i, ver := range emote.Versions {
			oldVer := eb.InitialVersions()[i]
			if oldVer == nil {
				continue // cannot update version that didn't exist
			}
			// Update: listed
			if ver.State.Listed != oldVer.State.Listed {
				if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
					return errors.ErrInsufficientPrivilege().SetDetail("Not allowed to modify listed state of version %s", strconv.Itoa(i))
				}
			}
			if ver.Name != "" && ver.Name != oldVer.Name {
				if err := ver.Validator().Name(); err != nil {
					return err
				}
			}
			if ver.Description != "" && ver.Description != oldVer.Description {
				if err := ver.Validator().Description(); err != nil {
					return err
				}
			}
		}
	}

	// Update the emote
	if err := m.mongo.Collection(mongo.CollectionNameEmotes).FindOneAndUpdate(
		ctx,
		bson.M{"versions.id": emote.ID},
		eb.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(emote); err != nil {
		logrus.WithError(err).Error("mongo, couldn't edit emote")
		return errors.ErrInternalServerError().SetDetail(err.Error())
	}

	eb.MarkAsTainted()
	return nil
}

type EmoteEditOptions struct {
	Actor          *structures.User
	SkipValidation bool
}
