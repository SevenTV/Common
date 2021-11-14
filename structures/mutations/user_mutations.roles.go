package mutations

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetRole: add or remove a role for the user
func (um *UserMutation) SetRole(ctx context.Context, inst mongo.Instance, opt SetUserRoleOptions) (*UserMutation, error) {
	if um.UserBuilder == nil || um.UserBuilder.User == nil {
		return nil, structures.ErrIncompleteMutation
	}

	// Check for actor's permission to do this
	if opt.Actor != nil && !opt.Actor.HasPermission(structures.RolePermissionManageRoles) {
		return nil, structures.ErrInsufficientPrivilege
	}

	target := um.UserBuilder.User
	// Change the role
	switch opt.Action {
	case ListItemActionAdd:
		um.UserBuilder.Update.AddToSet("role_ids", opt.RoleID)
	case ListItemActionRemove:
		um.UserBuilder.Update.Pull("role_ids", opt.RoleID)
	}

	if err := inst.Collection(mongo.CollectionNameUsers).FindOneAndUpdate(
		ctx,
		bson.M{"_id": target.ID},
		um.UserBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(target); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, structures.ErrInternalError
	}

	return um, nil
}

type SetUserRoleOptions struct {
	RoleID primitive.ObjectID
	Actor  *structures.User
	Action ListItemAction
}
