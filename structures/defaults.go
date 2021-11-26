package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

var DeletedUser = &User{
	ID:            primitive.NilObjectID,
	UserType:      "",
	Username:      "*deleted_user",
	DisplayName:   "*DeletedUser",
	Discriminator: "0000",
	Email:         "",
	ChannelEmotes: nil,
	RoleIDs:       nil,
	Editors:       nil,
	AvatarURL:     "",
	Biography:     "",
	TokenVersion:  0,
	Connections:   nil,
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
