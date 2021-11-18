package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmoteBuilder: Wraps an Emote and offers methods to fetch and mutate emote data
type EmoteBuilder struct {
	Update UpdateMap
	Emote  *Emote
}

// NewEmoteBuilder: create a new emote builder
func NewEmoteBuilder(emote *Emote) *EmoteBuilder {
	return &EmoteBuilder{
		Update: UpdateMap{},
		Emote:  emote,
	}
}

// SetName: change the name of the emote
func (eb *EmoteBuilder) SetName(name string) *EmoteBuilder {
	eb.Emote.Name = name
	eb.Update.Set("flags", eb.Emote.Flags)
	return eb
}

func (eb *EmoteBuilder) SetOwnerID(id primitive.ObjectID) *EmoteBuilder {
	eb.Emote.OwnerID = id
	eb.Update.Set("owner_id", id)
	return eb
}

// SetPrivacy: change the private state of the emote
func (eb *EmoteBuilder) SetPrivacy(isPrivate bool) *EmoteBuilder {
	if isPrivate {
		eb.Emote.Flags |= EmoteFlagsPrivate
	} else {
		eb.Emote.Flags &= EmoteFlagsPrivate
	}

	eb.Update.Set("flags", eb.Emote.Flags)
	return eb
}

// SetListed: change the listing state of the emote
func (eb *EmoteBuilder) SetListed(isListed bool) *EmoteBuilder {
	if isListed {
		eb.Emote.Flags |= EmoteFlagsListed
	} else {
		eb.Emote.Flags &= EmoteFlagsListed
	}

	eb.Update.Set("flags", eb.Emote.Flags)
	return eb
}

// SetStatus: change the emote's status
func (eb *EmoteBuilder) SetStatus(status EmoteStatus) *EmoteBuilder {
	eb.Emote.Status = status
	eb.Update.Set("status", status)
	return eb
}

// AddSize: add a size reference to the emote
func (eb *EmoteBuilder) AddSize(name string, width int32, height int32, byteSize int) *EmoteBuilder {
	size := EmoteSize{
		Name:     name,
		Width:    width,
		Height:   height,
		ByteSize: byteSize,
	}
	eb.Emote.Sizes = append(eb.Emote.Sizes, size)
	eb.Update.AddToSet("sizes", size)
	return eb
}

type Emote struct {
	ID      ObjectID    `json:"id" bson:"_id"`
	OwnerID ObjectID    `json:"owner_id" bson:"owner_id"`
	Name    string      `json:"name" bson:"name"`
	Flags   EmoteFlag   `json:"visibility" bson:"visibility"` // DEPRECATED: no longer used in v3
	Status  EmoteStatus `json:"status" bson:"status"`
	Tags    []string    `json:"tags" bson:"tags"`

	// Meta

	Sizes    []EmoteSize   `json:"sizes" bson:"sizes"`
	Animated bool          `json:"animated" bson:"animated"`                   // Whether or not the emote is animated
	Formats  []EmoteFormat `json:"formats,omitempty" bson:"formats,omitempty"` // Whether or not the emote is available in AVIF (AV1 Image File) Format

	// Moderation Data
	Moderation *EmoteModeration `json:"moderation,omitempty" bson:"moderation,omitempty"`

	// Versioning

	ParentID   *primitive.ObjectID `json:"parent_id,omitempty" bson:"parent_id"`
	Versioning *EmoteVersioning    `json:"version,omitempty" bson:"version,omitempty"`

	// Non-structural

	Links [][]string `json:"urls" bson:"-"` // CDN URLs

	// Relational

	Owner    *User `json:"owner" bson:"owner_user,skip"`
	Channels []*User
}

type EmoteStatus int32

const (
	EmoteStatusDeleted EmoteStatus = iota - 1
	EmoteStatusProcessing
	EmoteStatusPending
	EmoteStatusDisabled
	EmoteStatusLive
)

type EmoteFlag int32

const (
	EmoteFlagsPrivate   EmoteFlag = 1 << 0
	EmoteFlagsListed    EmoteFlag = 1 << 1
	EmoteFlagsZeroWidth EmoteFlag = 1 << 8

	EmoteFlagsAll int32 = (1 << iota) - 1
)

type EmoteSize struct {
	Name     string `json:"n" bson:"name"`   // The name of the size
	Width    int32  `json:"w" bson:"width"`  // The pixel width of the emote
	Height   int32  `json:"h" bson:"height"` // The pixel height of the emote
	ByteSize int    `json:"b" bson:"b"`

	Link string `json:"l" bson:"-"` // The CDN URL to the emote
}

type EmoteFormat string

const (
	EmoteFormatWEBP EmoteFormat = "WEBP"
	EmoteFormatGIF  EmoteFormat = "GIF"
	EmoteFormatAVIF EmoteFormat = "AVIF"
	EmoteFormatPNG  EmoteFormat = "PNG"
)

type EmoteModeration struct {
	// The reason given by a moderator for the emote not being allowed in public listing
	RejectionReason string `json:"reject_reason,omitempty" bson:"reject_reason,omitempty"`
}

type EmoteVersioning struct {
	// The displayed label for the version
	Tag string `json:"tag" bson:"tag"`
	// Whether or not this version is diverging (i.e a holiday variant)
	// If true, this emote will never be prompted as an update
	Diverged bool `json:"diverged" bson:"diverged"`
	// The time at which the emote became a version
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}
