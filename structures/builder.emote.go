package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

// EmoteBuilder: Wraps an Emote and offers methods to fetch and mutate emote data
type EmoteBuilder struct {
	Update UpdateMap
	Emote  *Emote
}

// SetPrivacy: change the private state of the emote
func (eb *EmoteBuilder) SetPrivacy(isPrivate bool) *EmoteBuilder {
	if isPrivate {
		eb.Emote.Flags |= EmoteFlagsPrivate
	} else {
		eb.Emote.Flags &= EmoteFlagsPrivate
	}

	eb.Update.Set("flags", eb.Emote.Flags)
	return eb
}

// SetListed: change the listing state of the emote
func (eb *EmoteBuilder) SetListed(isListed bool) *EmoteBuilder {
	if isListed {
		eb.Emote.Flags |= EmoteFlagsListed
	} else {
		eb.Emote.Flags &= EmoteFlagsListed
	}

	eb.Update.Set("flags", eb.Emote.Flags)
	return eb
}

// SetParentID: set the emote's parent (used for versioning)
func (eb *EmoteBuilder) SetParentID(parentID primitive.ObjectID) *EmoteBuilder {
	eb.Emote.ParentID = &parentID
	eb.Update.Set("parent_id", parentID)
	return eb
}

// SetVersioningData: Set metadata for versioning
func (eb *EmoteBuilder) SetVersioningData(v EmoteVersioning) *EmoteBuilder {
	eb.Emote.Versioning = &v
	eb.Update.Set("version", v)
	return eb
}
