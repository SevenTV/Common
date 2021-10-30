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

// Emotes
// Range: 1 << 1 - 1 << 12
const (
	RolePermissionCreateEmote     RolePermission = 1 << 0 // 1 - Allows creating emotes
	RolePermissionEditEmote       RolePermission = 1 << 1 // 2 - Allows editing / creating new versions of an emote
	RolePermissionSetChannelEmote RolePermission = 1 << 2 // 4 - Allows adding or removing channel emotes
)

// User / Misc / Special
// Range: 1 << 13 - 1 << 1 << 29
const (
	RolePermissionReportCreate RolePermission = 1 << 13 // 8192 - Allows creating reports

	RolePermissionUseZeroWidthEmoteType RolePermission = 1 << 23 // 8388608 - Allows using the Zero-Width emote type
	RolePermissionAnimateProfilePicture RolePermission = 1 << 24 // 16777216 - Allows the user's profile picture to be animated
)

// Moderation
// Range: 1 << 30 - 1 << 53
const (
	RolePermissionManageBans      RolePermission = 1 << 30 // 1073741824 - (Mod) Allows creating or deleting bans
	RolePermissionManageRoles     RolePermission = 1 << 31 // 2147483648 - (Mod) Allows creating, deleting and assigning roles to users
	RolePermissionEditAnyEmote    RolePermission = 1 << 32 // 4294967296 - (Mod) Allows editing any emote
	RolePermissionEditAnyEmoteSet RolePermission = 1 << 33 // 8589934592 - (Mod) Allows editing any emote set, unless it is a privileged set
)

// Administration
// Range: 1 << 54 - 1 << 62
const (
	RolePermissionSuperAdministrator RolePermission = 1 << 62 // 4611686018427387904 - (Admin) GRANTS EVERY PERMISSION /!\
	RolePermissionManageNews         RolePermission = 1 << 54 // 18014398509481984 - (Admin) Allows creating and editing news
	RolePermissionManageStack        RolePermission = 1 << 55 // 36028797018963968 - (Admin) Allows managing the application stack
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
