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
func (esm *EmoteSetMutation) SetEmote(ctx context.Context, inst mongo.Instance, opt EmoteSetMutationSetEmoteOptions) (*EmoteSetMutation, error) {
	esm.l.Lock()
	defer esm.l.Unlock()
	if esm.EmoteSetBuilder == nil || esm.EmoteSetBuilder.EmoteSet == nil {
		return nil, errors.ErrInternalIncompleteMutation()
	}
	if len(opt.Emotes) == 0 {
		return nil, errors.ErrMissingRequiredField().SetDetail("EmoteIDs")
	}

	// Can actor do this?
	actor := opt.Actor
	if actor == nil || !actor.HasPermission(structures.RolePermissionEditEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege().SetFields(errors.Fields{"MISSING_PERMISSION": "EDIT_EMOTE_SET"})
	}

	// Get relevant data
	targetEmoteIDs := []primitive.ObjectID{}
	targetEmoteMap := map[primitive.ObjectID]EmoteSetMutationSetEmoteItem{}
	set := esm.EmoteSetBuilder.EmoteSet
	{
		// Find emote set owner
		if set.Owner == nil {
			set.Owner = &structures.User{}
			cur, err := inst.Collection(mongo.CollectionNameUsers).Aggregate(ctx, append(mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"_id": set.OwnerID}}},
			}, aggregations.UserRelationEditors...))
			cur.Next(ctx)
			if err = multierror.Append(err, cur.Decode(set.Owner), cur.Close(ctx)).ErrorOrNil(); err != nil {
				if err == mongo.ErrNoDocuments {
					return nil, errors.ErrUnknownUser().SetDetail("emote set owner")
				}
				logrus.WithError(err).Error("failed to find emote set owner")
				return nil, err
			}
		}

		// Fetch set emotes
		if len(set.Emotes) == 0 {
			cur, err := inst.Collection(mongo.CollectionNameEmoteSets).Aggregate(ctx, append(mongo.Pipeline{
				// Match only the target set
				{{Key: "$match", Value: bson.M{"_id": set.ID}}},
			}, aggregations.EmoteSetRelationActiveEmotes...))
			if err = multierror.Append(err, cur.All(ctx, &set.Emotes)).ErrorOrNil(); err != nil {
				logrus.WithError(err).Error("failed to fetch emote data of active emote set emotes")
				return nil, err
			}
		}

		// Fetch target emotes
		for _, e := range opt.Emotes {
			targetEmoteIDs = append(targetEmoteIDs, e.ID)
			targetEmoteMap[e.ID] = e
		}
		targetEmotes := []*structures.Emote{}
		cur, err := inst.Collection(mongo.CollectionNameEmotes).Aggregate(ctx, append(mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"versions.id": bson.M{"$in": targetEmoteIDs}}}},
		}, aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{Roles: true, Editors: true})...))
		err = multierror.Append(err, cur.All(ctx, &targetEmotes)).ErrorOrNil()
		if err != nil {
			logrus.WithError(err).Error("failed to fetch target emotes during emote set mutation")
			return nil, err
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
			return nil, errors.ErrInsufficientPrivilege().SetDetail("emote set is privileged")
		}
		if set.Owner != nil {
			for _, ed := range set.Owner.Editors {
				if ed.ID != actor.ID {
					continue
				}
				if !ed.HasPermission(structures.UserEditorPermissionModifyEmotes) {
					return nil, errors.ErrInsufficientPrivilege().SetFields(errors.Fields{
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

	// Iterate through the target emotes
	// Check for permissions
	for _, tgt := range targetEmoteMap {
		if tgt.emote == nil {
			continue
		}
		tgt.Name = utils.Ternary(tgt.Name != "", tgt.Name, tgt.emote.Name).(string)
		tgt.emote.Name = tgt.Name
		if err := tgt.emote.Validator().Name(); err != nil {
			return nil, err
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
					return nil, errors.ErrInsufficientPrivilege().SetFields(errors.Fields{
						"EMOTE_ID": tgt.ID.Hex(),
					}).SetDetail("emote is private")
				}
			}

			// Verify that the set has available slots
			if !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
				if len(set.Emotes) >= int(set.EmoteSlots) {
					return nil, errors.ErrNoSpaceAvailable().
						SetDetail("This set does not have enough slots").
						SetFields(errors.Fields{"SLOTS": set.EmoteSlots})
				}
			}

			// Check for conflicts with existing emotes
			for _, e := range set.Emotes {
				// Cannot enable the same emote twice
				if tgt.ID == e.ID {
					return nil, errors.ErrEmoteAlreadyEnabled()
				}
				// Cannot have the same emote name as another active emote
				if tgt.Name == e.Name {
					return nil, errors.ErrEmoteNameConflict()
				}
			}

			// Add active emote
			esm.EmoteSetBuilder.AddActiveEmote(tgt.ID, tgt.Name, time.Now())
		case ListItemActionUpdate, ListItemActionRemove:
			// The emote must already be active
			found := false
			for _, e := range set.Emotes {
				if tgt.Action == ListItemActionUpdate && e.Name == tgt.Name {
					return nil, errors.ErrEmoteNameConflict().SetFields(errors.Fields{
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
				return nil, errors.ErrEmoteNotEnabled().SetFields(errors.Fields{
					"EMOTE_ID": tgt.ID.Hex(),
				})
			}

			if tgt.Action == ListItemActionUpdate {
				esm.EmoteSetBuilder.UpdateActiveEmote(tgt.ID, tgt.Name)
			} else if tgt.Action == ListItemActionRemove {
				esm.EmoteSetBuilder.RemoveActiveEmote(tgt.ID)
			}
		}
	}

	// Update the document
	if len(esm.EmoteSetBuilder.Update) == 0 {
		return nil, errors.ErrUnknownEmote().SetDetail("no target emotes found")
	}
	if err := inst.Collection(mongo.CollectionNameEmoteSets).FindOneAndUpdate(
		ctx,
		bson.M{"_id": set.ID},
		esm.EmoteSetBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(esm.EmoteSetBuilder.EmoteSet); err != nil {
		logrus.WithError(err).WithField("emote_set_id", set.ID).Error("mongo, failed to update emote set")
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	esm.EmoteSetBuilder.Update.Clear()
	return esm, nil
}
