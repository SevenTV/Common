package mutations

type ListItemAction string

const (
	ListItemActionAdd    ListItemAction = "ADD"
	ListItemActionUpdate                = "UPDATE"
	ListItemActionRemove                = "REMOVE"
)
