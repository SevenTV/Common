package structures

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserBuilder struct {
	Update UpdateMap
	User   *User
}

// NewUserBuilder: create a new user builder
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		Update: UpdateMap{},
		User: &User{
			ChannelEmotes: []UserEmote{},
			Editors:       []UserEditor{},
			Connections:   []primitive.ObjectID{},
		},
	}
}

// SetUsername: set the username for the user
func (ub *UserBuilder) SetUsername(username string) *UserBuilder {
	ub.User.Username = username
	ub.Update.Set("username", username)

	return ub
}

// SetEmail: set the email for the user
func (ub *UserBuilder) SetEmail(email string) *UserBuilder {
	ub.User.Email = email
	ub.Update.Set("email", email)

	return ub
}

func (ub *UserBuilder) AddConnection(id primitive.ObjectID) *UserBuilder {
	ub.User.Connections = append(ub.User.Connections, id)
	ub.Update = ub.Update.AddToSet("connections", &id)

	return ub
}

// User: A standard app user object
type User struct {
	ID            primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Username      string               `json:"username" bson:"username"`
	Email         string               `json:"email" bson:"email"`
	ChannelEmotes []UserEmote          `json:"channel_emotes" bson:"channel_emotes"`
	Editors       []UserEditor         `json:"editors" bson:"editors"`
	AvatarURL     string               `json:"avatar_url" bson:"avatar_url"`
	Biography     string               `json:"biography" bson:"biography"`
	TokenVersion  float32              `json:"token_version" bson:"token_version"`
	Connections   []primitive.ObjectID `json:"connections" bson:"connections"`
}

// UserConnectionPlatform: Represents a platform that the app supports
type UserConnectionPlatform string

var (
	UserConnectionPlatformTwitch  UserConnectionPlatform = "TWITCH"
	UserConnectionPlatformYouTube UserConnectionPlatform = "YOUTUBE"
)

// UserConnection: Represents an external connection to a platform for a user
type UserConnection struct {
	ID       primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	Platform UserConnectionPlatform `json:"platform" bson:"platform"`
	LinkedAt time.Time              `json:"linked_at" bson:"linked_at"`
	Data     bson.Raw               `json:"data" bson:"data"`
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
	ID string `json:"id" bson:"id"`
}

type UserEmote struct {
	ID primitive.ObjectID `json:"id" bson:"id"`
	// When this has 1 or more items, the emote will only be availablle for these connections (i.e specific twitch/youtube channels)
	Connections []primitive.ObjectID `json:"connections" bson:"connections"`
	// An alias for this emote
	Alias string `json:"alias,omitempty" bson:"alias,omitempty"`
	// Whether or not the emote will be made zero width for the particular channel
	ZeroWidth bool `json:"zero_width,omitempty" bson:"zero_width,omitempty"`
}

type UserEditor struct {
	ID primitive.ObjectID `json:"id" bson:"id"`
	// When this has 1 or more items, this editor will only have access to these connections (i.e specific twitch/youtube channels)
	Connections []primitive.ObjectID `json:"connections" bson:"connections"`
	// The permissions this editor has
	Permissions UserEditorPermission `json:"permissions" bson:"permissions"`
	// Whether or not that editor will be visible on the user's profile page
	Visible bool `json:"visible" bson:"visible"`
}

type UserEditorPermission int32

const (
	UserEditorPermissionAddChannelEmotes    UserEditorPermission = 1 << iota // 1 - Allows adding emotes
	UserEditorPermissionRemoveChannelEmotes                                  // 2 - Allows removing emotes
	UserEditorPermissionManageProfile                                        // 4 - Allows managing the user's public profile
	UserEditorPermissionManageBilling                                        // 8 - Allows managing billing and payments, such as subscriptions
	UserEditorPermissionManageOwnedEmotes                                    // 16 - Allows managing the user's owned emotes
	UserEditorPermissionMUsePrivateEmotes                                    // 32 - Allows using the user's private emotes
)
