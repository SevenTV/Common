package structures

import (
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	// the role's name
	Name string `json:"name" bson:"name"`
	// the role's privilege position
	Position int32 `json:"position" bson:"position"`
	// the role's display color
	Color int32 `json:"color" bson:"color"`
	// the role's allowed permission bits
	Allowed RolePermission `json:"allowed" bson:"allowed"`
	// the role's denied permission bits
	Denied RolePermission `json:"denied" bson:"denied"`
	// whether or not this role is the default role
	Default bool `json:"default" bson:"default"`
}

// HasPermissionBit: Check for specific bit in the role's allowed permissions
func (r *Role) HasPermissionBit(bit RolePermission) bool {
	sum := r.Allowed

	return utils.BitField.HasBits(int64(sum), int64(bit))
}

// RolePermission: Role permission bits
type RolePermission int64

const (
	// Emotes
	// Range: 1 << 1 - 1 << 8
	RolePermissionEmotesCreate int64 = 1 << iota // 1 - Allows creating emotes
	RolePermissionEmotesEdit                     // 2 - Allows editing owned emotes

	// Reporting
	// Range: 1 << 9- 1 << 12

	// Moderation
	// Range: 1 << 13 - 1 << 24

	// Administration
	// Range: 1 << 25 - 1 << 32
)

type RoleBuilder struct {
	Update UpdateMap
	Role   *Role
}

// NewRoleBuilder: create a new role builder
func NewRoleBuilder(role *Role) *RoleBuilder {
	return &RoleBuilder{
		Update: UpdateMap{},
		Role:   role,
	}
}
