package mutations

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetRole: add or remove a role for the user
func (m *Mutate) SetRole(ctx context.Context, ub *structures.UserBuilder, opt SetUserRoleOptions) error {
	if ub == nil || ub.User == nil {
		return structures.ErrIncompleteMutation
	} else if ub.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}

	// Check for actor's permission to do this
	actor := opt.Actor
	if actor != nil {
		if !actor.HasPermission(structures.RolePermissionManageRoles) {
			return structures.ErrInsufficientPrivilege
		}
		if len(actor.Roles) == 0 {
			return structures.ErrInsufficientPrivilege
		}
		highestRole := actor.Roles[0]
		if opt.Role.Position >= highestRole.Position {
			return structures.ErrInsufficientPrivilege
		}
	}

	target := ub.User
	// Change the role
	switch opt.Action {
	case ListItemActionAdd:
		ub.Update.AddToSet("role_ids", opt.Role.ID)
	case ListItemActionRemove:
		ub.Update.Pull("role_ids", opt.Role.ID)
	}

	if err := m.mongo.Collection(mongo.CollectionNameUsers).FindOneAndUpdate(
		ctx,
		bson.M{"_id": target.ID},
		ub.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(target); err != nil {
		logrus.WithError(err).Error("mongo")
		return structures.ErrInternalError
	}

	ub.MarkAsTainted()
	return nil
}

type SetUserRoleOptions struct {
	Role   *structures.Role
	Actor  *structures.User
	Action ListItemAction
}
