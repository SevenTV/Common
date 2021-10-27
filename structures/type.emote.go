package structures

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmoteBuilder: Wraps an Emote and offers methods to fetch and mutate emote data
type EmoteBuilder struct {
	Emote *Emote
}

// FetchByID: Get an emote by its ID
func (b EmoteBuilder) FetchByID(ctx context.Context, inst mongo.Instance, id primitive.ObjectID) (*Emote, error) {
	doc := inst.Collection(mongo.CollectionNameEmotes).FindOne(ctx, bson.M{
		"_id": id,
	})
	if err := doc.Err(); err != nil {
		return nil, err
	}

	var emote Emote
	if err := doc.Decode(&emote); err != nil {
		return nil, err
	}

	return &emote, nil
}

type Emote struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	OwnerID    primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	Name       string             `json:"name" bson:"name"`
	Visibility int32              `json:"visibility" bson:"visibility"`
	Status     int32              `json:"status" bson:"status"`
	Tags       []string           `json:"tags" bson:"tags"`

	// Meta
	Width    []int32 `json:"width" bson:"width"`       // The pixel width of the emote
	Height   []int32 `json:"height" bson:"height"`     // The pixel height of the emote
	Animated bool    `json:"animated" bson:"animated"` // Whether or not the emote is animated
	AVIF     bool    `json:"avif" bson:"avif"`         // Whether or not the emote is available in AVIF (AV1 Image File) Format
	ByteSize int32   `json:"byte_size,omitempty" bson:"byte_size,omitempty"`

	// Non-structural

	Links [][]string `json:"urls" bson:"-"` // CDN URLs

	// Relational

	Owner    *User `json:"owner" bson:"owner_user,skip"`
	Channels []*User
}

const (
	EmoteStatusDeleted int32 = iota - 1
	EmoteStatusProcessing
	EmoteStatusPending
	EmoteStatusDisabled
	EmoteStatusLive
)

const (
	EmoteVisibilityPrivate int32 = 1 << iota
	EmoteVisibilityGlobal
	EmoteVisibilityUnlisted
	EmoteVisibilityOverrideBTTV
	EmoteVisibilityOverrideFFZ
	EmoteVisibilityOverrideTwitchGlobal
	EmoteVisibilityOverrideTwitchSubscriber
	EmoteVisibilityZeroWidth

	EmoteVisibilityAll int32 = (1 << iota) - 1
)
