package structures

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmoteBuilder Wraps an Emote and offers methods to fetch and mutate emote data
type EmoteBuilder struct {
	Update UpdateMap
	Emote  Emote

	initial         Emote
	initialVersions []EmoteVersion
	tainted         bool
}

// NewEmoteBuilder: create a new emote builder
func NewEmoteBuilder(emote Emote) *EmoteBuilder {
	vers := make([]EmoteVersion, len(emote.Versions))
	copy(vers, emote.Versions)

	return &EmoteBuilder{
		Update:          UpdateMap{},
		initial:         emote,
		initialVersions: vers,
		Emote:           emote,
	}
}

// Initial returns a pointer to the value first passed to this Builder
func (eb *EmoteBuilder) Initial() Emote {
	return eb.initial
}

// IsTainted returns whether or not this Builder has been mutated before
func (eb *EmoteBuilder) IsTainted() bool {
	return eb.tainted
}

// MarkAsTainted taints the builder, preventing it from being mutated again
func (eb *EmoteBuilder) MarkAsTainted() {
	eb.tainted = true
}

func (eb *EmoteBuilder) InitialVersions() []*EmoteVersion {
	a := make([]*EmoteVersion, len(eb.initialVersions))
	for i, v := range eb.initialVersions {
		a[i] = &v
	}
	return a
}

// SetName: change the name of the emote
func (eb *EmoteBuilder) SetName(name string) *EmoteBuilder {
	eb.Emote.Name = name
	eb.Update.Set("name", eb.Emote.Name)
	return eb
}

func (eb *EmoteBuilder) SetOwnerID(id primitive.ObjectID) *EmoteBuilder {
	eb.Emote.OwnerID = id
	eb.Update.Set("owner_id", id)
	return eb
}

func (eb *EmoteBuilder) SetFlags(sum EmoteFlag) *EmoteBuilder {
	eb.Emote.Flags = sum
	eb.Update.Set("flags", sum)
	return eb
}

func (eb *EmoteBuilder) SetTags(tags []string, validate bool) *EmoteBuilder {
	uniqueTags := map[string]bool{}
	for _, v := range tags {
		if v == "" {
			continue
		}
		if !emoteTagRegex.MatchString(v) {
			continue
		}
		uniqueTags[v] = true
	}

	tags = make([]string, len(uniqueTags))
	i := 0
	for k := range uniqueTags {
		tags[i] = k
		i++
	}

	eb.Emote.Tags = tags
	eb.Update.Set("tags", tags)
	return eb
}

func (eb *EmoteBuilder) AddVersion(v EmoteVersion) *EmoteBuilder {
	for _, vv := range eb.Emote.Versions {
		if vv.ID == v.ID {
			return eb
		}
	}

	eb.Emote.Versions = append(eb.Emote.Versions, v)
	eb.Update.AddToSet("versions", v)
	return eb
}

func (eb *EmoteBuilder) UpdateVersion(id ObjectID, v EmoteVersion) *EmoteBuilder {
	ind := -1
	for i, vv := range eb.Emote.Versions {
		if vv.ID == v.ID {
			ind = i
			break
		}
	}

	eb.Emote.Versions[ind] = v
	eb.Update.Set(fmt.Sprintf("versions.%d", ind), v)
	return eb
}

func (eb *EmoteBuilder) RemoveVersion(id ObjectID) *EmoteBuilder {
	ind := -1
	for i := range eb.Emote.Versions {
		if eb.Emote.Versions[i].ID.IsZero() {
			continue
		}
		if eb.Emote.Versions[i].ID != id {
			continue
		}
		ind = i
		break
	}
	if ind == -1 {
		return eb
	}

	copy(eb.Emote.Versions[ind:], eb.Emote.Versions[ind+1:])
	eb.Emote.Versions = eb.Emote.Versions[:len(eb.Emote.Versions)-1]
	eb.Update.Pull("versions", bson.M{"id": id})
	return eb
}
