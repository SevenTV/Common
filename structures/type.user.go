package structures

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/SevenTV/Common/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserBuilder struct {
	Update UpdateMap
	User   *User
}

// NewUserBuilder: create a new user builder
func NewUserBuilder(user *User) *UserBuilder {
	return &UserBuilder{
		Update: UpdateMap{},
		User:   user,
	}
}

// SetUsername: set the username for the user
func (ub *UserBuilder) SetUsername(username string) *UserBuilder {
	ub.User.Username = username
	ub.Update.Set("username", username)

	return ub
}

func (ub *UserBuilder) SetDiscriminator(discrim string) *UserBuilder {
	if discrim == "" {
		for i := 0; i < 4; i++ {
			discrim += strconv.Itoa(rand.Intn(9))
		}
	}

	ub.User.Discriminator = discrim
	ub.Update.Set("discriminator", discrim)
	return ub
}

// SetEmail: set the email for the user
func (ub *UserBuilder) SetEmail(email string) *UserBuilder {
	ub.User.Email = email
	ub.Update.Set("email", email)

	return ub
}

func (ub *UserBuilder) SetAvatarURL(url string) *UserBuilder {
	ub.User.AvatarURL = url
	ub.Update.Set("avatar_url", url)

	return ub
}

func (ub *UserBuilder) AddConnection(id primitive.ObjectID) *UserBuilder {
	ub.User.ConnectionIDs = append(ub.User.ConnectionIDs, id)
	ub.Update = ub.Update.AddToSet("connection_ids", &id)

	return ub
}

// User: A standard app user object
type User struct {
	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
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
	// the user's channel-bound emotes
	ChannelEmotes []*UserEmote `json:"channel_emotes" bson:"channel_emotes"`
	// list of role IDs directly bound to the user (not via an entitlement)
	RoleIDs []primitive.ObjectID `json:"role_ids" bson:"role_ids"`
	// the user's editors
	Editors []*UserEditor `json:"editors" bson:"editors"`
	// the user's avatar URL
	AvatarURL string `json:"avatar_url" bson:"avatar_url"`
	// the user's biography
	Biography string `json:"biography" bson:"biography"`
	// token version
	TokenVersion float64 `json:"token_version" bson:"token_version"`
	// third party connections. it's like third parties for a third party.
	ConnectionIDs []primitive.ObjectID `json:"connection_ids" bson:"connection_ids"`

	// Relational

	Roles       []*Role           `json:"roles" bson:"roles,skip,omitempty"`
	Connections []*UserConnection `json:"connections" bson:"connections,skip,omitempty"`
	OwnedEmotes []*Emote          `json:"owned_emotes" bson:"owned_emotes,skip,omitempty"`
	EditorOf    []*UserEditor     `json:"editor_of" bson:"editor_of,skip,omitempty"`
	Bans        []*Ban            `json:"bans" bson:"bans,skip,omitempty"`
}

// HasPermission: checks relational roles against a permission bit
func (u *User) HasPermission(bit RolePermission) bool {
	total := RolePermission(0)
	for _, r := range u.Roles {
		a := r.Allowed

		total |= a
	}
	for _, r := range u.Roles {
		d := r.Denied

		total &= ^d
	}

	if (total & RolePermissionSuperAdministrator) == RolePermissionSuperAdministrator {
		return true
	}
	return utils.BitField.HasBits(int64(total), int64(bit))
}

func (u *User) AddRoles(roles ...*Role) {
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

func (u *User) GetHighestRole() *Role {
	u.SortRoles()
	if len(u.Roles) == 0 {
		return NilRole
	}

	return u.Roles[0]
}

type UserDiscriminator uint8

// UserConnectionPlatform: Represents a platform that the app supports
type UserConnectionPlatform string

var (
	UserConnectionPlatformTwitch  UserConnectionPlatform = "TWITCH"
	UserConnectionPlatformYouTube UserConnectionPlatform = "YOUTUBE"
)

type UserType string

var (
	UserTypeRegular UserType = ""
	UserTypeBot     UserType = "BOT"
	UserTypeSystem  UserType = "SYSTEM"
)

// UserConnection: Represents an external connection to a platform for a user
type UserConnection struct {
	ID       primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	Platform UserConnectionPlatform `json:"platform" bson:"platform"`
	LinkedAt time.Time              `json:"linked_at" bson:"linked_at"`
	Data     bson.Raw               `json:"data" bson:"data"`
	Grant    *UserConnectionGrant   `json:"-" bson:"grant"`
}

type UserConnectionGrant struct {
	AccessToken  string    `json:"access_token" bson:"access_token"`
	RefreshToken string    `json:"refresh_token" bson:"refresh_token"`
	Scope        []string  `json:"scope" bson:"scope"`
	ExpiresAt    time.Time `json:"expires_at" bson:"expires_at"`
}

// UserConnectionBuilder: utility for creating a new UserConnection
type UserConnectionBuilder struct {
	Update         UpdateMap
	UserConnection *UserConnection
}

// NewUserConnectionBuilder: create a new user connection builder
func NewUserConnectionBuilder() *UserConnectionBuilder {
	return &UserConnectionBuilder{
		Update:         UpdateMap{},
		UserConnection: &UserConnection{},
	}
}

// SetPlatform: defines the platform a connection is for (i.e twitch/youtube)
func (ucb *UserConnectionBuilder) SetPlatform(platform UserConnectionPlatform) *UserConnectionBuilder {
	ucb.UserConnection.Platform = platform
	ucb.Update.Set("platform", platform)

	return ucb
}

// SetLinkedAt: set the time at which the connection was linked
func (ucb *UserConnectionBuilder) SetLinkedAt(date time.Time) *UserConnectionBuilder {
	ucb.UserConnection.LinkedAt = date
	ucb.Update.Set("linked_at", date)

	return ucb
}

// SetTwitchData: set the data for a twitch connection
func (ucb *UserConnectionBuilder) SetTwitchData(data *TwitchConnection) *UserConnectionBuilder {
	return ucb.setPlatformData(data)
}

// SetYouTubeData: set the data for a youtube connection
func (ucb *UserConnectionBuilder) SetYouTubeData(data *YouTubeConnection) *UserConnectionBuilder {
	return ucb.setPlatformData(data)
}

func (ucb *UserConnectionBuilder) setPlatformData(v interface{}) *UserConnectionBuilder {
	b, err := bson.Marshal(v)
	if err != nil {
		logrus.WithError(err).Error("bson")
		return ucb
	}

	ucb.UserConnection.Data = b
	ucb.Update.Set("data", v)
	return ucb
}

func (ucb *UserConnectionBuilder) SetGrant(at string, rt string, ex int, sc []string) *UserConnectionBuilder {
	g := &UserConnectionGrant{
		AccessToken:  at,
		RefreshToken: rt,
		Scope:        sc,
		ExpiresAt:    time.Now().Add(time.Second * time.Duration(ex)),
	}

	ucb.UserConnection.Grant = g
	ucb.Update.Set("grant", g)
	return ucb
}

// DecodeTwitch: get the data of a twitch user connnection
func (uc *UserConnection) DecodeTwitch() (*TwitchConnection, error) {
	if uc.Platform != UserConnectionPlatformTwitch {
		return nil, fmt.Errorf("wrong platform %s for DecodeTwitch", uc.Platform)
	}

	var c *TwitchConnection
	if err := bson.Unmarshal(uc.Data, &c); err != nil {
		return nil, err
	}

	return c, nil
}

// DecodeYouTube: get the data of a youtube user connection
func (uc *UserConnection) DecodeYouTube() (*YouTubeConnection, error) {
	if uc.Platform != UserConnectionPlatformYouTube {
		return nil, fmt.Errorf("wrong platform %s for DecodeYouTube", uc.Platform)
	}

	var c *YouTubeConnection
	if err := bson.Unmarshal(uc.Data, &c); err != nil {
		return nil, err
	}

	return c, nil
}

type TwitchConnection struct {
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

type YouTubeConnection struct {
	ID          string `json:"id" bson:"id"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
}

type UserEmote struct {
	ID primitive.ObjectID `json:"id" bson:"id"`
	// When this has 1 or more items, the emote will only be availablle for these connections (i.e specific twitch/youtube channels)
	Connections []primitive.ObjectID `json:"connections" bson:"connections,omitempty"`
	// An alias for this emote
	Alias string `json:"alias,omitempty" bson:"alias,omitempty"`
	// Whether or not the emote will be made zero width for the particular channel
	ZeroWidth bool `json:"zero_width,omitempty" bson:"zero_width,omitempty"`

	// Relational
	Emote *Emote `json:"emote" bson:"emote,omitempty,skip"`
}

type UserEditor struct {
	ID primitive.ObjectID `json:"id" bson:"id"`
	// When this has 1 or more items, this editor will only have access to these connections (i.e specific twitch/youtube channels)
	Connections []primitive.ObjectID `json:"connections" bson:"connections"`
	// The permissions this editor has
	Permissions UserEditorPermission `json:"permissions" bson:"permissions"`
	// Whether or not that editor will be visible on the user's profile page
	Visible bool `json:"visible" bson:"visible"`

	// Relational
	User *User `json:"user" bson:"user,skip,omitempty"`
}

// HasPermission: check whether or not the editor has a permission
func (ed *UserEditor) HasPermission(bit UserEditorPermission) bool {
	return utils.BitField.HasBits(int64(ed.Permissions), int64(bit))
}

type UserEditorPermission int32

const (
	UserEditorPermissionAddChannelEmotes    UserEditorPermission = 1 << iota // 1 - Allows adding emotes
	UserEditorPermissionRemoveChannelEmotes                                  // 2 - Allows removing emotes
	UserEditorPermissionMUsePrivateEmotes                                    // 4 - Allows using the user's private emotes
	UserEditorPermissionManageProfile                                        // 8 - Allows managing the user's public profile
	UserEditorPermissionManageBilling                                        // 16 - Allows managing billing and payments, such as subscriptions
	UserEditorPermissionManageOwnedEmotes                                    // 32 - Allows managing the user's owned emotes
	UserEditorPermissionManageEmoteSets                                      // 64 - Allows managing the user's owned emote sets
)
