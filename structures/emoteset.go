package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

type EmoteSetBuidler struct {
	EmoteSet *EmoteSet
}

// SetImmutable: choose whether or not the emote set is immutable
func (b EmoteSetBuidler) SetImmutable(v bool) EmoteSetBuidler {
	b.EmoteSet.Immutable = v
	return b
}

// SetActive: choose whether or not the emote set is active
func (b EmoteSetBuidler) SetActive(v bool) EmoteSetBuidler {
	b.EmoteSet.Active = v
	return b
}

// AddEmotes: add new emotes to the set
func (b EmoteSetBuidler) AddEmotes(ids ...primitive.ObjectID) EmoteSetBuidler {
	b.EmoteSet.EmoteIDs = append(b.EmoteSet.EmoteIDs, ids...)
	return b
}

// SetEmoteSlots: define the maximum amount of emotes that are allowed in this set
func (b EmoteSetBuidler) SetEmoteSlots(count int32) EmoteSetBuidler {
	b.EmoteSet.EmoteSlots = count
	return b
}

// AddEditors: add new editors to the set
func (b EmoteSetBuidler) AddEditors(ids ...primitive.ObjectID) EmoteSetBuidler {
	b.EmoteSet.EditorIDs = append(b.EmoteSet.EditorIDs, ids...)
	return b
}

// SetType: define the set's type
func (b EmoteSetBuidler) SetType(t EmoteSetType) EmoteSetBuidler {
	b.EmoteSet.Type = t
	return b
}

type EmoteSet struct {
	ID primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	// Numeric unique ID for the set
	// Starts from 1 and increments per set created
	NumericID string `json:"id" bson:"id"`
	// Whether or not the emote set can be edited
	Immutable bool `json:"immutable" bson:"immutable"`
	// Whether or not the set is active. When false, the set isn't returned in various API endpoints
	Active bool `json:"active" bson:"active"`
	// The emotes assigned to this set
	EmoteIDs []primitive.ObjectID `json:"emote_ids" bson:"emote_ids"`
	// The maximum amount of emotes this set is allowed to contain
	EmoteSlots int32 `json:"emote_slots" bson:"emote_slots"`
	// The set's editors, who are allowed to edit the set's emotes
	EditorIDs []primitive.ObjectID `json:"editor_ids" bson:"editor_ids"`

	// The type of emote set. Can be SELECTIVE or GLOBAL. if SELECTIVE, the only applies to select channels
	Type EmoteSetType `json:"type" bson:"type"`

	// Relational
	Editors []*User `json:"editors" bson:"-"`
}

type EmoteSetType string

var (
	// Global Sets apply to all channels, regardless of channel selection
	EmoteSetTypeGlobal EmoteSetType = "GLOBAL"
	// Selective Sets apply only to select channels
	EmoteSetTypeSelective EmoteSetType = "SELECTIVE"
)
