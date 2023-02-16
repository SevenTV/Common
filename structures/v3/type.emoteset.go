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
	Tags []string `json:"tags" bson:"tags"`
	// Bit flags for the emote set
	Flags BitField[EmoteSetFlag] `json:"flags" bson:"flags"`
	// Whether or not the emote set can be edited
	//
	// DEPRECATED - use Flags instead
	Immutable bool `json:"immutable" bson:"immutable"`
	// If true, the set is "privileged" and can only be modified by its owner or a super administrator, regardless of the "Edit Any Emote Set" permission
	//
	// DEPRECATED - use Flags instead
	Privileged bool `json:"privilleged" bson:"privileged"`
	// The emotes assigned to this set
	Emotes []ActiveEmote `json:"emotes" bson:"emotes"`
	// The set's emote capacity (slots), the maximum amount of emotes it can hold
	Capacity int32 `json:"capacity" bson:"capacity"`
	// The set's emote capacity using the quota system
	// Experimental - replaces slot capacity
	QuotaCapacity float64 `json:"capacity_experimental,omitempty" bson:"capacity_experimental,omitempty"`
	// Other emote sets that this set is based upon
	Origins []EmoteSetOrigin `json:"origins,omitempty" bson:"origins,omitempty"`
	// The ID of the user who owns this emote set
	OwnerID ObjectID `json:"owner_id" bson:"owner_id"`

	// Conditions governing how this emote set may be utilized
	Condition EmoteSetCondition `json:"condition" bson:"condition"`

	// Relational

	Owner *User `json:"owner,omitempty" bson:"owner_user,skip,omitempty"`
}

type EmoteSetFlag int32

const (
	// 1 - Set is immutable, meaning it cannot be modified
	EmoteSetFlagImmutable EmoteSetFlag = 1 << 0
	// 2 - Set is privileged, meaning it can only be modified by its owner
	// or a super administrator, regardless of the "Edit Any Emote Set" permission
	EmoteSetFlagPrivileged EmoteSetFlag = 1 << 1
	// 4 - Set can be personally entitled to a user, meaning its content can be used globally
	// but may require additional review by content moderators
	EmoteSetFlagPersonal EmoteSetFlag = 1 << 2
	// 8 - Set is distributed via a purchase, meaning its content must be approved
	// for commercial use by rightsholders
	EmoteSetFlagCommercial EmoteSetFlag = 1 << 3
)

type EmoteSetOrigin struct {
	// The ID of the referenced emote set
	ID primitive.ObjectID `json:"id" bson:"id"`
	// The weight of this set for serving emotes with names
	Weight int32 `json:"weight" bson:"weight"`
	// Slicing of active emotes inside the origin set
	Slices []uint32 `json:"slices" bson:"slices"`

	// Relational

	Set *EmoteSet
}

type EmoteSetCondition struct {
	// If true, this emote set may be entitled on to a user and used globally
	Entitlable bool `json:"entitlable,omitempty" bson:"entitlable,omitempty"`
	// A list of channel IDs (user connections) where this set is allowed to be used. If empty it is unrestricted
	Channels []string `json:"channels,omitempty" bson:"channels,omitempty"`
}

const (
	EmoteSetNameLengthLeast int = 48
	EmoteSetNameLengthMost  int = 3
)

type ActiveEmote struct {
	ID              primitive.ObjectID        `json:"id" bson:"id"`
	Name            string                    `json:"name" bson:"name"`
	Flags           BitField[ActiveEmoteFlag] `json:"flags" bson:"flags"`
	Timestamp       time.Time                 `json:"timestamp" bson:"timestamp"`
	TimestampUpdate *time.Time                `json:"timestamp_update,omitempty" bson:"timestamp_update,omitempty"`
	MergedFromID    primitive.ObjectID        `json:"merged_from_id,omitempty" bson:"merged_from_id,omitempty"`
	MergedAt        time.Time                 `json:"merged_at,omitempty" bson:"merged_at,omitempty"`
	ActorID         primitive.ObjectID        `json:"actor_id,omitempty" bson:"actor_id,omitempty"`

	// Relational

	Origin EmoteSetOrigin `json:"-" bson:"-"`
	Emote  *Emote         `json:"emote" bson:"emote,omitempty,skip"`
	Actor  *User          `json:"actor" bson:"actor,omitempty,skip"`
}

type ActiveEmoteFlag int32

const (
	ActiveEmoteFlagZeroWidth                ActiveEmoteFlag = 1 << 0  // 1 - Emote is zero-width
	ActiveEmoteFlagPending                  ActiveEmoteFlag = 1 << 8  // 256 - Emote is pending review
	ActiveEmoteFlagOverrideTwitchGlobal     ActiveEmoteFlag = 1 << 16 // 65536 - Overrides Twitch Global emotes with the same name
	ActiveEmoteFlagOverrideTwitchSubscriber ActiveEmoteFlag = 1 << 17 // 131072 - Overrides Twitch Subscriber emotes with the same name
	ActiveEmoteFlagOverrideBetterTTV        ActiveEmoteFlag = 1 << 18 // 262144 - Overrides BetterTTV emotes with the same name
	ActiveEmoteFlagOverrideFrankerFaceZ     ActiveEmoteFlag = 1 << 19 // 524288 - Overrides FrankerFaceZ emotes with the same name
)

// HasEmote: returns whether or not the set has an emote active, as well as its index
func (es EmoteSet) GetEmote(id primitive.ObjectID) (ActiveEmote, int) {
	for i, ae := range es.Emotes {
		if ae.ID == id {
			return ae, i
		}
	}
	return ActiveEmote{}, -1
}
