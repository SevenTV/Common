package mutations

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create: create the new role
func (rm *RoleMutation) Create(ctx context.Context, inst mongo.Instance, opt RoleMutationOptions) (*RoleMutation, error) {
	if rm.RoleBuilder == nil || rm.RoleBuilder.Role == nil {
		return nil, structures.ErrIncompleteMutation
	}
	if rm.RoleBuilder.Role.Name == "" {
		return nil, fmt.Errorf("missing name for role")
	}

	// Check actor's permissions
	if opt.Actor != nil && !opt.Actor.HasPermission(structures.RolePermissionManageRoles) {
		return nil, structures.ErrInsufficientPrivilege
	}

	// Create the role
	rm.RoleBuilder.Role.ID = primitive.NewObjectID()
	result, err := inst.Collection(mongo.CollectionNameRoles).InsertOne(ctx, rm.RoleBuilder.Role)
	if err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Get the newly created role
	if inst.Collection(mongo.CollectionNameRoles).FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(rm.RoleBuilder.Role); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"role_id": rm.RoleBuilder.Role.ID,
	}).Info("Role Created")
	return rm, nil
}

// Edit: edit the role. Modify the RoleBuilder beforehand!
func (rm *RoleMutation) Edit(ctx context.Context, inst mongo.Instance, opt RoleEditOptions) (*RoleMutation, error) {
	if rm.RoleBuilder == nil || rm.RoleBuilder.Role == nil {
		return nil, structures.ErrIncompleteMutation
	}

	// Check actor's permissions
	actor := opt.Actor
	if actor != nil {
		if !actor.HasPermission(structures.RolePermissionManageRoles) {
			return nil, structures.ErrInsufficientPrivilege
		}
		if len(opt.Actor.Roles) > 0 {
			// ensure that the actor's role is higher than the role being deleted
			actor.SortRoles()
			highestRole := actor.Roles[0]
			if opt.OriginalPosition >= highestRole.Position {
				return nil, structures.ErrInsufficientPrivilege
			}
		}
	}

	// Update the role
	if err := inst.Collection(mongo.CollectionNameRoles).FindOneAndUpdate(
		ctx,
		bson.M{"_id": rm.RoleBuilder.Role.ID},
		rm.RoleBuilder.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(rm.RoleBuilder.Role); err != nil {
		return nil, err
	}

	return rm, nil
}

// Delete: delete the role
func (rm *RoleMutation) Delete(ctx context.Context, inst mongo.Instance, opt RoleMutationOptions) (*RoleMutation, error) {
	if rm.RoleBuilder == nil || rm.RoleBuilder.Role == nil {
		return nil, structures.ErrIncompleteMutation
	}

	// Check actor's permissions
	actor := opt.Actor
	if actor != nil {
		if !actor.HasPermission(structures.RolePermissionManageRoles) {
			return nil, structures.ErrInsufficientPrivilege
		}
		if len(opt.Actor.Roles) > 0 {
			// ensure that the actor's role is higher than the role being deleted
			actor.SortRoles()
			highestRole := actor.Roles[0]
			if rm.RoleBuilder.Role.Position >= highestRole.Position {
				return nil, structures.ErrInsufficientPrivilege
			}
		}
	}

	// Delete the role
	if _, err := inst.Collection(mongo.CollectionNameRoles).DeleteOne(ctx, bson.M{"_id": rm.RoleBuilder.Role.ID}); err != nil {
		logrus.WithError(err).Error("mongo")
		return nil, err
	}

	// Remove the role from any user who had it
	ur, err := inst.Collection(mongo.CollectionNameUsers).UpdateMany(ctx, bson.M{
		"role_ids": rm.RoleBuilder.Role.ID,
	}, bson.M{
		"$pull": bson.M{
			"role_ids": rm.RoleBuilder.Role.ID,
		},
	})
	if err != nil {
		logrus.WithError(err).Error("mongo, failed to remove deleted role from user assignments")
	}

	logrus.WithFields(logrus.Fields{
		"role_id":       rm.RoleBuilder.Role.ID,
		"users_updated": ur.ModifiedCount,
	}).Info("Role Deleted")
	return rm, nil
}

type RoleMutationOptions struct {
	Actor *structures.User
}

type RoleEditOptions struct {
	Actor            *structures.User
	OriginalPosition int32
}
