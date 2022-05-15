package events

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessagePayload interface {
	json.RawMessage | HelloPayload | DispatchPayload
}

type HelloPayload struct {
	HeartbeatInterval int64               `json:"heartbeat_interval"`
	Actor             *primitive.ObjectID `json:"actor,omitempty"`
}

type DispatchPayload struct {
	Type MessageType            `json:"type"`
	Body ChangeMap[EmptyObject] `json:"body"`
}
