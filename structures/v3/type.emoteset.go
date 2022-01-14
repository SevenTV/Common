package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

type EmoteSet struct {
	ID primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	// Numeric unique ID for the set
	// Starts from 1 and increments per set created
	NumericID int32 `json:"id" bson:"num_id"`
	// A unique tag for the set, between 2 and 5 characters (letters and numbers only)
	// This can be used for targeting and parsing
	Tag string `json:"tag" bson:"tag"`
	// Whether or not the emote set can be edited
	Immutable bool `json:"immutable" bson:"immutable"`
	// If true, the set is "privileged" and can only be modified by its owner or a super administrator, regardless of the "Edit Any Emote Set" permission
	Privileged bool `json:"privilleged" bson:"privileged"`
	// Whether or not the set is active. When false, the set isn't returned in various API endpoints
	Active bool `json:"active" bson:"active"`
	// The emotes assigned to this set
	Emotes []*ActiveEmote `json:"emotes" bson:"emotes"`
	// The maximum amount of emotes this set is allowed to contain
	EmoteSlots int32 `json:"emote_slots" bson:"emote_slots"`
	// The ID of the user who owns this emote
	OwnerID primitive.ObjectID `json:"owner_id" bson:"owner_id"`

	// TODO: filters, i.e allow set to be used only in specific channels

	// Relational

	Owner *User `json:"owner" bson:"owner"`
}

type ActiveEmote struct {
	ID    primitive.ObjectID `json:"id" bson:"id"`
	Alias string             `json:"alias,omitempty" bson:"alias,omitempty"`

	// Relational

	Emote *Emote `json:"emote" bson:"emote,omitempty,skip"`
}

type EmoteSetBuilder struct {
	Update   UpdateMap
	EmoteSet *EmoteSet
}

func (esb *EmoteSetBuilder) SetNumericID(i int32) *EmoteSetBuilder {
	esb.EmoteSet.NumericID = i
	esb.Update.Set("num_id", i)
	return esb
}

func (esb *EmoteSetBuilder) SetTag(tag string) *EmoteSetBuilder {
	esb.EmoteSet.Tag = tag
	esb.Update.Set("tag", tag)
	return esb
}

func (esb *EmoteSetBuilder) SetImmutable(b bool) *EmoteSetBuilder {
	esb.EmoteSet.Immutable = b
	esb.Update.Set("immutable", b)
	return esb
}

func (esb *EmoteSetBuilder) SetPrivileged(b bool) *EmoteSetBuilder {
	esb.EmoteSet.Privileged = b
	esb.Update.Set("privileged", b)
	return esb
}

func (esb *EmoteSetBuilder) SetActive(b bool) *EmoteSetBuilder {
	esb.EmoteSet.Active = b
	esb.Update.Set("active", b)
	return esb
}

func (esb *EmoteSetBuilder) SetEmoteSlots(slots int32) *EmoteSetBuilder {
	esb.EmoteSet.EmoteSlots = slots
	esb.Update.Set("emote_slots", slots)
	return esb
}

func (esb *EmoteSetBuilder) SetOwnerID(id primitive.ObjectID) *EmoteSetBuilder {
	esb.EmoteSet.OwnerID = id
	esb.Update.Set("owner_id", id)
	return esb
}
