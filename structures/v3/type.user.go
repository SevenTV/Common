package structures

import (
	"fmt"
	"sort"
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User A standard app user object
type User struct {
	ID ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	// the type of this user. empty when a regular user, but could also be "BOT" or "SYSTEM"
	UserType UserType `json:"type,omitempty" bson:"type,omitempty"`
	// the user's username
	Username string `json:"username" bson:"username"`
	// the user's display name
	DisplayName string `json:"display_name" bson:"display_name"`
	// the user's discriminatory space
	Discriminator string `json:"discriminator" bson:"discriminator"`
	// the user's email
	Email string `json:"email" bson:"email"`
	// list of role IDs directly bound to the user (not via an entitlement)
	RoleIDs []ObjectID `json:"role_ids" bson:"role_ids"`
	// the user's editors
	Editors []UserEditor `json:"editors" bson:"editors"`
	// the user's avatar URL
	AvatarID string `json:"avatar_id" bson:"avatar_id"`
	// the user's avatar
	Avatar *UserAvatar `json:"avatar" bson:"avatar"`
	// the user's biography
	Biography string `json:"biography" bson:"biography"`
	// token version. When this value changes all existing auth tokens are invalidated
	TokenVersion float64 `json:"token_version" bson:"token_version"`
	// third party connections. Who's the third party now?
	Connections UserConnectionList `json:"connections" bson:"connections"`
	// the ID of users who have been blocked by the user
	BlockedUserIDs []ObjectID `json:"blocked_user_ids,omitempty" bson:"blocked_user_ids,omitempty"`
	// persisted non-structural data that can be used internally for querying
	State UserState `json:"-" bson:"state"`
	// Special treatments attributed to the user;
	// This is used to run experiments
	Treatments []string `json:"treatments" bson:"treatments"`

	// Relational

	Emotes       []Emote                 `json:"emotes" bson:"emotes,skip,omitempty"`
	OwnedEmotes  []Emote                 `json:"owned_emotes" bson:"owned_emotes,skip,omitempty"`
	Bans         []Ban                   `json:"bans" bson:"bans,skip,omitempty"`
	Entitlements []Entitlement[bson.Raw] `json:"entitlements" bson:"entitlements,skip,omitempty"`

	// API-specific

	Roles     []Role       `json:"roles" bson:"roles,skip,omitempty"`
	EditorOf  []UserEditor `json:"editor_of" bson:"editor_of,skip,omitempty"`
	AvatarURL string       `json:"avatar_url" bson:"-"`
}

type UserState struct {
	RolePosition int `json:"-" bson:"role_position"`
}

type UserAvatar struct {
	ID         primitive.ObjectID  `json:"id" bson:"id"`
	InputFile  ImageFile           `json:"input_file" bson:"input_file"`
	ImageFiles []ImageFile         `json:"image_files" bson:"image_files"`
	PendingID  *primitive.ObjectID `json:"pending_id,omitempty" bson:"pending_id,omitempty"`
}

// HasPermission checks relational roles against a permission bit
func (u *User) HasPermission(bit RolePermission) bool {
	total := u.FinalPermission()

	if (total & RolePermissionSuperAdministrator) != 0 {
		return true
	}

	return utils.BitField.HasBits(int64(total), int64(bit))
}

func (u *User) FinalPermission() (total RolePermission) {
	for _, r := range u.Roles {
		total &= ^r.Denied
		total |= r.Allowed
	}
	return
}

func (u *User) AddRoles(roles ...Role) {
	for _, r := range roles {
		exists := false
		for _, ur := range u.Roles {
			if r.ID == ur.ID {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		u.Roles = append(u.Roles, r)
	}
}

func (u *User) SortRoles() {
	if len(u.Roles) == 0 {
		return
	}
	sort.Slice(u.Roles, func(i, j int) bool {
		a := u.Roles[i]
		b := u.Roles[j]

		return a.Position > b.Position
	})
}

func (u *User) GetHighestRole() Role {
	u.SortRoles()
	if len(u.Roles) == 0 {
		return NilRole
	}

	return u.Roles[0]
}

// GetEditor returns the specified user editor
func (u *User) GetEditor(id primitive.ObjectID) (UserEditor, bool, int) {
	for i, ue := range u.Editors {
		if ue.ID == id {
			return ue, true, i
		}
	}
	return UserEditor{}, false, -1
}

func (e User) WebURL(origin string) string {
	return fmt.Sprintf("%s/users/%s", origin, e.ID.Hex())
}

type UserDiscriminator uint8

type UserEditor struct {
	ID ObjectID `json:"id" bson:"id"`
	// The permissions this editor has
	Permissions UserEditorPermission `json:"permissions" bson:"permissions"`
	// Whether or not that editor will be visible on the user's profile page
	Visible bool `json:"visible" bson:"visible"`

	AddedAt time.Time `json:"added_at,omitempty" bson:"added_at,omitempty"`

	// Relational
	User *User `json:"user" bson:"user,skip,omitempty"`
}

// HasPermission: check whether or not the editor has a permission
func (ed *UserEditor) HasPermission(bit UserEditorPermission) bool {
	return utils.BitField.HasBits(int64(ed.Permissions), int64(bit))
}

type UserEditorPermission int32

const (
	UserEditorPermissionModifyEmotes      UserEditorPermission = 1 << 0 // 1 - Allows modifying emotes in the user's active emote sets
	UserEditorPermissionUsePrivateEmotes  UserEditorPermission = 1 << 1 // 2 - Allows using the user's private emotes
	UserEditorPermissionManageProfile     UserEditorPermission = 1 << 2 // 4 - Allows managing the user's public profile
	UserEditorPermissionManageOwnedEmotes UserEditorPermission = 1 << 3 // 8 - Allows managing the user's owned emotes
	UserEditorPermissionManageEmoteSets   UserEditorPermission = 1 << 4 // 16 - Allows managing the user's owned emote sets
	UserEditorPermissionManageBilling     UserEditorPermission = 1 << 5 // 32 - Allows managing billing and payments, such as subscriptions
	UserEditorPermissionManageEditors     UserEditorPermission = 1 << 6 // 64 - Allows adding or removing editors for the user
	UserEditorPermissionViewMessages      UserEditorPermission = 1 << 7 // 128 - Allows viewing the user's private messages, such as inbox
)
