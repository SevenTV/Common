package structures

import (
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageData interface {
	bson.Raw | MessageDataEmoteComment | MessageDataInbox | MessageDataPlaceholder | MessageDataModRequest
}

type Message[D MessageData] struct {
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
	Data D `json:"data" bson:"data"`

	// Relational

	Author    *User     `json:"author,omitempty" bson:"author,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Read      bool      `json:"read,omitempty" bson:"read,omitempty"`
}

func (m Message[D]) ToRaw() Message[bson.Raw] {
	switch x := utils.ToAny(m.Data).(type) {
	case bson.Raw:
		return Message[bson.Raw]{
			ID:        m.ID,
			Kind:      m.Kind,
			AuthorID:  m.AuthorID,
			Anonymous: m.Anonymous,
			CreatedAt: m.CreatedAt,
			Data:      x,
			Author:    m.Author,
			Read:      m.Read,
		}
	}

	raw, _ := bson.Marshal(m.Data)
	return Message[bson.Raw]{
		ID:        m.ID,
		Kind:      m.Kind,
		AuthorID:  m.AuthorID,
		Anonymous: m.Anonymous,
		CreatedAt: m.CreatedAt,
		Data:      raw,
		Author:    m.Author,
		Read:      m.Read,
	}
}

func ConvertMessage[D MessageData](c Message[bson.Raw]) (Message[D], error) {
	var d D
	err := bson.Unmarshal(c.Data, &d)
	c2 := Message[D]{
		ID:        c.ID,
		Kind:      c.Kind,
		AuthorID:  c.AuthorID,
		Anonymous: c.Anonymous,
		CreatedAt: c.CreatedAt,
		Data:      d,
		Author:    c.Author,
		Read:      c.Read,
	}

	return c2, err
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
	Subject      string             `json:"subject" bson:"subject"`                               // the message's subject
	Components   []MessageComponent `json:"components" bson:"components"`                         // the message's components
	Content      string             `json:"content" bson:"content"`                               // the content of the message
	Important    bool               `json:"important,omitempty" bson:"important,omitempty"`       // whether or not the message is important
	Starred      bool               `json:"starred,omitempty" bson:"starred,omitempty"`           // whether or not the message is started
	Pinned       bool               `json:"pinned,omitempty" bson:"pinned,omitempty"`             // whether or not the message is pinned
	Locked       bool               `json:"locked,omitempty" bson:"locked,omitempty"`             // whether or not replies can be added to this message
	Locale       bool               `json:"locale,omitempty" bson:"locale,omitempty"`             // whether or not this message can use locale strings
	System       bool               `json:"system,omitempty" bson:"system,omitempty"`             // whether or not the message is a system message
	Placeholders map[string]string  `json:"placeholders,omitempty" bson:"placeholders,omitempty"` // placeholders for localization
}

type MessageComponent struct {
	Type    MessageComponentType `json:"type" bson:"type"`                           // the type of component
	Heading uint8                `json:"heading,omitempty" bson:"heading,omitempty"` // component heading level
	Weight  uint8                `json:"weight,omitempty" bson:"weight,omitempty"`   // component font weight (1-9)
	Color   utils.Color          `json:"color" bson:"color"`                         // text color (ineffective on some types)
	Locale  bool                 `json:"locale" bson:"locale"`                       // whether or not the content of this component is a locale string
	Content string               `json:"content" bson:"content"`                     // the content of the component
}

type MessageComponentType string

const (
	MessageComponentTypeImage    MessageComponentType = "image"
	MessageComponentTypeLink     MessageComponentType = "link"
	MessageComponentTypeText     MessageComponentType = "text"
	MessageComponentTypeMention  MessageComponentType = "mention"
	MessageComponentTypeInteract MessageComponentType = "interact"
)

type MessageDataPlaceholder struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type MessageDataModRequest struct {
	TargetKind ObjectKind         `json:"target_kind" bson:"target_kind"`
	TargetID   primitive.ObjectID `json:"target_id" bson:"target_id"`
	Wish       string             `json:"wish" bson:"wish"`

	Target bson.Raw `json:"target" bson:"target,omitempty"`
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

	// Relational

	Message *Message[bson.Raw] `json:"message,omitempty" bson:"message,skip,omitempty"`
}
