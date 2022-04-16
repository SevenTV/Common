package mutations

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create: create the new role
func (m *Mutate) CreateRole(ctx context.Context, rb *structures.RoleBuilder, opt RoleMutationOptions) error {
	if rb == nil {
		return errors.ErrInternalIncompleteMutation()
	}
	if rb.Role.Name == "" {
		return fmt.Errorf("missing name for role")
	}

	// Check actor's permissions
	if opt.Actor != nil && !opt.Actor.HasPermission(structures.RolePermissionManageRoles) {
		return structures.ErrInsufficientPrivilege
	}

	// Create the role
	rb.Role.ID = primitive.NewObjectID()
	result, err := m.mongo.Collection(mongo.CollectionNameRoles).InsertOne(ctx, rb.Role)
	if err != nil {
		return err
	}

	// Get the newly created role
	if err = m.mongo.Collection(mongo.CollectionNameRoles).FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(rb.Role); err != nil {
		return err
	}

	return nil
}

// Edit: edit the role. Modify the RoleBuilder beforehand!
func (m *Mutate) EditRole(ctx context.Context, rb *structures.RoleBuilder, opt RoleEditOptions) error {
	if rb == nil {
		return errors.ErrInternalIncompleteMutation()
	}

	// Check actor's permissions
	actor := opt.Actor
	if actor != nil {
		if !actor.HasPermission(structures.RolePermissionManageRoles) {
			return structures.ErrInsufficientPrivilege
		}
		if len(opt.Actor.Roles) > 0 {
			// ensure that the actor's role is higher than the role being deleted
			actor.SortRoles()
			highestRole := actor.Roles[0]
			if opt.OriginalPosition >= highestRole.Position {
				return structures.ErrInsufficientPrivilege
			}
		}
	}

	// Update the role
	if err := m.mongo.Collection(mongo.CollectionNameRoles).FindOneAndUpdate(
		ctx,
		bson.M{"_id": rb.Role.ID},
		rb.Update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(rb.Role); err != nil {
		return err
	}

	return nil
}

// Delete: delete the role
func (m *Mutate) DeleteRole(ctx context.Context, rb *structures.RoleBuilder, opt RoleMutationOptions) error {
	if rb == nil {
		return structures.ErrIncompleteMutation
	}

	// Check actor's permissions
	actor := opt.Actor
	if actor != nil {
		if !actor.HasPermission(structures.RolePermissionManageRoles) {
			return structures.ErrInsufficientPrivilege
		}
		if len(opt.Actor.Roles) > 0 {
			// ensure that the actor's role is higher than the role being deleted
			actor.SortRoles()
			highestRole := actor.Roles[0]
			if rb.Role.Position >= highestRole.Position {
				return structures.ErrInsufficientPrivilege
			}
		}
	}

	// Delete the role
	if _, err := m.mongo.Collection(mongo.CollectionNameRoles).DeleteOne(ctx, bson.M{"_id": rb.Role.ID}); err != nil {
		return err
	}

	// Remove the role from any user who had it
	_, err := m.mongo.Collection(mongo.CollectionNameUsers).UpdateMany(ctx, bson.M{
		"role_ids": rb.Role.ID,
	}, bson.M{
		"$pull": bson.M{
			"role_ids": rb.Role.ID,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

type RoleMutationOptions struct {
	Actor *structures.User
}

type RoleEditOptions struct {
	Actor            *structures.User
	OriginalPosition int32
}
