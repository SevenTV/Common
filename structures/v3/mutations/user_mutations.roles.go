package mutations

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetRole: add or remove a role for the user
func (um *UserMutation) SetRole(ctx context.Context, inst mongo.Instance, opt SetUserRoleOptions) (*UserMutation, error) {
	if um.UserBuilder == nil || um.UserBuilder.User == nil {
		return nil, structures.ErrIncompleteMutation
	}

	// Check for actor's permission to do this
	actor := opt.Actor
	if actor != nil {
		if !actor.HasPermission(structures.RolePermissionManageRoles) {
			return nil, structures.ErrInsufficientPrivilege
		}
		if len(actor.Roles) == 0 {
			return nil, structures.ErrInsufficientPrivilege
		}
		highestRole := actor.Roles[0]
		if opt.Role.Position >= highestRole.Position {
			return nil, structures.ErrInsufficientPrivilege
		}
	}

	target := um.UserBuilder.User
	// Change the role
	switch opt.Action {
	case ListItemActionAdd:
		um.UserBuilder.Update.AddToSet("role_ids", opt.Role.ID)
	case ListItemActionRemove:
		um.UserBuilder.Update.Pull("role_ids", opt.Role.ID)
	}

	if err := inst.Collection(structures.CollectionNameUsers).FindOneAndUpdate(
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
	Role   *structures.Role
	Actor  *structures.User
	Action ListItemAction
}
