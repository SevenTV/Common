package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserPresence[T UserPresenceData] struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ActorID primitive.ObjectID `json:"actor_id" bson:"actor_id"`
	// Authentic is whether or not this presence was authenticated; confirmed that the actor issued this presence
	//
	// If false, such a presence cannot be trusted as truth of a user's location.
	// This may be used for non-sensitive information delivery, for example
	// to allow a client to announce a user as active in a chat room, but not require
	// a full authentication flow to attain this functionality.
	Authentic bool `json:"authenticated" bson:"authenticated"`
	// Timestamp is the time at which this presence was issued
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	// TTL is how long this presence is valid for, before it expires
	TTL time.Time `json:"ttl" bson:"ttl"`
	// Kind is the type of presence this is
	Kind UserPresenceKind `json:"kind" bson:"kind"`
	Data T                `json:"data" bson:"data"`
}

type UserPresenceKind uint8

const (
	UserPresenceKindUnknown UserPresenceKind = iota
	UserPresenceKindChannel
	UserPresenceKindWebPage
)

type UserPresenceData interface {
	bson.Raw | UserPresenceLocationDataChannel
}

type UserPresenceLocationDataChannel struct {
	HostID       primitive.ObjectID `json:"host_id" bson:"host_id"`
	ConnectionID primitive.ObjectID `json:"connection_id" bson:"connection_id"`
}

type UserPresenceLocationDataWebPage struct {
	URL       string `json:"url" bson:"url"`
	UserAgent string `json:"user_agent" bson:"user_agent"`
	IP        []byte `json:"ip" bson:"ip"`
}
