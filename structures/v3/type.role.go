package structures

import (
	"github.com/seventv/common/utils"
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
	Default bool `json:"default" bson:"default,omitempty"`
	// whether or not the role
	Invisible bool `json:"invisible" bson:"invisible,omitempty"`

	// the id of the linked role on discord
	DiscordID uint64 `json:"discord_id,omitempty" bson:"discord_id,omitempty"`
}

// HasPermissionBit: Check for specific bit in the role's allowed permissions
func (r Role) HasPermissionBit(bit RolePermission) bool {
	sum := r.Allowed

	return utils.BitField.HasBits(int64(sum), int64(bit))
}

// RolePermission Role permission bits
type RolePermission int64

// Emotes
// Range: 1 << 0 - 1 << 5
const (
	RolePermissionCreateEmote    RolePermission = 1 << 0 // 1 - Allows creating emotes
	RolePermissionEditEmote      RolePermission = 1 << 1 // 2 - Allows editing emotes
	RolePermissionCreateEmoteSet RolePermission = 1 << 2 // 4 - Allows creating emote sets
	RolePermissionEditEmoteSet   RolePermission = 1 << 3 // 8 - Allows creating and modifying emote sets
)

// Unused Space
// Range: 1 << 6 - 1 << 12
const ()

// User / Misc / Special
// Range: 1 << 13 - 1 << 1 << 29
const (
	RolePermissionReportCreate RolePermission = 1 << 13 // 8192 - Allows creating reports
	RolePermissionSendMessages RolePermission = 1 << 14 // 16384 - Allows sending messages (i.e comments or user inboxs)

	RolePermissionFeatureZeroWidthEmoteType      RolePermission = 1 << 23 // 8388608 - Allows using the Zero-Width emote type
	RolePermissionFeatureProfilePictureAnimation RolePermission = 1 << 24 // 16777216 - Allows the user's profile picture to be animated
)

// Moderation
// Range: 1 << 30 - 1 << 53
const (
	RolePermissionManageBans      RolePermission = 1 << 30 // 1073741824 - (Mod) Allows creating or deleting bans
	RolePermissionManageRoles     RolePermission = 1 << 31 // 2147483648 - (Mod) Allows creating, deleting and assigning roles to users
	RolePermissionManageReports   RolePermission = 1 << 32 // 4294967296 - (Mod) Allows managing reports
	RolePermissionManageUsers     RolePermission = 1 << 33 // 8589934592 - (Mod) Allows managing users
	RolePermissionEditAnyEmote    RolePermission = 1 << 41 // 2199023255552 - (Mod) Allows editing any emote
	RolePermissionEditAnyEmoteSet RolePermission = 1 << 42 // 4398046511104 - (Mod) Allows editing any emote set, unless it is a privileged set
	RolePermissionBypassPrivacy   RolePermission = 1 << 48 // 281474976710656 - (Mod) Lets the user see all non-public content
)

// Administration

// Range: 1 << 54 - 1 << 63
const (
	RolePermissionSuperAdministrator RolePermission = 1 << 62 // 4611686018427387904 - (Admin) GRANTS EVERY PERMISSION /!\
	RolePermissionManageNews         RolePermission = 1 << 54 // 18014398509481984 - (Admin) Allows creating and editing news
	RolePermissionManageStack        RolePermission = 1 << 55 // 36028797018963968 - (Admin) Allows managing the application stack
	RolePermissionManageCosmetics    RolePermission = 1 << 56 // 72057594037927936 - (Admin) Allows managing user cosmetics
	RolePermissionRunJobs            RolePermission = 1 << 57 // 144115188075855872 - (Admin) Allows firing processing jobs
)

// All permissions
const (
	RolePermissionAll = RolePermissionCreateEmote | RolePermissionEditEmote | RolePermissionEditEmoteSet |
		RolePermissionReportCreate | RolePermissionFeatureZeroWidthEmoteType | RolePermissionFeatureProfilePictureAnimation |
		RolePermissionManageBans | RolePermissionManageRoles | RolePermissionManageReports |
		RolePermissionEditAnyEmote | RolePermissionEditAnyEmoteSet | RolePermissionSuperAdministrator |
		RolePermissionManageNews | RolePermissionManageStack | RolePermissionManageCosmetics
)

type RoleBuilder struct {
	Update UpdateMap
	Role   Role
}

// NewRoleBuilder: create a new role builder
func NewRoleBuilder(role Role) *RoleBuilder {
	return &RoleBuilder{
		Update: UpdateMap{},
		Role:   role,
	}
}

func (rb *RoleBuilder) SetName(name string) *RoleBuilder {
	rb.Role.Name = name
	rb.Update.Set("name", name)
	return rb
}

func (rb *RoleBuilder) SetPosition(pos int32) *RoleBuilder {
	rb.Role.Position = pos
	rb.Update.Set("position", pos)
	return rb
}

func (rb *RoleBuilder) SetColor(color int32) *RoleBuilder {
	rb.Role.Color = color
	rb.Update.Set("color", color)
	return rb
}

func (rb *RoleBuilder) SetAllowed(allowed RolePermission) *RoleBuilder {
	rb.Role.Allowed = allowed
	rb.Update.Set("allowed", allowed)
	return rb
}

func (rb *RoleBuilder) SetDenied(denied RolePermission) *RoleBuilder {
	rb.Role.Denied = denied
	rb.Update.Set("denied", denied)
	return rb
}
