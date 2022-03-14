package structures

import (
	"time"

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

type MessageKind int8

const (
	MessageKindEmoteComment MessageKind = 1 // a comment
	MessageKindModRequest   MessageKind = 2 // a moderator action request
	MessageKindInbox        MessageKind = 3 // an inbox message
	MessageKindNews         MessageKind = 4 // a news post
)

func (k MessageKind) String() string {
	switch k {
	case MessageKindEmoteComment:
		return "EMOTE_COMMENT"
	case MessageKindModRequest:
		return "MOD_REQUEST"
	case MessageKindInbox:
		return "INBOX"
	case MessageKindNews:
		return "NEWS"
	default:
		return ""
	}
}

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

type MessageDataModRequest struct {
	TargetKind ObjectKind         `json:"target_kind" bson:"target_kind"`
	TargetID   primitive.ObjectID `json:"target_id" bson:"target_id"`
}

// MessageRead read/unread state for a message
type MessageRead struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Kind        MessageKind        `json:"kind" bson:"kind"`
	Timestamp   time.Time          `json:"timestamp" bson:"timestamp"`
	MessageID   primitive.ObjectID `json:"message_id" bson:"message_id"`
	RecipientID primitive.ObjectID `json:"recipient_id,omitempty" bson:"recipient_id,omitempty"`
	Read        bool               `json:"read" bson:"read"`
	ReadAt      time.Time          `json:"read_at,omitempty" bson:"read_at,omitempty"`
}
