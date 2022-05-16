package events

import (
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageType string

const (
	// Emote

	MessageTypeCreateEmote MessageType = "CREATE_EMOTE"
	MessageTypeUpdateEmote MessageType = "UPDATE_EMOTE"
	MessageTypeDeleteEmote MessageType = "DELETE_EMOTE"

	// Emote Set

	MessageTypeCreateEmoteSet MessageType = "CREATE_EMOTE_SET"
	MessageTypeUpdateEmoteSet MessageType = "UPDATE_EMOTE_SET"
	MessageTypeDeleteEmoteSet MessageType = "DELETE_EMOTE_SET"

	// User

	MessageTypeCreateUser           MessageType = "CREATE_USER"
	MessageTypeUpdateUser           MessageType = "UPDATE_USER"
	MessageTypeDeleteUser           MessageType = "DELETE_USER"
	MessageTypeAddUserConnection    MessageType = "ADD_USER_CONNECTION"
	MessageTypeUpdateUserConnection MessageType = "UPDATE_USER_CONNECTION"
	MessageTypeDeleteUserConnection MessageType = "DELETE_USER_CONNECTION"
)

type Topic uint32

const (
	TopicEmote    Topic = 1 << 0
	TopicEmoteSet Topic = 1 << 1
	TopîcUser     Topic = 1 << 2
)

type EmptyObject = struct{}

type ChangeMap[O EmptyObject | structures.Object] struct {
	// The object's ID
	ID primitive.ObjectID `json:"id"`
	// The type of the object
	Kind structures.ObjectKind `json:"kind"`
	// A list of added fields
	Added []ChangeField `json:"added,omitempty"`
	// A list of updated fields
	Updated []ChangeField `json:"updated,omitempty"`
	// A list of removed fields
	Removed []ChangeField `json:"removed,omitempty"`
	// A full object. Only available during a "create" event
	Object *O `json:"object,omitempty"`
}

type ChangeField struct {
	Key      string `json:"key"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}
