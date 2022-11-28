package structures

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmoteSetBuilder struct {
	Update   UpdateMap
	EmoteSet EmoteSet

	initial EmoteSet
	tainted bool
}

func NewEmoteSetBuilder(emoteSet EmoteSet) *EmoteSetBuilder {
	return &EmoteSetBuilder{
		Update:   map[string]interface{}{},
		EmoteSet: emoteSet,
		initial:  emoteSet,
	}
}

// Initial returns a pointer to the value first passed to this Builder
func (esb *EmoteSetBuilder) Initial() *EmoteSet {
	return &esb.initial
}

// IsTainted returns whether or not this Builder has been mutated before
func (esb *EmoteSetBuilder) IsTainted() bool {
	return esb.tainted
}

// MarkAsTainted taints the builder, preventing it from being mutated again
func (esb *EmoteSetBuilder) MarkAsTainted() {
	esb.tainted = true
}

func (esb *EmoteSetBuilder) SetName(name string) *EmoteSetBuilder {
	esb.EmoteSet.Name = name
	esb.Update.Set("name", name)
	return esb
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

func (esb *EmoteSetBuilder) SetOrigins(origins []EmoteSetOrigin) *EmoteSetBuilder {
	esb.EmoteSet.Origins = origins
	esb.Update.Set("origins", origins)
	return esb
}

func (esb *EmoteSetBuilder) AddOrigin(id ObjectID, weight int32) *EmoteSetBuilder {
	v := EmoteSetOrigin{
		ID:     id,
		Weight: weight,
	}

	esb.EmoteSet.Origins = append(esb.EmoteSet.Origins, v)
	esb.Update.Push("origins", v)
	return esb
}

func (esb *EmoteSetBuilder) RemoveOrigin(id ObjectID) *EmoteSetBuilder {
	ind := -1
	for i := range esb.EmoteSet.Origins {
		if esb.EmoteSet.Origins[i].ID.IsZero() {
			continue
		}
		if esb.EmoteSet.Origins[i].ID != id {
			continue
		}
		ind = i
		break
	}
	if ind == -1 {
		return esb // did not find index
	}

	copy(esb.EmoteSet.Origins[ind:], esb.EmoteSet.Origins[ind+1:])
	esb.EmoteSet.Origins = esb.EmoteSet.Origins[:len(esb.EmoteSet.Origins)-1]
	esb.Update.Pull("origins", bson.M{"id": id})
	return esb
}

func (esb *EmoteSetBuilder) SetCapacity(slots int32) *EmoteSetBuilder {
	esb.EmoteSet.Capacity = slots
	esb.Update.Set("capacity", slots)
	return esb
}

func (esb *EmoteSetBuilder) SetOwnerID(id ObjectID) *EmoteSetBuilder {
	esb.EmoteSet.OwnerID = id
	esb.Update.Set("owner_id", id)
	return esb
}

func (esb *EmoteSetBuilder) AddActiveEmote(id ObjectID, alias string, at time.Time, actorID *primitive.ObjectID) *EmoteSetBuilder {
	for _, e := range esb.EmoteSet.Emotes {
		if e.ID == id {
			return esb // emote already added.
		}
	}

	v := ActiveEmote{
		ID:        id,
		Name:      alias,
		Timestamp: at,
	}
	if actorID != nil && !actorID.IsZero() {
		v.ActorID = *actorID
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
	v.Name = alias
	esb.Update.Set(fmt.Sprintf("emotes.%d", ind), v)
	return esb
}

func (esb *EmoteSetBuilder) RemoveActiveEmote(id ObjectID) (*EmoteSetBuilder, int) {
	ind := -1
	for i := range esb.EmoteSet.Emotes {
		if esb.EmoteSet.Emotes[i].ID.IsZero() {
			continue
		}
		if esb.EmoteSet.Emotes[i].ID != id {
			continue
		}
		ind = i
		break
	}
	if ind == -1 {
		return esb, ind // did not find index
	}

	copy(esb.EmoteSet.Emotes[ind:], esb.EmoteSet.Emotes[ind+1:])
	esb.EmoteSet.Emotes = esb.EmoteSet.Emotes[:len(esb.EmoteSet.Emotes)-1]
	esb.Update.Pull("emotes", bson.M{"id": id})
	return esb, ind
}
