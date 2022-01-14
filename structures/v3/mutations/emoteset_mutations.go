package mutations

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/structures/v3/aggregations"
	"github.com/SevenTV/Common/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create: create the new emote set
func (esm *EmoteSetMutation) Create(ctx context.Context, inst mongo.Instance, opt EmoteSetMutationOptions) (*EmoteSetMutation, error) {
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

	logrus.WithFields(logrus.Fields{
		"actor_id":     opt.Actor,
		"emote_set_id": esm.EmoteSetBuilder.EmoteSet.ID,
	}).Info("Emote Set Created")
	return esm, nil
}

func (esm *EmoteSetMutation) SetEmote(ctx context.Context, inst mongo.Instance, opt SetEmoteSetEmoteOptions) (*EmoteSetMutation, error) {
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
			if err := inst.Collection(structures.CollectionNameUsers).FindOne(ctx, bson.M{
				"_id": esm.EmoteSetBuilder.EmoteSet.OwnerID,
			}).Decode(set.Owner); err != nil {
				if err == mongo.ErrNoDocuments {
					return nil, errors.ErrUnknownUser.SetDetail("emote set owner")
				}
				logrus.WithError(err).Error("failed to find emote set owner")
				return nil, err
			}
		}

		// Fetch set emotes
		if len(set.Emotes) > 0 {
			pipeline := append(mongo.Pipeline{
				{{ // Match only the target set
					Key:   "$match",
					Value: bson.M{"_id": set.ID},
				}},
			}, aggregations.EmoteSetRelationActiveEmotes...) // add "emote" field in active emotes

			cur, err := inst.Collection(structures.CollectionNameEmoteSets).Aggregate(ctx, pipeline)
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
			if v, ok := targetEmoteMap[e.ID]; ok {
				targetEmoteMap[e.ID] = v
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
		switch opt.Action {
		// ADD EMOTE
		case ListItemActionAdd:
			for _, e := range set.Emotes {
				// Cannot enable the same emote twice
				if tgt.ID == e.ID {
					return nil, errors.ErrEmoteAlreadyEnabled
				}
				// Cannot have the same emote name as another active emote
				if (tgt.Alias != "" && e.Alias != "") && tgt.Alias == e.Alias || tgt.Emote.Name == e.Emote.Name {
					return nil, errors.ErrEmoteNameConflict
				}

				// Emote is private
				if utils.BitField.HasBits(int64(tgt.Emote.Flags), int64(structures.EmoteFlagsPrivate)) {
					usable := false
					if actor.HasPermission(structures.RolePermissionBypassPrivacy) {
						usable = true
					}

					if !usable {
						return nil, errors.ErrInsufficientPrivilege.SetFields(errors.Fields{
							"EMOTE_ID": tgt.ID.Hex(),
						}).SetDetail("emote is private")
					}
				}
			}
		}
	}

	return esm, nil
}

type EmoteSetMutationOptions struct {
	Actor *structures.User
}

type SetEmoteSetEmoteOptions struct {
	Actor    *structures.User
	Emotes   []*structures.ActiveEmote
	Channels []primitive.ObjectID
	Action   ListItemAction
}
