package mutations

import (
	"context"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/Common/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetEmote: enable, edit or disable active emotes in the set
func (m *Mutate) EditEmotesInSet(ctx context.Context, esb *structures.EmoteSetBuilder, opt EmoteSetMutationSetEmoteOptions) error {
	if esb == nil {
		return errors.ErrInternalIncompleteMutation()
	} else if esb.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}
	if len(opt.Emotes) == 0 {
		return errors.ErrMissingRequiredField().SetDetail("EmoteIDs")
	}

	// Can actor do this?
	actor := opt.Actor
	if actor == nil || !actor.HasPermission(structures.RolePermissionEditEmoteSet) {
		return errors.ErrInsufficientPrivilege().SetFields(errors.Fields{"MISSING_PERMISSION": "EDIT_EMOTE_SET"})
	}

	// Get relevant data
	targetEmoteIDs := []primitive.ObjectID{}
	targetEmoteMap := map[primitive.ObjectID]EmoteSetMutationSetEmoteItem{}
	set := esb.EmoteSet
	{
		// Find emote set owner
		if set.Owner == nil {
			set.Owner = &structures.User{}
			cur, err := m.mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, append(mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"_id": set.OwnerID}}},
			}, aggregations.UserRelationEditors...))
			cur.Next(ctx)
			if err = multierror.Append(err, cur.Decode(set.Owner), cur.Close(ctx)).ErrorOrNil(); err != nil {
				if err == mongo.ErrNoDocuments {
					return errors.ErrUnknownUser().SetDetail("emote set owner")
				}
				logrus.WithError(err).Error("failed to find emote set owner")
				return err
			}
		}

		// Fetch set emotes
		if len(set.Emotes) == 0 {
			cur, err := m.mongo.Collection(mongo.CollectionNameEmoteSets).Aggregate(ctx, append(mongo.Pipeline{
				// Match only the target set
				{{Key: "$match", Value: bson.M{"_id": set.ID}}},
			}, aggregations.EmoteSetRelationActiveEmotes...))
			if err = multierror.Append(err, cur.All(ctx, &set.Emotes)).ErrorOrNil(); err != nil {
				logrus.WithError(err).Error("failed to fetch emote data of active emote set emotes")
				return err
			}
		}

		// Fetch target emotes
		for _, e := range opt.Emotes {
			targetEmoteIDs = append(targetEmoteIDs, e.ID)
			targetEmoteMap[e.ID] = e
		}
		targetEmotes := []*structures.Emote{}
		cur, err := m.mongo.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, append(mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"versions.id": bson.M{"$in": targetEmoteIDs}}}},
		}, aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{Roles: true, Editors: true})...))
		err = multierror.Append(err, cur.All(ctx, &targetEmotes)).ErrorOrNil()
		if err != nil {
			logrus.WithError(err).Error("failed to fetch target emotes during emote set mutation")
			return err
		}
		for _, e := range targetEmotes {
			for _, ver := range e.Versions {
				if v, ok := targetEmoteMap[ver.ID]; ok {
					v.emote = e
					targetEmoteMap[e.ID] = v
				}
			}
		}
	}

	// The actor must have access to the emote set
	if set.OwnerID != actor.ID && !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
		if set.Privileged && !actor.HasPermission(structures.RolePermissionSuperAdministrator) {
			return errors.ErrInsufficientPrivilege().SetDetail("emote set is privileged")
		}
		if set.Owner != nil {
			for _, ed := range set.Owner.Editors {
				if ed.ID != actor.ID {
					continue
				}
				if !ed.HasPermission(structures.UserEditorPermissionModifyEmotes) {
					return errors.ErrInsufficientPrivilege().SetFields(errors.Fields{
						"MISSING_EDITOR_PERMISSION": "MODIFY_EMOTES",
					})
				}
				break
			}
		}
	}

	// Make a map of active set emotes
	activeEmotes := map[primitive.ObjectID]*structures.Emote{}
	for _, e := range set.Emotes {
		activeEmotes[e.ID] = e.Emote
	}

	// Set up audit log entry
	c := structures.AuditLogChange{
		Format: structures.AuditLogChangeFormatArrayChange,
		Key:    "emotes",
	}
	log := structures.NewAuditLogBuilder(structures.AuditLog{}).
		SetKind(structures.AuditLogKindUpdateEmoteSet).
		SetActor(actor.ID).
		SetTargetKind(structures.ObjectKindEmoteSet).
		SetTargetID(set.ID).
		AddChanges(c)

	// Iterate through the target emotes
	// Check for permissions
	for _, tgt := range targetEmoteMap {
		if tgt.emote == nil {
			continue
		}
		tgt.Name = utils.Ternary(tgt.Name != "", tgt.Name, tgt.emote.Name)
		tgt.emote.Name = tgt.Name
		if err := tgt.emote.Validator().Name(); err != nil {
			return err
		}

		switch tgt.Action {
		// ADD EMOTE
		case ListItemActionAdd:
			// Handle emote privacy
			if utils.BitField.HasBits(int64(tgt.emote.Flags), int64(structures.EmoteFlagsPrivate)) {
				usable := false
				// Usable if actor has Bypass Privacy permission
				if actor.HasPermission(structures.RolePermissionBypassPrivacy) {
					usable = true
				}
				// Usable if actor is an editor of emote owner
				// and has the correct permission
				if tgt.emote.Owner != nil {
					var editor *structures.UserEditor
					for _, ed := range tgt.emote.Owner.Editors {
						if opt.Actor.ID == ed.ID {
							editor = ed
							break
						}
					}
					if editor != nil && editor.HasPermission(structures.UserEditorPermissionUsePrivateEmotes) {
						usable = true
					}
				}

				if !usable {
					return errors.ErrInsufficientPrivilege().SetFields(errors.Fields{
						"EMOTE_ID": tgt.ID.Hex(),
					}).SetDetail("emote is private")
				}
			}

			// Verify that the set has available slots
			if !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
				if len(set.Emotes) >= int(set.EmoteSlots) {
					return errors.ErrNoSpaceAvailable().
						SetDetail("This set does not have enough slots").
						SetFields(errors.Fields{"SLOTS": set.EmoteSlots})
				}
			}

			// Check for conflicts with existing emotes
			for _, e := range set.Emotes {
				// Cannot enable the same emote twice
				if tgt.ID == e.ID {
					return errors.ErrEmoteAlreadyEnabled()
				}
				// Cannot have the same emote name as another active emote
				if tgt.Name == e.Name {
					return errors.ErrEmoteNameConflict()
				}
			}

			// Add active emote
			at := time.Now()
			esb.AddActiveEmote(tgt.ID, tgt.Name, at, &actor.ID)
			c.WriteArrayAdded(structures.ActiveEmote{
				ID:        tgt.ID,
				Name:      tgt.Name,
				Flags:     tgt.Flags,
				Timestamp: at,
				ActorID:   actor.ID,
			})
		case ListItemActionUpdate, ListItemActionRemove:
			// The emote must already be active
			found := false
			for _, e := range set.Emotes {
				if tgt.Action == ListItemActionUpdate && e.Name == tgt.Name {
					return errors.ErrEmoteNameConflict().SetFields(errors.Fields{
						"EMOTE_ID":          tgt.ID.Hex(),
						"CONFLICT_EMOTE_ID": tgt.ID.Hex(),
					})
				}
				if e.ID == tgt.ID {
					found = true
					break
				}
			}
			if !found {
				return errors.ErrEmoteNotEnabled().SetFields(errors.Fields{
					"EMOTE_ID": tgt.ID.Hex(),
				})
			}

			if tgt.Action == ListItemActionUpdate {
				ae, ind := esb.EmoteSet.GetEmote(tgt.ID)
				if !ae.ID.IsZero() {
					c.WriteArrayUpdated(structures.AuditLogChangeSingleValue{
						New: structures.ActiveEmote{
							ID:        tgt.ID,
							Name:      tgt.Name,
							Flags:     tgt.Flags,
							Timestamp: ae.Timestamp,
						},
						Old:      ae,
						Position: int32(ind),
					})
					esb.UpdateActiveEmote(tgt.ID, tgt.Name)
				}
			} else if tgt.Action == ListItemActionRemove {
				esb.RemoveActiveEmote(tgt.ID)
				c.WriteArrayRemoved(structures.ActiveEmote{
					ID: tgt.ID,
				})
			}
		}
	}

	// Update the document
	if len(esb.Update) == 0 {
		return errors.ErrUnknownEmote().SetDetail("no target emotes found")
	}
	if err := m.mongo.Collection(mongo.CollectionNameEmoteSets).FindOneAndUpdate(
		ctx,
		bson.M{"_id": set.ID},
		esb.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(esb.EmoteSet); err != nil {
		logrus.WithError(err).WithField("emote_set_id", set.ID).Error("mongo, failed to update emote set")
		return errors.ErrInternalServerError().SetDetail(err.Error())
	}

	// Write audit log entry
	go func() {
		if _, err := m.mongo.Collection(mongo.CollectionNameAuditLogs).InsertOne(ctx, log.AuditLog); err != nil {
			logrus.WithError(err).Error("failed to write audit log")
		}
	}()

	esb.MarkAsTainted()
	return nil
}
