package mutations

import (
	"context"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Mutate) ModifyUserEditors(ctx context.Context, ub *structures.UserBuilder, opt UserEditorsOptions) error {
	if ub == nil {
		return errors.ErrInternalIncompleteMutation()
	} else if ub.IsTainted() {
		return errors.ErrMutateTaintedObject()
	}

	// Fetch relevant data
	target := ub.User
	editor := opt.Editor
	if editor == nil {
		return errors.ErrUnknownUser()
	}

	// Check permissions
	// The actor must either be privileged, the target user, or an editor with sufficient permissions
	actor := opt.Actor
	if actor.ID != target.ID && !actor.HasPermission(structures.RolePermissionManageUsers) {
		ed, ok, _ := target.GetEditor(actor.ID)
		if !ok {
			return errors.ErrInsufficientPrivilege()
		}
		// actor is an editor of target but they must also have "Manage Editors" permission to do this
		if !ed.HasPermission(structures.UserEditorPermissionManageEditors) {
			// the actor is allowed to *remove* themselve as an editor
			if !(actor.ID == editor.ID && opt.Action == structures.ListItemActionRemove) {
				return errors.ErrInsufficientPrivilege().SetDetail("You don't have permission to manage this user's editors")
			}
		}
	}

	switch opt.Action {
	// add editor
	case structures.ListItemActionAdd:
		ub.AddEditor(editor.ID, opt.EditorPermissions, opt.EditorVisible)
	case structures.ListItemActionUpdate:
		ub.UpdateEditor(editor.ID, opt.EditorPermissions, opt.EditorVisible)
	case structures.ListItemActionRemove:
		ub.RemoveEditor(editor.ID)
	}

	// Write mutation
	if _, err := m.mongo.Collection(mongo.CollectionNameUsers).UpdateOne(ctx, bson.M{
		"_id": target.ID,
	}, ub.Update); err != nil {
		return errors.ErrInternalServerError().SetDetail(err.Error())
	}

	ub.MarkAsTainted()
	return nil
}

type UserEditorsOptions struct {
	Actor             *structures.User
	Editor            *structures.User
	EditorPermissions structures.UserEditorPermission
	EditorVisible     bool
	Action            structures.ListItemAction
}
