package structures

import (
	"fmt"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var emoteTagRegex = regexp.MustCompile(`^[0-9a-z]{3,30}$`)

type Emote struct {
	ID      ObjectID  `json:"id" bson:"_id"`
	OwnerID ObjectID  `json:"owner_id" bson:"owner_id"`
	Name    string    `json:"name" bson:"name"`
	Flags   EmoteFlag `json:"flags" bson:"flags"`
	Tags    []string  `json:"tags" bson:"tags"`

	// Versioning

	Versions    []*EmoteVersion      `json:"versions,omitempty" bson:"versions,omitempty"`
	ChildrenIDs []primitive.ObjectID `json:"children_ids,omitempty" bson:"children_ids,omitempty"`
	ParentID    *primitive.ObjectID  `json:"parent_id,omitempty" bson:"parent_id,omitempty"`

	// Relational

	Owner    *User   `json:"owner" bson:"owner_user,skip,omitempty"`
	Channels []*User `json:"channels" bson:"channels,skip,omitempty"`
}

type EmoteLifecycle int32

const (
	EmoteLifecycleDeleted EmoteLifecycle = iota - 1
	EmoteLifecyclePending
	EmoteLifecycleProcessing
	EmoteLifecycleDisabled
	EmoteLifecycleLive
	EmoteLifecycleFailed EmoteLifecycle = -2
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
	Files []EmoteFile     `json:"files" bson:"files"`
}

type EmoteFile struct {
	Name           string `json:"n" bson:"name"`     // The name of the file
	Width          int32  `json:"w" bson:"width"`    // The pixel width of the emote
	Height         int32  `json:"h" bson:"height"`   // The pixel height of the emote
	Animated       bool   `json:"a" bson:"animated"` // Whether or not this file is animated
	ProcessingTime int64  `json:"-" bson:"time"`     // The amount of time in nanoseconds it took for this file to be processed
	Length         int64  `json:"b" bson:"length"`   // The file size in bytes
}

type EmoteFormatName string

const (
	EmoteFormatNameWEBP EmoteFormatName = "image/webp"
	EmoteFormatNameAVIF EmoteFormatName = "image/avif"
	EmoteFormatNameGIF  EmoteFormatName = "image/gif"
	EmoteFormatNamePNG  EmoteFormatName = "image/png"
)

type EmoteState struct {
	// The current life cycle of the emote
	// indicating whether it's processing, live, deleted, etc.
	Lifecycle EmoteLifecycle `json:"lifecycle" bson:"lifecycle"`
	// The ranked position for the amount of channels this emote is added on to.
	// This value is to be determined with an external cron job
	ChannelCountRank int32 `json:"-" bson:"channel_count_rank,omitempty"`
	// The amount of channels this emote is added on to.
	// This value is to be determined with an external cron job
	ChannelCount int32 `json:"-" bson:"channel_count,omitempty"`
	// The time at which the ChannelCount value was last checked
	ChannelCountCheckAt time.Time `json:"-" bson:"channel_count_check_at,omitempty"`
}

type EmoteVersion struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Timestamp   time.Time          `json:"timestamp" bson:"timestamp"`
	State       EmoteState         `json:"state" bson:"state"`
	FrameCount  int32              `json:"frame_count" bson:"frame_count"`
	Formats     []EmoteFormat      `json:"formats,omitempty" bson:"formats,omitempty"`
}

// EmoteBuilder Wraps an Emote and offers methods to fetch and mutate emote data
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

func (eb *EmoteBuilder) GetVersion(id ObjectID) *EmoteVersion {
	for _, v := range eb.Emote.Versions {
		if v.ID == id {
			return v
		}
	}
	return nil
}

func (eb *EmoteBuilder) AddVersion(v *EmoteVersion) *EmoteBuilder {
	for _, vv := range eb.Emote.Versions {
		if vv.ID == v.ID {
			return eb
		}
	}

	eb.Emote.Versions = append(eb.Emote.Versions, v)
	eb.Update.AddToSet("versions", v)
	return eb
}

func (eb *EmoteBuilder) UpdateVersion(id ObjectID, v *EmoteVersion) *EmoteBuilder {
	ind := -1
	for i, vv := range eb.Emote.Versions {
		if vv.ID == v.ID {
			ind = i
			break
		}
	}

	eb.Emote.Versions[ind] = v
	eb.Update.Set(fmt.Sprintf("versions.%d", ind), v)
	return eb
}

func (eb *EmoteBuilder) RemoveVersion(id ObjectID) *EmoteBuilder {
	ind := -1
	for i := range eb.Emote.Versions {
		if eb.Emote.Versions[i] == nil {
			continue
		}
		if eb.Emote.Versions[i].ID != id {
			continue
		}
		ind = i
		break
	}
	if ind == -1 {
		return eb
	}

	copy(eb.Emote.Versions[ind:], eb.Emote.Versions[ind+1:])
	eb.Emote.Versions[len(eb.Emote.Versions)-1] = nil
	eb.Emote.Versions = eb.Emote.Versions[:len(eb.Emote.Versions)-1]
	eb.Update.Pull("versions", bson.M{"id": id})
	return eb
}
