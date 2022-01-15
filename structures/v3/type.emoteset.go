package structures

import (
	"fmt"
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
	// Whether or not the set is active. When false, the set isn't returned in various API endpoints
	Active bool `json:"active" bson:"active"`
	// The emotes assigned to this set
	Emotes []*ActiveEmote `json:"emotes" bson:"emotes"`
	// The maximum amount of emotes this set is allowed to contain
	EmoteSlots uint32 `json:"emote_slots" bson:"emote_slots"`
	// The ID of the user who owns this emote set
	OwnerID primitive.ObjectID `json:"owner_id" bson:"owner_id"`

	// TODO: filters, i.e allow set to be used only in specific channels

	// Relational

	Owner *User `json:"owner" bson:"owner"`
}

type ActiveEmote struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	Alias     string             `json:"alias,omitempty" bson:"alias,omitempty"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`

	// Relational

	Emote *Emote `json:"emote" bson:"emote,omitempty,skip"`
}

type EmoteSetBuilder struct {
	Update   UpdateMap
	EmoteSet *EmoteSet
}

func (esb *EmoteSetBuilder) SetTags(tags []string) *EmoteSetBuilder {
	esb.EmoteSet.Tags = tags
	esb.Update.Set("tags", tags)
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

func (esb *EmoteSetBuilder) SetEmoteSlots(slots uint32) *EmoteSetBuilder {
	esb.EmoteSet.EmoteSlots = slots
	esb.Update.Set("emote_slots", slots)
	return esb
}

func (esb *EmoteSetBuilder) SetOwnerID(id primitive.ObjectID) *EmoteSetBuilder {
	esb.EmoteSet.OwnerID = id
	esb.Update.Set("owner_id", id)
	return esb
}

func (esb *EmoteSetBuilder) AddActiveEmote(id ObjectID, alias string, at time.Time) *EmoteSetBuilder {
	for _, e := range esb.EmoteSet.Emotes {
		if e.ID == id {
			return esb // emote already added.
		}
	}

	v := &ActiveEmote{
		ID:        id,
		Alias:     alias,
		Timestamp: at,
	}
	esb.EmoteSet.Emotes = append(esb.EmoteSet.Emotes, v)
	esb.Update.AddToSet("emotes", v)
	return esb
}

func (esb *EmoteSetBuilder) UpdateActiveEmote(id ObjectID, alias string) *EmoteSetBuilder {
	ind := -1
	for i, e := range esb.EmoteSet.Emotes {
		if e.ID == id {
			ind = i
			break
		}
	}

	v := esb.EmoteSet.Emotes[ind]
	v.Alias = alias
	esb.Update.Set(fmt.Sprintf("emotes.%d", ind), v)
	return esb
}

func (esb *EmoteSetBuilder) RemoveActiveEmote(id ObjectID) *EmoteSetBuilder {
	ind := -1
	for i := range esb.EmoteSet.Emotes {
		if esb.EmoteSet.Emotes[i] == nil {
			continue
		}
		if esb.EmoteSet.Emotes[i].ID != id {
			continue
		}
		ind = i
		break
	}
	if ind == -1 {
		return esb // did not find index
	}

	copy(esb.EmoteSet.Emotes[ind:], esb.EmoteSet.Emotes[ind+1:])
	esb.EmoteSet.Emotes[len(esb.EmoteSet.Emotes)-1] = nil
	esb.EmoteSet.Emotes = esb.EmoteSet.Emotes[:len(esb.EmoteSet.Emotes)-1]
	esb.Update.Pull("emotes.id", id)
	return esb
}
