package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicActiveEmote struct {
	ID        primitive.ObjectID `json:"id"`
	Name      string             `json:"name"`
	Flags     ActiveEmoteFlag    `json:"flags"`
	Timestamp time.Time          `json:"timestamp"`
	ActorID   primitive.ObjectID `json:"actor_id"`

	Emote *PublicEmote `json:"emote,omitempty"`
}

func (ae ActiveEmote) ToPublic(emote PublicEmote) PublicActiveEmote {
	var e *PublicEmote
	if !emote.ID.IsZero() {
		e = &emote
	}

	return PublicActiveEmote{
		ID:        ae.ID,
		Name:      ae.Name,
		Flags:     ae.Flags,
		Timestamp: ae.Timestamp,
		ActorID:   ae.ActorID,
		Emote:     e,
	}
}
