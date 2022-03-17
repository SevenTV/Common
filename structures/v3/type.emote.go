package structures

import (
	"regexp"
	"time"

	"github.com/SevenTV/Common/utils"
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
	EmoteFlagsZeroWidth EmoteFlag = 1 << 8

	// Content Flags

	EmoteFlagsContentSexual           EmoteFlag = 1 << 16 // Sexually Suggesive
	EmoteFlagsContentEpilepsy         EmoteFlag = 1 << 17 // Rapid flashing
	EmoteFlagsContentEdgy             EmoteFlag = 1 << 18 // Edgy or distasteful, may be offensive to some users
	EmoteFlagsContentTwitchDisallowed EmoteFlag = 1 << 24 // Not allowed specifically on the Twitch platform

	EmoteFlagsAll EmoteFlag = (1 << iota) - 1
)

func (e EmoteFlag) String() string {
	switch e {
	case EmoteFlagsPrivate:
		return "PRIVATE"
	case EmoteFlagsZeroWidth:
		return "ZERO_WIDTH"
	case EmoteFlagsContentSexual:
		return "SEXUALLY_SUGGESTIVE"
	case EmoteFlagsContentEpilepsy:
		return "EPILEPSY"
	case EmoteFlagsContentEdgy:
		return "EDGY_OR_DISASTEFUL"
	case EmoteFlagsContentTwitchDisallowed:
		return "TWITCH_DISALLOWED"
	}
	return ""
}

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
	// Whether or not the emote is listed
	Listed bool `json:"listed" bson:"listed"`
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

func (e *Emote) HasFlag(flag EmoteFlag) bool {
	return utils.BitField.HasBits(int64(e.Flags), int64(flag))
}

func (e *Emote) GetVersion(id ObjectID) (*EmoteVersion, int) {
	for i, v := range e.Versions {
		if v.ID == id {
			return v, i
		}
	}
	return nil, -1
}

func (e *Emote) GetLatestVersion(onlyListed bool) *EmoteVersion {
	var ver *EmoteVersion
	for _, v := range e.Versions {
		if onlyListed && !v.State.Listed {
			continue
		}
		if v.IsUnavailable() {
			continue
		}
		if ver == nil || ver.Timestamp.Before(v.Timestamp) {
			ver = v
		}
	}
	return ver
}

func (ev *EmoteVersion) IsUnavailable() bool {
	return ev.State.Lifecycle == EmoteLifecycleDeleted || ev.State.Lifecycle == EmoteLifecycleDisabled || ev.State.Lifecycle == EmoteLifecycleFailed
}

func (ev *EmoteVersion) IsProcessing() bool {
	return ev.State.Lifecycle == EmoteLifecyclePending || ev.State.Lifecycle == EmoteLifecycleProcessing
}
