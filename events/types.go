package events

import (
	"encoding/json"
	"strings"

	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventType string

const (
	// System

	EventTypeAnySystem          EventType = "system.*"
	EventTypeSystemAnnouncement EventType = "system.announcement"

	// Emote

	EventTypeAnyEmote    EventType = "emote.*"
	EventTypeCreateEmote EventType = "emote.create"
	EventTypeUpdateEmote EventType = "emote.update"
	EventTypeDeleteEmote EventType = "emote.delete"

	// Emote Set

	EventTypeAnyEmoteSet    EventType = "emote_set.*"
	EventTypeCreateEmoteSet EventType = "emote_set.create"
	EventTypeUpdateEmoteSet EventType = "emote_set.update"
	EventTypeDeleteEmoteSet EventType = "emote_set.delete"

	// User

	EventTypeAnyUser              EventType = "user.*"
	EventTypeCreateUser           EventType = "user.create"
	EventTypeUpdateUser           EventType = "user.update"
	EventTypeDeleteUser           EventType = "user.delete"
	EventTypeAddUserConnection    EventType = "user.add_connection"
	EventTypeUpdateUserConnection EventType = "user.update_connection"
	EventTypeDeleteUserConnection EventType = "user.delete_connection"
)

func (et EventType) Split() []string {
	a := strings.Split(string(et), ".")
	if len(a) == 0 {
		return []string{"any", "*"}
	}
	return a
}

func (et EventType) ObjectName() string {
	return et.Split()[0]
}

type EmptyObject = struct{}

type ChangeMap struct {
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
	Object json.RawMessage `json:"object,omitempty"`
}

type ChangeField struct {
	Key      string `json:"key"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

type SessionMutation struct {
	Ref       string                 `json:"ref"`
	SessionID string                 `json:"session_id"`
	Events    []SessionMutationEvent `json:"events,omitempty"`
	ActorID   primitive.ObjectID     `json:"actor_id,omitempty"`
}

type SessionMutationEvent struct {
	Action  structures.ListItemAction `json:"action"`
	Type    EventType                 `json:"type"`
	Targets []string                  `json:"targets"`
}
