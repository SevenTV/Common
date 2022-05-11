package mutations

import (
	"context"
	"strconv"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const EMOTE_CLAIMANTS_MOST = 10

// Edit: edit the emote. Modify the EmoteBuilder beforehand!
//
// To account for editor permissions, the "editor_of" relation should be included in the actor's data
func (m *Mutate) EditEmote(ctx context.Context, eb *structures.EmoteBuilder, opt EmoteEditOptions) error {
	if eb == nil {
		return structures.ErrIncompleteMutation
	} else if eb.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}
	actor := opt.Actor
	actorID := primitive.NilObjectID
	if actor != nil {
		actorID = actor.ID
	}
	emote := &eb.Emote

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
	} else if !opt.SkipValidation {
		// if validation is not skipped then an Actor is mandatory
		return errors.ErrUnauthorized()
	}

	// Set up audit logs
	log := structures.NewAuditLogBuilder(structures.AuditLog{}).
		SetKind(structures.AuditLogKindUpdateEmote).
		SetActor(actorID).
		SetTargetKind(structures.ObjectKindEmote).
		SetTargetID(emote.ID)

	if !opt.SkipValidation {
		init := eb.Initial()
		validator := eb.Emote.Validator()
		// Change: Name
		if init.Name != emote.Name {
			if err := validator.Name(); err != nil {
				return err
			}
			c := structures.AuditLogChange{
				Key:    "name",
				Format: structures.AuditLogChangeFormatSingleValue,
			}
			c.WriteSingleValues(init.Name, emote.Name)
			log.AddChanges(&c)
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

			// If the user is not privileged:
			// we will add the specified owner_id to list of claimants and send an inbox message
			switch init.OwnerID == actorID { // original owner is actor?
			case true: // yes: means emote owner is transferring away
				if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
					if utils.Contains(emote.State.Claimants, emote.OwnerID) { // error if target new owner is already a claimant
						return errors.ErrInsufficientPrivilege().SetDetail("Target user was already requested to claim ownership of this emote")
					}
					if len(emote.State.Claimants) > EMOTE_CLAIMANTS_MOST {
						return errors.ErrInvalidRequest().SetDetail("Too Many Claimants (%d)", EMOTE_CLAIMANTS_MOST)
					}

					// Add to claimants
					eb.Update.AddToSet("state.claimants", emote.OwnerID)

					// Send a message to the claimant's inbox
					mb := structures.NewMessageBuilder(structures.Message[structures.MessageDataInbox]{}).
						SetKind(structures.MessageKindInbox).
						SetAuthorID(actorID).
						SetTimestamp(time.Now()).
						SetData(structures.MessageDataInbox{
							Subject: "inbox.generic.emote_ownership_claim_request.subject",
							Content: "inbox.generic.emote_ownership_claim_request.content",
							Locale:  true,
							Placeholders: map[string]string{
								"OWNER_DISPLAY_NAME":  utils.Ternary(emote.Owner.DisplayName != "", emote.Owner.DisplayName, emote.Owner.Username),
								"EMOTE_VERSION_COUNT": strconv.Itoa(len(emote.Versions)),
								"EMOTE_NAME":          emote.Name,
							},
						})
					if err := m.SendInboxMessage(ctx, mb, SendInboxMessageOptions{
						Actor:                actor,
						Recipients:           []primitive.ObjectID{emote.OwnerID},
						ConsiderBlockedUsers: true,
					}); err != nil {
						return err
					}
					// Undo owner update
					eb.Update.UndoSet("owner_id")
					emote.OwnerID = init.OwnerID
				}
			case false: // no: a user wants to claim ownership
				// Check if actor is allowed to do that
				if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
					if emote.OwnerID != actorID { //
						return errors.ErrInsufficientPrivilege().SetDetail("You are not permitted to change this emote's owner")
					}
					if !utils.Contains(emote.State.Claimants, emote.OwnerID) {
						return errors.ErrInsufficientPrivilege().SetDetail("You are not allowed to claim ownership of this emote")
					}
				}
				// At this point the actor has successfully claimed ownership of the emote and we clear the list of claimants
				eb.Update.Set("state.claimants", []primitive.ObjectID{})
			}

			// Write as audit change
			c := structures.AuditLogChange{
				Key:    "owner_id",
				Format: structures.AuditLogChangeFormatSingleValue,
			}
			c.WriteSingleValues(init.OwnerID, emote.OwnerID)
			log.AddChanges(&c)
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
			c := structures.AuditLogChange{
				Key:    "flags",
				Format: structures.AuditLogChangeFormatSingleValue,
			}
			c.WriteSingleValues(init.Flags, emote.Flags)
			log.AddChanges(&c)
		}
		// Change versions
		for i, ver := range emote.Versions {
			oldVer := eb.InitialVersions()[i]
			if oldVer == nil {
				continue // cannot update version that didn't exist
			}
			c := structures.AuditLogChange{
				Key:    "versions",
				Format: structures.AuditLogChangeFormatArrayChange,
			}

			// Update: listed
			changeCount := 0
			if ver.State.Listed != oldVer.State.Listed {
				if !actor.HasPermission(structures.RolePermissionEditAnyEmote) {
					return errors.ErrInsufficientPrivilege().SetDetail("Not allowed to modify listed state of version %s", strconv.Itoa(i))
				}
				changeCount++
			}
			if ver.Name != "" && ver.Name != oldVer.Name {
				if err := ver.Validator().Name(); err != nil {
					return err
				}
				changeCount++
			}
			if ver.Description != "" && ver.Description != oldVer.Description {
				if err := ver.Validator().Description(); err != nil {
					return err
				}
				changeCount++
			}
			if changeCount > 0 {
				c.WriteArrayUpdated(structures.AuditLogChangeSingleValue{
					New:      ver,
					Old:      oldVer,
					Position: int32(i),
				})
			}
		}
	}

	// Update the emote
	if len(eb.Update) > 0 {
		if err := m.mongo.Collection(mongo.CollectionNameEmotes).FindOneAndUpdate(
			ctx,
			bson.M{"versions.id": emote.ID},
			eb.Update,
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(emote); err != nil {
			zap.S().Errorw("mongo, couldn't edit emote",
				"error", err,
			)
			return errors.ErrInternalServerError().SetDetail(err.Error())
		}

		// Write audit log entry
		go func() {
			if _, err := m.mongo.Collection(mongo.CollectionNameAuditLogs).InsertOne(ctx, log.AuditLog); err != nil {
				zap.S().Errorw("failed to write audit log",
					"error", err,
				)
			}
		}()
	}

	eb.MarkAsTainted()
	return nil
}

type EmoteEditOptions struct {
	Actor          *structures.User
	SkipValidation bool
}
