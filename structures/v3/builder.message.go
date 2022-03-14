package structures

import (
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageBuilder struct {
	Message *Message
	Update  UpdateMap
}

func NewMessageBuilder(msg *Message) *MessageBuilder {
	if msg == nil {
		msg = &Message{
			CreatedAt: time.Time{},
		}
	} else if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	return &MessageBuilder{
		Update:  UpdateMap{},
		Message: msg,
	}
}

func (mb *MessageBuilder) SetKind(kind MessageKind) *MessageBuilder {
	mb.Message.Kind = kind
	mb.Update.Set("kind", kind)
	return mb
}

func (mb *MessageBuilder) SetAuthorID(id primitive.ObjectID) *MessageBuilder {
	mb.Message.AuthorID = id
	mb.Update.Set("author_id", id)
	return mb
}

func (mb *MessageBuilder) SetAnonymous(b bool) *MessageBuilder {
	mb.Message.Anonymous = b
	mb.Update.Set("anonymous", b)
	return mb
}

func (mb *MessageBuilder) SetTimestamp(t time.Time) *MessageBuilder {
	mb.Message.CreatedAt = t
	mb.Update.Set("created_at", t)
	return mb
}

func (mb *MessageBuilder) AsEmoteComment(d MessageDataEmoteComment) *MessageBuilder {
	mb.encodeData(d)
	return mb
}

func (mb *MessageBuilder) AsModRequest(d MessageDataModRequest) *MessageBuilder {
	mb.encodeData(d)
	return mb
}

func (mb *MessageBuilder) AsInbox(d MessageDataInbox) *MessageBuilder {
	mb.encodeData(d)
	return mb
}

func (mb *MessageBuilder) encodeData(i interface{}) {
	b, err := bson.Marshal(i)
	if err != nil {
		logrus.WithError(err).Error("message, encoding message data failed")
		return
	}
	mb.Message.Data = b
}

func (mb *MessageBuilder) DecodeEmoteComment() *MessageDataEmoteComment {
	return mb.unmarshal(&MessageDataEmoteComment{}).(*MessageDataEmoteComment)
}

func (mb *MessageBuilder) DecodeModRequest() *MessageDataModRequest {
	return mb.unmarshal(&MessageDataModRequest{}).(*MessageDataModRequest)
}

func (mb *MessageBuilder) DecodeInbox() *MessageDataInbox {
	return mb.unmarshal(&MessageDataInbox{}).(*MessageDataInbox)
}

func (mb *MessageBuilder) unmarshal(i interface{}) interface{} {
	if err := bson.Unmarshal(mb.Message.Data, i); err != nil {
		logrus.WithError(err).Error("message, decoding message data failed")
	}
	return i
}
