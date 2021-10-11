package structures

import (
	"context"
	"fmt"
	"time"

	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserBuilder struct {
	User *User
}

// FetchByID: Get a user by their ID
func (b UserBuilder) FetchByID(ctx context.Context, id primitive.ObjectID) (*UserBuilder, error) {
	doc := mongo.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
		"_id": id,
	})
	if err := doc.Err(); err != nil {
		return nil, err
	}

	var user *User
	if err := doc.Decode(&user); err != nil {
		return nil, err
	}

	b.User = user
	return &b, nil
}

// User: A standard app user object
type User struct {
	ID            primitive.ObjectID   `json:"id" bson:"_id"`
	Username      string               `json:"username" bson:"username"`
	Email         string               `json:"email" bson:"email"`
	ChannelEmotes []UserEmote          `json:"channel_emotes" bson:"channel_emotes"`
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
	Platform UserConnectionPlatform `json:"platform" bson:"platform"`
	LinkedAt time.Time              `json:"linked_at" bson:"linked_at"`
	Data     bson.Raw               `json:"data" bson:"data"`
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
	ID    string `json:"id" bson:"id"`
	Login string `json:"login" bson:"login"`
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
