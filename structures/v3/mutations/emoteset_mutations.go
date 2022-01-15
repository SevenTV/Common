package mutations

import (
	"context"
	"strconv"
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

// Create: create the new emote set
func (esm *EmoteSetMutation) Create(ctx context.Context, inst mongo.Instance, opt EmoteSetMutationOptions) (*EmoteSetMutation, error) {
	esm.l.Lock()
	defer esm.l.Unlock()
	if esm.EmoteSetBuilder == nil || esm.EmoteSetBuilder.EmoteSet == nil {
		return nil, errors.ErrInternalIncompleteMutation
	}
	if esm.EmoteSetBuilder.EmoteSet.Name == "" {
		return nil, errors.ErrMissingRequiredField.SetDetail("Name")
	}

	// Check actor's permissions
	if opt.Actor == nil || !opt.Actor.HasPermission(structures.RolePermissionEditEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{"MISSING_PERMISSION": "EDIT_EMOTE_SET"})
	}

	// Create the emote set
	esm.EmoteSetBuilder.EmoteSet.ID = primitive.NewObjectID()
	result, err := inst.Collection(structures.CollectionNameEmoteSets).InsertOne(ctx, esm.EmoteSetBuilder.EmoteSet)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Get the newly created emote set
	if err = inst.Collection(structures.CollectionNameEmoteSets).FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(esm.EmoteSetBuilder.EmoteSet); err != nil {
		return nil, err
	}
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"actor_id":     opt.Actor.ID,
			"emote_set_id": result.InsertedID,
		}).Error("mongo, was unable to return the created emote set")
	}

	logrus.WithFields(logrus.Fields{
		"actor_id":     opt.Actor.ID,
		"emote_set_id": esm.EmoteSetBuilder.EmoteSet.ID,
	}).Info("Emote Set Created")
	esm.EmoteSetBuilder.Update.Clear()
	return esm, nil
}

// Edit: change the emote set
func (esm *EmoteSetMutation) Edit(ctx context.Context, inst mongo.Instance, opt EmoteSetMutationOptions) (*EmoteSetMutation, error) {
	esm.l.Lock()
	defer esm.l.Unlock()
	if esm.EmoteSetBuilder == nil || esm.EmoteSetBuilder.EmoteSet == nil {
		return nil, errors.ErrInternalIncompleteMutation
	}

	// Check actor's permissions
	actor := opt.Actor
	set := esm.EmoteSetBuilder.EmoteSet
	if actor == nil || !actor.HasPermission(structures.RolePermissionEditEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{"MISSING_PERMISSION": "EDIT_EMOTE_SET"})
	}
	if set.Privileged && !actor.HasPermission(structures.RolePermissionSuperAdministrator) {
		return nil, errors.ErrInsufficientPrivilege.SetDetail("emote set is privileged")
	}
	if actor.ID != set.OwnerID && !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege.SetDetail("you do not own this emote set")
	}

	u := esm.EmoteSetBuilder.Update
	if !opt.SkipValidation {
		// Change: Name
		if _, ok := u["name"]; ok {
			// TODO: use a regex to validate
			if len(set.Name) < structures.EmoteSetNameLengthLeast || len(set.Name) >= structures.EmoteSetNameLengthMost {
				return nil, errors.ErrValidationRejected.SetFields(errors.Fields{
					"FIELD":          "Name",
					"MIN_LENGTH":     strconv.FormatInt(int64(structures.EmoteSetNameLengthLeast), 10),
					"MAX_LENGTH":     strconv.FormatInt(int64(structures.EmoteSetNameLengthMost), 10),
					"CURRENT_LENGTH": strconv.FormatInt(int64(len(set.Name)), 10),
				})
			}
		}

		// Change: Privileged
		// Must be super admin
		if _, ok := u["privileged"]; ok && !opt.Actor.HasPermission(structures.RolePermissionSuperAdministrator) {
			return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{
				"FIELD":              "Privileged",
				"MISSING_PERMISSION": "SUPER_ADMINISTRATOR",
			})
		}

		// Change: owner
		// Must be the current owner, or have "edit any emote set" permission
		if _, ok := u["owner_id"]; ok && !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
			if actor.ID != set.OwnerID {
				return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{
					"FIELD": "OwnerID",
				}).SetDetail("you do not own this emote set")
			}
		}
	}

	// Update the document
	if err := inst.Collection(structures.CollectionNameEmoteSets).FindOneAndUpdate(
		ctx, bson.M{
			"_id": set.ID,
		},
		esm.EmoteSetBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(esm.EmoteSetBuilder.EmoteSet); err != nil {
		logrus.WithError(err).WithField("emote_set_id", set.ID).Error("mongo, failed to update emote set")
		return nil, errors.ErrInternalServerError.SetDetail(err.Error())
	}

	esm.EmoteSetBuilder.Update.Clear()
	logrus.WithFields(logrus.Fields{
		"actor_id":     opt.Actor.ID,
		"emote_set_id": esm.EmoteSetBuilder.EmoteSet.ID,
	}).Info("Emote Set Updated")
	return esm, nil
}

// SetEmote: enable, edit or disable active emotes in the set
func (esm *EmoteSetMutation) SetEmote(ctx context.Context, inst mongo.Instance, opt SetEmoteSetEmoteOptions) (*EmoteSetMutation, error) {
	esm.l.Lock()
	defer esm.l.Unlock()
	if esm.EmoteSetBuilder == nil || esm.EmoteSetBuilder.EmoteSet == nil {
		return nil, errors.ErrInternalIncompleteMutation
	}
	if len(opt.Emotes) == 0 {
		return nil, errors.ErrMissingRequiredField.SetDetail("EmoteIDs")
	}

	// Can actor do this?
	actor := opt.Actor
	if actor == nil || actor.HasPermission(structures.RolePermissionEditEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{"MISSING_PERMISSION": "EDIT_EMOTE_SET"})
	}

	// Get relevant data
	targetEmoteIDs := []primitive.ObjectID{}
	targetEmoteMap := map[primitive.ObjectID]*structures.ActiveEmote{}
	set := esm.EmoteSetBuilder.EmoteSet
	{
		// Find emote set owner
		if set.Owner == nil {
			cur, err := inst.Collection(structures.CollectionNameUsers).Aggregate(ctx, append(mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"_id": set.OwnerID}}},
			}, aggregations.UserRelationEditors...))
			cur.Next(ctx)
			if err = multierror.Append(err, cur.Decode(set.Owner), cur.Close(ctx)).ErrorOrNil(); err != nil {
				if err == mongo.ErrNoDocuments {
					return nil, errors.ErrUnknownUser.SetDetail("emote set owner")
				}
				logrus.WithError(err).Error("failed to find emote set owner")
				return nil, err
			}
		}

		// Fetch set emotes
		if len(set.Emotes) > 0 {
			cur, err := inst.Collection(structures.CollectionNameEmoteSets).Aggregate(ctx, append(mongo.Pipeline{
				// Match only the target set
				{{Key: "$match", Value: bson.M{"_id": set.ID}}},
			}, aggregations.EmoteSetRelationActiveEmotes...))
			if err = multierror.Append(err, cur.All(ctx, set.Emotes)).ErrorOrNil(); err != nil {
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
		cur, err := inst.Collection(structures.CollectionNameEmotes).Aggregate(ctx, append(mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": targetEmoteIDs}}}},
		}, aggregations.GetEmoteRelationshipOwner(aggregations.UserRelationshipOptions{Editors: true})...))
		err = multierror.Append(err, cur.All(ctx, targetEmotes)).ErrorOrNil()
		if err != nil {
			logrus.WithError(err).Error("failed to fetch target emotes during emote set mutation")
			return nil, err
		}
		for _, e := range targetEmotes {
			if v, ok := targetEmoteMap[e.ID]; ok && v.Emote != nil {
				targetEmoteMap[e.ID] = v
			}
		}
	}

	// The actor must have access to the emote set
	if set.OwnerID != actor.ID && !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
		if set.Privileged && !actor.HasPermission(structures.RolePermissionSuperAdministrator) {
			return nil, errors.ErrInsufficientPrivilege.SetDetail("emote set is privileged")
		}
		if set.Owner != nil {
			for _, ed := range set.Owner.Editors {
				if ed.ID != actor.ID {
					continue
				}
				if !ed.HasPermission(structures.UserEditorPermissionModifyActiveEmotes) {
					return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{
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
		tgt.Name = tgt.Emote.Name

		switch opt.Action {
		// ADD EMOTE
		case ListItemActionAdd:
			// Handle emote privacy
			if utils.BitField.HasBits(int64(tgt.Emote.Flags), int64(structures.EmoteFlagsPrivate)) {
				usable := false
				// Usable if user has Bypass Privacy permission
				if actor.HasPermission(structures.RolePermissionBypassPrivacy) {
					usable = true
				}
				// Usable if actor is an editor of emote owner
				// and has the correct permission
				if tgt.Emote.Owner != nil {
					var editor *structures.UserEditor
					for _, ed := range tgt.Emote.Owner.Editors {
						if opt.Actor.ID == ed.ID {
							editor = ed
							break
						}
					}
					if editor.HasPermission(structures.UserEditorPermissionUsePrivateEmotes) {
						usable = true
					}
				}

				if !usable {
					return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{
						"EMOTE_ID": tgt.ID.Hex(),
					}).SetDetail("emote is private")
				}
			}

			// Check for conflicts with existing emotes
			for _, e := range set.Emotes {
				// Cannot enable the same emote twice
				if tgt.ID == e.ID {
					return nil, errors.ErrEmoteAlreadyEnabled
				}
				// Cannot have the same emote name as another active emote
				if tgt.Name == e.Name {
					return nil, errors.ErrEmoteNameConflict
				}
			}

			// Add active emote
			esm.EmoteSetBuilder.AddActiveEmote(tgt.ID, tgt.Name, time.Now())
		case ListItemActionUpdate, ListItemActionRemove:
			// The emote must already be active
			for _, e := range set.Emotes {
				if e.ID == tgt.ID {
					return nil, errors.ErrEmoteNotEnabled.SetFields(errors.Fields{
						"EMOTE_ID": tgt.ID.Hex(),
					})
				}
			}

			if opt.Action == ListItemActionUpdate {
				esm.EmoteSetBuilder.UpdateActiveEmote(tgt.ID, tgt.Name)
			} else if opt.Action == ListItemActionRemove {
				esm.EmoteSetBuilder.RemoveActiveEmote(tgt.ID)
			}
		}
	}

	// Update the document
	if err := inst.Collection(structures.CollectionNameEmoteSets).FindOneAndUpdate(
		ctx,
		bson.M{
			"_id": set.ID,
		},
		esm.EmoteSetBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(esm.EmoteSetBuilder.EmoteSet); err != nil {
		logrus.WithError(err).WithField("emote_set_id", set.ID).Error("mongo, failed to update emote set")
		return nil, errors.ErrInternalServerError.SetDetail(err.Error())
	}

	esm.EmoteSetBuilder.Update.Clear()
	return esm, nil
}

type EmoteSetMutationOptions struct {
	Actor          *structures.User
	SkipValidation bool
}

type SetEmoteSetEmoteOptions struct {
	Actor    *structures.User
	Emotes   []*structures.ActiveEmote
	Channels []primitive.ObjectID
	Action   ListItemAction
}
