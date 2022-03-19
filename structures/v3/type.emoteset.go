package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmoteSet struct {
	ID primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	// The emote set's name
	Name string `json:"name" bson:"name"`
	// Search tags for the emote set
	Tags []string `json:"tags,omitempty" bson:"tags,omitempty"`
	// Whether or not the emote set can be edited
	Immutable bool `json:"immutable" bson:"immutable"`
	// If true, the set is "privileged" and can only be modified by its owner or a super administrator, regardless of the "Edit Any Emote Set" permission
	Privileged bool `json:"privilleged" bson:"privileged"`
	// The emotes assigned to this set
	Emotes []*ActiveEmote `json:"emotes" bson:"emotes"`
	// The maximum amount of emotes this set is allowed to contain
	EmoteSlots int32 `json:"emote_slots" bson:"emote_slots"`
	// The ID of the parent set. If defined, this set is treated as a child set
	// and its emotes are derived from the parent
	ParentID *ObjectID `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
	// The ID of the user who owns this emote set
	OwnerID ObjectID `json:"owner_id" bson:"owner_id"`

	// TODO: filters, i.e allow set to be used only in specific channels

	// Relational

	Owner *User `json:"owner,omitempty" bson:"owner_user,skip,omitempty"`
}

const (
	EmoteSetNameLengthLeast int = 48
	EmoteSetNameLengthMost  int = 3
)

type ActiveEmote struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty"`
	Flags     ActiveEmoteFlag    `json:"flags" bson:"flags"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	ActorID   primitive.ObjectID `json:"actor_id,omitempty" bson:"actor_id,omitempty"`

	// Relational

	Emote *Emote `json:"emote" bson:"emote,omitempty,skip"`
}

type ActiveEmoteFlag int32

const (
	ActiveEmoteFlagZeroWidth                ActiveEmoteFlag = 1 << 0  // 1 - Emote is zero-width
	ActiveEmoteFlagOverrideTwitchGlobal     ActiveEmoteFlag = 1 << 16 // 65536 - Overrides Twitch Global emotes with the same name
	ActiveEmoteFlagOverrideTwitchSubscriber ActiveEmoteFlag = 1 << 17 // 131072 - Overrides Twitch Subscriber emotes with the same name
	ActiveEmoteFlagOverrideBetterTTV        ActiveEmoteFlag = 1 << 18 // 262144 - Overrides BetterTTV emotes with the same name
	ActiveEmoteFlagOverrideFrankerFaceZ     ActiveEmoteFlag = 1 << 19 // 524288 - Overrides FrankerFaceZ emotes with the same name
)

// HasEmote: returns whether or not the set has an emote active, as well as its index
func (es *EmoteSet) HasEmote(id primitive.ObjectID) (bool, int) {
	for i, ae := range es.Emotes {
		if ae.ID == id {
			return true, i
		}
	}
	return false, -1
}
