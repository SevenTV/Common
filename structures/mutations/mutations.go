package mutations

import (
	"github.com/SevenTV/Common/structures"
)

type ListItemAction string

const (
	ListItemActionAdd    ListItemAction = "ADD"
	ListItemActionUpdate ListItemAction = "UPDATE"
	ListItemActionRemove ListItemAction = "REMOVE"
)

type UserMutation struct {
	UserBuilder *structures.UserBuilder
}

type RoleMutation struct {
	RoleBuilder *structures.RoleBuilder
}

type EmoteMutation struct {
	EmoteBuilder *structures.EmoteBuilder
}

type EmoteSetMutation struct {
	EmoteSetBuilder *structures.EmoteSetBuilder
}
