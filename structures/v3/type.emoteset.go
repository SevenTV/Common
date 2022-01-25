package structures

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

type EmoteSetBuilder struct {
	Update   UpdateMap
	EmoteSet *EmoteSet

	Initial EmoteSet
}

func NewEmoteSetBuilder(emoteSet *EmoteSet) *EmoteSetBuilder {
	init := EmoteSet{
		Tags:   []string{},
		Emotes: []*ActiveEmote{},
	}
	if emoteSet == nil {
		emoteSet = &init
	}
	return &EmoteSetBuilder{
		Update:   map[string]interface{}{},
		EmoteSet: emoteSet,
		Initial:  init,
	}
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

func (esb *EmoteSetBuilder) SetParentID(id *ObjectID) *EmoteSetBuilder {
	esb.EmoteSet.ParentID = id
	esb.Update.Set("parent_id", id)
	return esb
}

func (esb *EmoteSetBuilder) SetEmoteSlots(slots int32) *EmoteSetBuilder {
	esb.EmoteSet.EmoteSlots = slots
	esb.Update.Set("emote_slots", slots)
	return esb
}

func (esb *EmoteSetBuilder) SetOwnerID(id ObjectID) *EmoteSetBuilder {
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
		Name:      alias,
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
	v.Name = alias
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
	esb.Update.Pull("emotes", bson.M{"id": id})
	return esb
}
