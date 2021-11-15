package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

type EmoteSetBuilder struct {
	Update   UpdateMap
	EmoteSet *EmoteSet
}

type EmoteSet struct {
	ID primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	// Numeric unique ID for the set
	// Starts from 1 and increments per set created
	NumericID string `json:"id" bson:"num_id"`
	// A unique tag for the set, between 2 and 5 characters (letters and numbers only)
	// This can be used for targeting and parsing
	Tag string `json:"tag" bson:"tag"`
	// Whether or not the emote set can be edited
	Immutable bool `json:"immutable" bson:"immutable"`
	// If true, the set is "privileged" and can only be modified by its editors, regardless of the "Edit Any Emote Set" permission
	Privileged bool `json:"privilleged" bson:"privileged"`
	// Whether or not the set is active. When false, the set isn't returned in various API endpoints
	Active bool `json:"active" bson:"active"`
	// The emotes assigned to this set
	Emotes []*ActiveEmote `json:"emote_ids" bson:"emote_ids"`
	// The maximum amount of emotes this set is allowed to contain
	EmoteSlots int32 `json:"emote_slots" bson:"emote_slots"`
	// The set's editors, who are allowed to edit the set's emotes
	EditorIDs []primitive.ObjectID `json:"editor_ids" bson:"editor_ids"`

	// The type of emote set. Can be SELECTIVE or GLOBAL. if SELECTIVE, the only applies to select channels
	Type EmoteSetType `json:"type" bson:"type"`
	// The channels this set may apply to. This is only relevant when the set type is "SELECTIVE"
	Selection []primitive.ObjectID `json:"selection" bson:"selection"`

	// Relational
	Editors []*User `json:"editors" bson:"editors,skip"`
}

// The type of emote set
type EmoteSetType string

var (
	// Global Sets apply to all channels, regardless of channel selection
	EmoteSetTypeGlobal EmoteSetType = "GLOBAL"
	// Selective Sets apply only to select channels
	EmoteSetTypeSelective EmoteSetType = "SELECTIVE"
)

type ActiveEmote struct {
	ID    primitive.ObjectID `json:"id" bson:"id"`
	Alias string             `json:"alias,omitempty" bson:"alias,omitempty"`

	// Relational

	Emote *Emote `json:"emote" bson:"emote,omitempty,skip"`
}
