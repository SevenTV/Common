package structures

import (
	"time"

	"github.com/seventv/common/utils"
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

type MessageComponentBuilder struct {
	MessageComponent MessageComponent
	Update           UpdateMap
}

func NewMessageComponentBuilder(mcb MessageComponent) *MessageComponentBuilder {
	return &MessageComponentBuilder{
		MessageComponent: mcb,
		Update:           UpdateMap{},
	}
}

func (mcb *MessageComponentBuilder) SetType(t MessageComponentType) *MessageComponentBuilder {
	mcb.MessageComponent.Type = t
	mcb.Update.Set("type", t)

	return mcb
}

func (mcb *MessageComponentBuilder) SetHeading(h uint8) *MessageComponentBuilder {
	mcb.MessageComponent.Heading = h
	mcb.Update.Set("heading", h)

	return mcb
}

func (mcb *MessageComponentBuilder) SetWeight(w uint8) *MessageComponentBuilder {
	mcb.MessageComponent.Weight = w
	mcb.Update.Set("weight", w)

	return mcb
}

func (mcb *MessageComponentBuilder) SetColor(c utils.Color) *MessageComponentBuilder {
	mcb.MessageComponent.Color = c
	mcb.Update.Set("color", c)

	return mcb
}

func (mcb *MessageComponentBuilder) SetLocale(l bool) *MessageComponentBuilder {
	mcb.MessageComponent.Locale = l
	mcb.Update.Set("locale", l)

	return mcb
}

func (mcb *MessageComponentBuilder) SetContent(c string) *MessageComponentBuilder {
	mcb.MessageComponent.Content = c
	mcb.Update.Set("content", c)

	return mcb
}
