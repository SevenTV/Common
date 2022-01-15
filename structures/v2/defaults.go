package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var DeletedUser = &User{
	ID:           primitive.NilObjectID,
	Login:        "*deleted_user",
	DisplayName:  "*DeletedUser",
	Email:        "",
	Editors:      nil,
	TokenVersion: "",
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
