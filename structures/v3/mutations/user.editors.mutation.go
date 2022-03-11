package mutations

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func (um *UserMutation) Editors(ctx context.Context, inst mongo.Instance, opt UserEditorsOptions) (*UserMutation, error) {
	if um.UserBuilder == nil || um.UserBuilder.User == nil {
		return nil, errors.ErrInternalIncompleteMutation()
	}

	// Fetch relevant data
	target := um.UserBuilder.User
	editor := opt.Editor
	if editor == nil {
		return nil, errors.ErrUnknownUser()
	}

	// Check permissions
	// The actor must either be privileged, the target user, or an editor with sufficient permissions
	actor := opt.Actor
	if actor.ID != target.ID && !actor.HasPermission(structures.RolePermissionManageUsers) {
		ed, ok, _ := target.GetEditor(actor.ID)
		if !ok {
			return nil, errors.ErrInsufficientPrivilege()
		}
		// actor is an editor of target but they must also have "Manage Editors" permission to do this
		if !ed.HasPermission(structures.UserEditorPermissionManageEditors) {
			// the actor is allowed to *remove* themselve as an editor
			if !(actor.ID == editor.ID && opt.Action == ListItemActionRemove) {
				return nil, errors.ErrInsufficientPrivilege().SetDetail("You don't have permission to manage this user's editors")
			}
		}
	}

	switch opt.Action {
	// add editor
	case ListItemActionAdd:
		um.UserBuilder.AddEditor(editor.ID, opt.EditorPermissions, opt.EditorVisible)
	case ListItemActionUpdate:
		um.UserBuilder.UpdateEditor(editor.ID, opt.EditorPermissions, opt.EditorVisible)
	case ListItemActionRemove:
		um.UserBuilder.RemoveEditor(editor.ID)
	}

	// Write mutation
	if _, err := inst.Collection(mongo.CollectionNameUsers).UpdateOne(ctx, bson.M{
		"_id": target.ID,
	}, um.UserBuilder.Update); err != nil {
		return nil, errors.ErrInternalServerError().SetDetail(err.Error())
	}

	um.UserBuilder.Update.Clear()
	return um, nil
}

type UserEditorsOptions struct {
	Actor             *structures.User
	Editor            *structures.User
	EditorPermissions structures.UserEditorPermission
	EditorVisible     bool
	Action            ListItemAction
}
