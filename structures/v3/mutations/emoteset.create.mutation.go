package mutations

import (
	"context"
	"strconv"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
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
		return nil, errors.ErrInternalIncompleteMutation()
	}
	if esm.EmoteSetBuilder.EmoteSet.Name == "" {
		return nil, errors.ErrMissingRequiredField().SetDetail("Name")
	}

	// Check actor's permissions
	if opt.Actor != nil && !opt.Actor.HasPermission(structures.RolePermissionEditEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege().SetFields(errors.Fields{"MISSING_PERMISSION": "EDIT_EMOTE_SET"})
	}

	// Create the emote set
	esm.EmoteSetBuilder.EmoteSet.ID = primitive.NewObjectID()
	result, err := inst.Collection(mongo.CollectionNameEmoteSets).InsertOne(ctx, esm.EmoteSetBuilder.EmoteSet)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Get the newly created emote set
	if err = inst.Collection(mongo.CollectionNameEmoteSets).FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(esm.EmoteSetBuilder.EmoteSet); err != nil {
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
		return nil, errors.ErrInternalIncompleteMutation()
	}

	// Check actor's permissions
	actor := opt.Actor
	set := esm.EmoteSetBuilder.EmoteSet
	if actor == nil || !actor.HasPermission(structures.RolePermissionEditEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege().SetFields(errors.Fields{"MISSING_PERMISSION": "EDIT_EMOTE_SET"})
	}
	if set.Privileged && !actor.HasPermission(structures.RolePermissionSuperAdministrator) {
		return nil, errors.ErrInsufficientPrivilege().SetDetail("emote set is privileged")
	}
	if actor.ID != set.OwnerID && !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
		return nil, errors.ErrInsufficientPrivilege().SetDetail("you do not own this emote set")
	}

	u := esm.EmoteSetBuilder.Update
	if !opt.SkipValidation {
		// Change: Name
		if _, ok := u["name"]; ok {
			// TODO: use a regex to validate
			if len(set.Name) < structures.EmoteSetNameLengthLeast || len(set.Name) >= structures.EmoteSetNameLengthMost {
				return nil, errors.ErrValidationRejected().SetFields(errors.Fields{
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
			return nil, errors.ErrInsufficientPrivilege().SetFields(errors.Fields{
				"FIELD":              "Privileged",
				"MISSING_PERMISSION": "SUPER_ADMINISTRATOR",
			})
		}

		// Change: owner
		// Must be the current owner, or have "edit any emote set" permission
		if _, ok := u["owner_id"]; ok && !actor.HasPermission(structures.RolePermissionEditAnyEmoteSet) {
			if actor.ID != set.OwnerID {
				return nil, errors.ErrInsufficientPrivilege().SetFields(errors.Fields{
					"FIELD": "OwnerID",
				}).SetDetail("you do not own this emote set")
			}
		}
	}

	// Update the document
	if err := inst.Collection(mongo.CollectionNameEmoteSets).FindOneAndUpdate(
		ctx, bson.M{
			"_id": set.ID,
		},
		esm.EmoteSetBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(esm.EmoteSetBuilder.EmoteSet); err != nil {
		logrus.WithError(err).WithField("emote_set_id", set.ID).Error("mongo, failed to update emote set")
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	esm.EmoteSetBuilder.Update.Clear()
	logrus.WithFields(logrus.Fields{
		"actor_id":     opt.Actor.ID,
		"emote_set_id": esm.EmoteSetBuilder.EmoteSet.ID,
	}).Info("Emote Set Updated")
	return esm, nil
}

type EmoteSetMutationOptions struct {
	Actor          *structures.User
	SkipValidation bool
}

type EmoteSetMutationSetEmoteOptions struct {
	Actor    *structures.User
	Emotes   []EmoteSetMutationSetEmoteItem
	Channels []primitive.ObjectID
}

type EmoteSetMutationSetEmoteItem struct {
	Action ListItemAction
	ID     primitive.ObjectID
	Name   string
	Flags  structures.ActiveEmoteFlag

	emote *structures.Emote
}
