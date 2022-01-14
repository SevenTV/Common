package structures

import (
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// The kind of message this is (i.e a comment, or inbox message)
	Kind MessageKind `json:"kind" bson:"kind"`
	// The ID of the user who created this message
	AuthorID primitive.ObjectID `json:"author_id" bson:"author_id,omitempty"`
	// Whether or not the message's author will not be displayed to unprivileged end users
	Anonymous bool `json:"anonymous,omitempty" bson:"anonymous,omitempty"`
	// The date on which this message was cretaed
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	// Message data
	Data bson.Raw `json:"data" bson:"data"`

	// Relational

	Author *User `json:"author,omitempty" bson:"author,skip,omitempty"`
	Read   bool  `json:"read,omitempty" bson:"read,omitempty"`
}

type MessageKind string

const (
	MessageKindEmoteComment MessageKind = "EMOTE_COMMENT" // a comment
	MessageKindInbox        MessageKind = "INBOX"         // an inbox message
	MessageKindNews         MessageKind = "NEWS"          // a news post
)

type MessageData bson.Raw

type MessageDataEmoteComment struct {
	EmoteID primitive.ObjectID `json:"emote_id" bson:"emote_id"`
	// Whether or not the comment is an official statement
	// i.e, a warning by a moderator
	Authoritative bool `json:"authoritative,omitempty" bson:"authoritative,omitempty"`
	// Whether or not the comment is pinned.
	// Pinned comments will always appear at the top
	Pinned bool `json:"pinned,omitempty" bson:"pinned,omitempty"`
	// The comment's text contents
	Content string `json:"content" bson:"content"`
}

type MessageDataInbox struct {
	Subject      string            `json:"subject" bson:"subject"`
	Content      string            `json:"content" bson:"content"`
	Important    bool              `json:"important,omitempty" bson:"important,omitempty"`
	Starred      bool              `json:"starred,omitempty" bson:"starred,omitempty"`
	Pinned       bool              `json:"pinned,omitempty" bson:"pinned,omitempty"`
	Placeholders map[string]string `json:"placeholders,omitempty" bson:"placeholders,omitempty"`
}

type MessageDataPlaceholder struct {
	Key   string
	Value string
}

type MessageBuilder struct {
	Message *Message
	Update  UpdateMap
}

func NewMessageBuilder(msg *Message) *MessageBuilder {
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

func (mb *MessageBuilder) DecodeInbox() *MessageDataInbox {
	return mb.unmarshal(&MessageDataInbox{}).(*MessageDataInbox)
}

func (mb *MessageBuilder) unmarshal(i interface{}) interface{} {
	if err := bson.Unmarshal(mb.Message.Data, i); err != nil {
		logrus.WithError(err).Error("message, decoding message data failed")
	}
	return i
}

// MessageRead read/unread state for a message
type MessageRead struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MessageID   primitive.ObjectID `json:"message_id" bson:"message_id"`
	RecipientID primitive.ObjectID `json:"recipient_id" bson:"recipient_id"`
	Read        bool               `json:"read" bson:"read"`
	ReadAt      time.Time          `json:"read_at,omitempty" bson:"read_at,omitempty"`
}
