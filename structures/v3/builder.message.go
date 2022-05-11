package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageBuilder[D MessageData] struct {
	Message Message[D]
	Update  UpdateMap

	tainted bool
}

func NewMessageBuilder[D MessageData](msg Message[D]) *MessageBuilder[D] {
	msg.CreatedAt = time.Now()
	return &MessageBuilder[D]{
		Update:  UpdateMap{},
		Message: msg,
	}
}

// IsTainted returns whether or not this Builder has been mutated before
func (eb *MessageBuilder[D]) IsTainted() bool {
	return eb.tainted
}

// MarkAsTainted taints the builder, preventing it from being mutated again
func (eb *MessageBuilder[D]) MarkAsTainted() {
	eb.tainted = true
}

func (mb *MessageBuilder[D]) SetKind(kind MessageKind) *MessageBuilder[D] {
	mb.Message.Kind = kind
	mb.Update.Set("kind", kind)
	return mb
}

func (mb *MessageBuilder[D]) SetAuthorID(id primitive.ObjectID) *MessageBuilder[D] {
	mb.Message.AuthorID = id
	mb.Update.Set("author_id", id)
	return mb
}

func (mb *MessageBuilder[D]) SetAnonymous(b bool) *MessageBuilder[D] {
	mb.Message.Anonymous = b
	mb.Update.Set("anonymous", b)
	return mb
}

func (mb *MessageBuilder[D]) SetTimestamp(t time.Time) *MessageBuilder[D] {
	mb.Message.CreatedAt = t
	mb.Update.Set("created_at", t)
	return mb
}

func (mb *MessageBuilder[D]) SetData(data D) *MessageBuilder[D] {
	mb.Message.Data = data
	mb.Update.Set("data", data)
	return mb
}
