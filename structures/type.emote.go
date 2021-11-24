package structures

import (
	"regexp"
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
	eb.Update.Set("name", eb.Emote.Name)
	return eb
}

func (eb *EmoteBuilder) SetOwnerID(id primitive.ObjectID) *EmoteBuilder {
	eb.Emote.OwnerID = id
	eb.Update.Set("owner_id", id)
	return eb
}

func (eb *EmoteBuilder) SetFlags(sum EmoteFlag) *EmoteBuilder {
	eb.Emote.Flags = sum
	eb.Update.Set("flags", sum)
	return eb
}

func (eb *EmoteBuilder) SetTags(tags []string, validate bool) *EmoteBuilder {
	uniqueTags := map[string]bool{}
	for _, v := range tags {
		if v == "" {
			continue
		}
		if !emoteTagRegex.MatchString(v) {
			continue
		}
		uniqueTags[v] = true
	}

	tags = make([]string, len(uniqueTags))
	i := 0
	for k := range uniqueTags {
		tags[i] = k
		i++
	}

	eb.Emote.Tags = tags
	eb.Update.Set("tags", tags)
	return eb
}

// SetStatus: change the emote's status
func (eb *EmoteBuilder) SetStatus(status EmoteStatus) *EmoteBuilder {
	eb.Emote.Status = status
	eb.Update.Set("status", status)
	return eb
}

var emoteTagRegex = regexp.MustCompile(`^[0-9a-z]{3,30}$`)

type Emote struct {
	ID      ObjectID  `json:"id" bson:"_id"`
	OwnerID ObjectID  `json:"owner_id" bson:"owner_id"`
	Name    string    `json:"name" bson:"name"`
	Flags   EmoteFlag `json:"flags" bson:"flags"`
	Tags    []string  `json:"tags" bson:"tags"`

	// Meta

	Status     EmoteStatus   `json:"status" bson:"status"`
	FrameCount int32         `json:"frame_count" bson:"frame_count"`             // The amount of frames this image has
	Formats    []EmoteFormat `json:"formats,omitempty" bson:"formats,omitempty"` // All formats the emote is available is, with width/height/length of each responsive size

	// Moderation Data
	Moderation *EmoteModeration `json:"moderation,omitempty" bson:"moderation,omitempty"`

	// Versioning

	ParentID   *primitive.ObjectID `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	Versioning *EmoteVersioning    `json:"version,omitempty" bson:"version,omitempty"`

	// Non-structural

	Links [][]string `json:"urls" bson:"-"` // CDN URLs

	// Relational

	Owner    *User   `json:"owner" bson:"owner_user,skip,omitempty"`
	Channels []*User `json:"channels" bson:"channels,skip,omitempty"`
}

type EmoteStatus int32

const (
	EmoteStatusDeleted EmoteStatus = iota - 1
	EmoteStatusProcessing
	EmoteStatusPending
	EmoteStatusDisabled
	EmoteStatusLive
	EmoteStatusFailed EmoteStatus = -2
)

type EmoteFlag int32

const (
	EmoteFlagsPrivate   EmoteFlag = 1 << 0
	EmoteFlagsListed    EmoteFlag = 1 << 1
	EmoteFlagsZeroWidth EmoteFlag = 1 << 8

	EmoteFlagsAll EmoteFlag = (1 << iota) - 1
)

type EmoteFormat struct {
	Name  EmoteFormatName `json:"name" bson:"name"`
	Sizes []EmoteSize     `json:"sizes" bson:"sizes"`
}

type EmoteSize struct {
	Scale          string `json:"s" bson:"scale"`    // The responsive scale
	Width          int32  `json:"w" bson:"width"`    // The pixel width of the emote
	Height         int32  `json:"h" bson:"height"`   // The pixel height of the emote
	Animated       bool   `json:"a" bson:"animated"` // Whether or not this size is animated
	ProcessingTime int64  `json:"-" bson:"time"`     // The amount of time in nanoseconds it took for this size to be processed
	Length         int    `json:"b" bson:"length"`   // The file size in bytes
}

type EmoteFormatName string

const (
	EmoteFormatNameWEBP EmoteFormatName = "image/webp"
	EmoteFormatNameAVIF EmoteFormatName = "image/avif"
	EmoteFormatNameGIF  EmoteFormatName = "image/gif"
	EmoteFormatNamePNG  EmoteFormatName = "image/png"
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
