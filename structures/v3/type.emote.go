package structures

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var emoteTagRegex = regexp.MustCompile(`^[0-9a-z]{3,30}$`)

type Emote struct {
	ID      ObjectID   `json:"id" bson:"_id"`
	OwnerID ObjectID   `json:"owner_id" bson:"owner_id"`
	Name    string     `json:"name" bson:"name"`
	Flags   EmoteFlag  `json:"flags" bson:"flags"`
	Tags    []string   `json:"tags" bson:"tags"`
	State   EmoteState `json:"state" bson:"state"`

	// Versioning

	Versions    []EmoteVersion       `json:"versions" bson:"versions"`
	ChildrenIDs []primitive.ObjectID `json:"children_ids" bson:"children_ids"`
	ParentID    *primitive.ObjectID  `json:"parent_id" bson:"parent_id"`

	// Relational

	Owner    *User  `json:"owner" bson:"owner_user,skip,omitempty"`
	Channels []User `json:"channels" bson:"channels,skip,omitempty"`
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
	EmoteFlagsPrivate   EmoteFlag = 1 << 0 // The emote is private and can only be accessed by its owner, editors and moderators
	EmoteFlagsAuthentic EmoteFlag = 1 << 1 // The emote was verified to be an original creation by the uploader
	EmoteFlagsZeroWidth EmoteFlag = 1 << 8 // The emote is recommended to be enabled as Zero-Width

	// Content Flags

	EmoteFlagsContentSexual           EmoteFlag = 1 << 16 // Sexually Suggesive
	EmoteFlagsContentEpilepsy         EmoteFlag = 1 << 17 // Rapid flashing
	EmoteFlagsContentEdgy             EmoteFlag = 1 << 18 // Edgy or distasteful, may be offensive to some users
	EmoteFlagsContentTwitchDisallowed EmoteFlag = 1 << 24 // Not allowed specifically on the Twitch platform
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

type EmoteFile struct {
	Name         string `json:"name" bson:"name"`                         // The name of the file
	Width        int32  `json:"width" bson:"width,omitempty"`             // The pixel width of the emote
	Height       int32  `json:"height" bson:"height,omitempty"`           // The pixel height of the emote
	FrameCount   int32  `json:"frame_count" bson:"frame_count,omitempty"` // Whether or not this file is animated
	Size         int64  `json:"size" bson:"size"`                         // The file size in bytes
	ContentType  string `json:"content_type" bson:"content_type"`
	SHA3         string `json:"sha3" bson:"sha3"`
	Key          string `json:"key" bson:"key"`
	Bucket       string `json:"bucket" bson:"bucket"`
	ACL          string `json:"acl"`
	CacheControl string `json:"cache_control"`
}

func (ef EmoteFile) IsStatic() bool {
	return strings.HasSuffix(ef.Name, fmt.Sprintf("_static%s", path.Ext(ef.Name)))
}

func (ev EmoteVersion) CountFiles(contentType string, omitStatic bool) int32 {
	var count int32
	for _, f := range ev.ImageFiles {
		if omitStatic && (ev.Animated && f.FrameCount == 1) {
			continue
		}
	}
	return count
}

func (ev EmoteVersion) GetFiles(contentType string, omitStatic bool) []EmoteFile {
	files := []EmoteFile{}
	for _, f := range ev.ImageFiles {
		if contentType != "" && f.ContentType != contentType {
			continue
		}
		if omitStatic && f.IsStatic() {
			continue
		}
		files = append(files, f)
	}
	return files
}

type EmoteState struct {
	// IDs of users who are eligible to claim ownership of this emote
	Claimants []primitive.ObjectID `json:"claimants" bson:"claimants"`
}

type EmoteVersionState struct {
	// The current life cycle of the emote
	// indicating whether it's processing, live, deleted, etc.
	Lifecycle EmoteLifecycle `json:"lifecycle" bson:"lifecycle"`
	// Whether or not the emote is listed
	Listed bool `json:"listed" bson:"listed"`
	// The ranked position for the amount of channels this emote is added on to.
	// This value is to be determined with an external cron job
	ChannelCountRank int32 `json:"-" bson:"channel_count_rank"`
	// The amount of channels this emote is added on to.
	// This value is to be determined with an external cron job
	ChannelCount int32 `json:"-" bson:"channel_count"`
	// The time at which the ChannelCount value was last checked
	ChannelCountCheckAt time.Time `json:"-" bson:"channel_count_check_at"`
}

type EmoteVersion struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Animated    bool               `json:"animated" bson:"animated"`
	State       EmoteVersionState  `json:"state" bson:"state"`

	InputFile   EmoteFile   `json:"input_file" bson:"input_file"`
	ImageFiles  []EmoteFile `json:"image_files" bson:"image_files"`
	ArchiveFile EmoteFile   `json:"archive_file" bson:"archive_file"`

	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	StartedAt   time.Time `json:"started_at" bson:"started_at"`
	CompletedAt time.Time `json:"completed_at" bson:"completed_at"`
}

func (e Emote) HasFlag(flag EmoteFlag) bool {
	return utils.BitField.HasBits(int64(e.Flags), int64(flag))
}

func (e Emote) GetVersion(id ObjectID) (EmoteVersion, int) {
	for i, v := range e.Versions {
		if v.ID == id {
			return v, i
		}
	}
	return EmoteVersion{}, -1
}

func (e Emote) GetLatestVersion(onlyListed bool) EmoteVersion {
	var ver EmoteVersion
	for _, v := range e.Versions {
		if onlyListed && !v.State.Listed {
			continue
		}
		if v.IsUnavailable() {
			continue
		}
		if ver.ID.IsZero() || ver.CreatedAt.Before(v.CreatedAt) {
			ver = v
		}
	}
	return ver
}

func (e Emote) WebURL(origin string) string {
	return fmt.Sprintf("%s/emotes/%s", origin, e.ID.Hex())
}

func (ev EmoteVersion) IsUnavailable() bool {
	return ev.State.Lifecycle == EmoteLifecycleDeleted || ev.State.Lifecycle == EmoteLifecycleDisabled || ev.State.Lifecycle == EmoteLifecycleFailed
}

func (ev EmoteVersion) IsProcessing() bool {
	return ev.State.Lifecycle == EmoteLifecyclePending || ev.State.Lifecycle == EmoteLifecycleProcessing
}
