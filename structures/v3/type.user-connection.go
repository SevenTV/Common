package structures

import (
	"fmt"
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
)

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

func (uc UserConnection[D]) Username() (string, string) {
	var displayName string
	var username string

	switch uc.Platform {
	case UserConnectionPlatformTwitch:
		if con, err := ConvertUserConnection[UserConnectionDataTwitch](uc.ToRaw()); err == nil {
			displayName = con.Data.DisplayName
			username = con.Data.Login
		}
	case UserConnectionPlatformYouTube:
		if con, err := ConvertUserConnection[UserConnectionDataYoutube](uc.ToRaw()); err == nil {
			displayName = con.Data.Title
			username = con.Data.ID
		}
	case UserConnectionPlatformDiscord:
		if con, err := ConvertUserConnection[UserConnectionDataDiscord](uc.ToRaw()); err == nil {
			displayName = con.Data.Username
			username = con.Data.Username + "#" + con.Data.Discriminator
		}
	}
	return username, displayName
}
