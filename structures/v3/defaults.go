package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

var DeletedUser = &User{
	ID:            primitive.NilObjectID,
	UserType:      UserTypeSystem,
	Username:      "*deleted_user",
	DisplayName:   "*DeletedUser",
	Discriminator: "0000",
	RoleIDs:       []primitive.ObjectID{},
	Editors:       []*UserEditor{},
	TokenVersion:  0,
	Connections:   []*UserConnection{},
}

var DeletedEmote = &Emote{
	ID:       primitive.NilObjectID,
	OwnerID:  DeletedUser.ID,
	Name:     "*UnknownEmote",
	Flags:    0,
	Tags:     []string{},
	Owner:    DeletedUser,
	Channels: []*User{},
}

var RevocationRole = &Role{
	ID:       [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32},
	Name:     "NORIGHTS",
	Denied:   RolePermissionAll,
	Position: 0,
}

var NilRole = &Role{
	ID:       primitive.NilObjectID,
	Name:     "NULL",
	Position: 0,
}
