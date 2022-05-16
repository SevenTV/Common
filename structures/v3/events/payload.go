package events

import (
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HelloPayload struct {
	HeartbeatInterval int64               `json:"heartbeat_interval"`
	Actor             *primitive.ObjectID `json:"actor,omitempty"`
}

type HeartbeatPayload struct {
	Count int64 `json:"count"`
}

type DispatchPayload[B EmptyObject | structures.Object] struct {
	Type MessageType  `json:"type"`
	Body ChangeMap[B] `json:"body"`
}

type SignalPayload struct {
	Sender SignalUser `json:"sender"`
	Host   SignalUser `json:"host"`
}

type SignalUser struct {
	ID          primitive.ObjectID `json:"id"`
	ChannelID   string             `json:"channel_id"`
	Username    string             `json:"username"`
	DisplayName string             `json:"display_name"`
}
