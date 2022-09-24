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
	RolePosition    int    `json:"-" bson:"role_position"`
	PendingAvatarID string `json:"-" bson:"pending_avatar_id,omitempty"`
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

type UserConnectionList []UserConnection[bson.Raw]

// Twitch returns the first Twitch user connection
func (ucl UserConnectionList) Twitch(filter ...string) (UserConnection[UserConnectionDataTwitch], int, error) {
	for idx, v := range ucl {
		if len(filter) > 0 && !utils.Contains(filter, v.ID) {
			continue // does not pass filter
		}

		if v.Platform == UserConnectionPlatformTwitch {
			conn, err := ConvertUserConnection[UserConnectionDataTwitch](v)
			return conn, idx, err
		}
	}
	return UserConnection[UserConnectionDataTwitch]{}, -1, fmt.Errorf("could not find any twitch connections")
}

// YouTube returns the first YouTube user connection
func (ucl UserConnectionList) YouTube(filter ...string) (UserConnection[UserConnectionDataYoutube], int, error) {
	for idx, v := range ucl {
		if len(filter) > 0 && !utils.Contains(filter, v.ID) {
			continue // does not pass filter
		}

		if v.Platform == UserConnectionPlatformYouTube {
			conn, err := ConvertUserConnection[UserConnectionDataYoutube](v)
			return conn, idx, err
		}
	}
	return UserConnection[UserConnectionDataYoutube]{}, -1, fmt.Errorf("could not find any youtube connections")
}

func (ucl UserConnectionList) Discord(filter ...string) (UserConnection[UserConnectionDataDiscord], int, error) {
	for idx, v := range ucl {
		if len(filter) > 0 && !utils.Contains(filter, v.ID) {
			continue // does not pass filter
		}

		if v.Platform == UserConnectionPlatformDiscord {
			conn, err := ConvertUserConnection[UserConnectionDataDiscord](v)
			return conn, idx, err
		}
	}
	return UserConnection[UserConnectionDataDiscord]{}, -1, fmt.Errorf("could not find any discord connections")
}

func (ucl UserConnectionList) Get(id string) (UserConnection[bson.Raw], int) {
	for idx, v := range ucl {
		if v.ID == id {
			return v, idx
		}
	}

	return UserConnection[bson.Raw]{}, -1
}

// UserConnectionPlatform Represents a platform that the app supports
type UserConnectionPlatform string

var (
	UserConnectionPlatformTwitch  UserConnectionPlatform = "TWITCH"
	UserConnectionPlatformYouTube UserConnectionPlatform = "YOUTUBE"
	UserConnectionPlatformDiscord UserConnectionPlatform = "DISCORD"
)

type UserType string

var (
	UserTypeRegular UserType = ""
	UserTypeBot     UserType = "BOT"
	UserTypeSystem  UserType = "SYSTEM"
)

type UserConnectionData interface {
	bson.Raw | UserConnectionDataTwitch | UserConnectionDataYoutube | UserConnectionDataDiscord
}

// UserConnection: Represents an external connection to a platform for a user
type UserConnection[D UserConnectionData] struct {
	ID string `json:"id,omitempty" bson:"id,omitempty"`
	// the platform of this connection
	Platform UserConnectionPlatform `json:"platform" bson:"platform"`
	// the time at which this connection was linked
	LinkedAt time.Time `json:"linked_at" bson:"linked_at"`
	// the maximum amount of emotes this connection may have have enabled, counting the total from active sets
	EmoteSlots int32 `json:"emote_slots" bson:"emote_slots"`
	// emote sets bound to this connection / channel
	EmoteSetID ObjectID `json:"emote_set_id,omitempty" bson:"emote_set_id,omitempty"`
	// third-party connection data
	Data D `json:"data" bson:"data"`
	// a list of different possible connection data objects
	// the user must choose one to confirm the connection
	ChoiceData []bson.Raw `json:"choice_data,omitempty" bson:"choice_data,omitempty"`
	// a full oauth2 token grant
	Grant *UserConnectionGrant `json:"-" bson:"grant,omitempty"`

	// Relational

	EmoteSet *EmoteSet `json:"emote_set" bson:"emote_set,skip,omitempty"`
}

func (u UserConnection[D]) ToRaw() UserConnection[bson.Raw] {
	switch x := utils.ToAny(u.Data).(type) {
	case bson.Raw:
		return UserConnection[bson.Raw]{
			ID:         u.ID,
			Platform:   u.Platform,
			LinkedAt:   u.LinkedAt,
			EmoteSlots: u.EmoteSlots,
			EmoteSetID: u.EmoteSetID,
			Data:       x,
			ChoiceData: u.ChoiceData,
			Grant:      u.Grant,
			EmoteSet:   u.EmoteSet,
		}
	}

	raw, _ := bson.Marshal(u.Data)
	return UserConnection[bson.Raw]{
		ID:         u.ID,
		Platform:   u.Platform,
		LinkedAt:   u.LinkedAt,
		EmoteSlots: u.EmoteSlots,
		EmoteSetID: u.EmoteSetID,
		Data:       raw,
		ChoiceData: u.ChoiceData,
		Grant:      u.Grant,
		EmoteSet:   u.EmoteSet,
	}
}

func ConvertUserConnection[D UserConnectionData](c UserConnection[bson.Raw]) (UserConnection[D], error) {
	var d D
	err := bson.Unmarshal(c.Data, &d)
	c2 := UserConnection[D]{
		ID:         c.ID,
		Platform:   c.Platform,
		LinkedAt:   c.LinkedAt,
		EmoteSlots: c.EmoteSlots,
		EmoteSetID: c.EmoteSetID,
		Data:       d,
		Grant:      c.Grant,
		EmoteSet:   c.EmoteSet,
	}

	return c2, err
}

type UserConnectionGrant struct {
	AccessToken  string    `json:"access_token" bson:"access_token"`
	RefreshToken string    `json:"refresh_token" bson:"refresh_token"`
	Scope        []string  `json:"scope" bson:"scope"`
	ExpiresAt    time.Time `json:"expires_at" bson:"expires_at"`
}

// UserConnectionBuilder: utility for creating a new UserConnection
type UserConnectionBuilder[D UserConnectionData] struct {
	Update         UpdateMap
	UserConnection UserConnection[D]
}

// NewUserConnectionBuilder: create a new user connection builder
func NewUserConnectionBuilder[D UserConnectionData](v UserConnection[D]) *UserConnectionBuilder[D] {
	return &UserConnectionBuilder[D]{
		Update:         UpdateMap{},
		UserConnection: v,
	}
}

func (ucb *UserConnectionBuilder[D]) SetID(id string) *UserConnectionBuilder[D] {
	ucb.UserConnection.ID = id
	ucb.Update.Set("connections.$.id", id)
	return ucb
}

// SetPlatform: defines the platform a connection is for (i.e twitch/youtube)
func (ucb *UserConnectionBuilder[D]) SetPlatform(platform UserConnectionPlatform) *UserConnectionBuilder[D] {
	ucb.UserConnection.Platform = platform
	ucb.Update.Set("connections.$.platform", platform)
	return ucb
}

// SetLinkedAt: set the time at which the connection was linked
func (ucb *UserConnectionBuilder[D]) SetLinkedAt(date time.Time) *UserConnectionBuilder[D] {
	ucb.UserConnection.LinkedAt = date
	ucb.Update.Set("connections.$.linked_at", date)
	return ucb
}

func (ucb *UserConnectionBuilder[D]) SetActiveEmoteSet(id ObjectID) *UserConnectionBuilder[D] {
	ucb.UserConnection.EmoteSetID = id
	ucb.Update.Set("connections.$.emote_set_id", id)
	return ucb
}

func (ucb *UserConnectionBuilder[D]) SetData(data D) *UserConnectionBuilder[D] {
	ucb.UserConnection.Data = data
	return ucb
}

func (ucb *UserConnectionBuilder[D]) SetGrant(at string, rt string, ex int, sc []string) *UserConnectionBuilder[D] {
	g := &UserConnectionGrant{
		AccessToken:  at,
		RefreshToken: rt,
		Scope:        sc,
		ExpiresAt:    time.Now().Add(time.Second * time.Duration(ex)),
	}

	ucb.UserConnection.Grant = g
	ucb.Update.Set("connections.$.grant", g)
	return ucb
}

type UserConnectionDataTwitch struct {
	ID              string    `json:"id" bson:"id"`
	Login           string    `json:"login" bson:"login"`
	DisplayName     string    `json:"display_name" bson:"display_name"`
	BroadcasterType string    `json:"broadcaster_type" bson:"broadcaster_type"`
	Description     string    `json:"description" bson:"description"`
	ProfileImageURL string    `json:"profile_image_url" bson:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url" bson:"offline_image_url"`
	ViewCount       int       `json:"view_count" bson:"view_count"`
	Email           string    `json:"email" bson:"email"`
	CreatedAt       time.Time `json:"created_at" bson:"twitch_created_at"`
}

type UserConnectionDataYoutube struct {
	ID              string `json:"id" bson:"id"`
	Title           string `json:"title" bson:"title"`
	Description     string `json:"description" bson:"description"`
	ViewCount       int64  `json:"view_count" bson:"view_count"`
	SubCount        int64  `json:"sub_count" bson:"sub_count"`
	ProfileImageURL string `json:"profile_image_url" bson:"profile_image_url"`
}

type UserConnectionDataDiscord struct {
	ID            string `json:"id" bson:"id"`
	Username      string `json:"username" bson:"username"`
	Discriminator string `json:"discriminator" bson:"discriminator"`
	Avatar        string `json:"avatar" bson:"avatar"`
	Bot           bool   `json:"bot" bson:"bot"`
	System        bool   `json:"system" bson:"system"`
	MFAEnabled    bool   `json:"mfa_enabled" bson:"mfa_enabled"`
	Banner        string `json:"banner" bson:"banner,omitempty"`
	AccentColor   int64  `json:"accent_color" bson:"accent_color,omitempty"`
	Locale        string `json:"locale" bson:"locale,omitempty"`
	Verified      bool   `json:"verified" bson:"verified"`
	Email         string `json:"email" bson:"email,omitempty"`
	Flags         int64  `json:"flags" bson:"flags"`
	PremiumType   uint32 `json:"premium_type" bson:"premium_type"`
	PublicFlags   uint32 `json:"public_flags" bson:"public_flags"`
}

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
