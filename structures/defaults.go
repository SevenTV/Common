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
