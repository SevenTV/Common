package mutations

import "github.com/SevenTV/Common/structures"

type ListItemAction string

const (
	ListItemActionAdd    ListItemAction = "ADD"
	ListItemActionUpdate                = "UPDATE"
	ListItemActionRemove                = "REMOVE"
)

type EmoteMutation struct {
	EmoteBuilder *structures.EmoteBuilder
	Actor        *structures.User
}
