package structures

import (
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserPresence[T UserPresenceData] struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	// Authentic is whether or not this presence was authenticated; confirmed that the actor issued this presence
	//
	// If false, such a presence cannot be trusted as truth of a user's location.
	// This may be used for non-sensitive information delivery, for example
	// to allow a client to announce a user as active in a chat room, but not require
	// a full authentication flow to attain this functionality.
	Authentic bool `json:"authentic" bson:"authentic"`
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

func (upk UserPresenceKind) String() string {
	switch upk {
	case UserPresenceKindChannel:
		return "channel"
	case UserPresenceKindWebPage:
		return "web_page"
	default:
		return "unknown"
	}
}

type UserPresenceData interface {
	bson.Raw | UserPresenceDataChannel | UserPresenceLocationDataWebPage
}

type UserPresenceDataChannel struct {
	HostID       primitive.ObjectID `json:"host_id" bson:"host_id"`
	ConnectionID string             `json:"connection_id" bson:"connection_id"`
}

type UserPresenceLocationDataWebPage struct {
	URL       string `json:"url" bson:"url"`
	UserAgent string `json:"user_agent" bson:"user_agent"`
	IP        []byte `json:"ip" bson:"ip"`
}

func (up UserPresence[T]) ToRaw() UserPresence[bson.Raw] {
	switch x := utils.ToAny(up.Data).(type) {
	case bson.Raw:
		return UserPresence[bson.Raw]{
			ID:        up.ID,
			UserID:    up.UserID,
			Authentic: up.Authentic,
			Timestamp: up.Timestamp,
			TTL:       up.TTL,
			Kind:      up.Kind,
			Data:      x,
		}
	}

	raw, _ := bson.Marshal(up.Data)

	return UserPresence[bson.Raw]{
		ID:        up.ID,
		UserID:    up.UserID,
		Authentic: up.Authentic,
		Timestamp: up.Timestamp,
		TTL:       up.TTL,
		Kind:      up.Kind,
		Data:      raw,
	}
}

func ConvertPresence[D UserPresenceData](presence UserPresence[bson.Raw]) (UserPresence[D], error) {
	var d D

	err := bson.Unmarshal(presence.Data, &d)

	up := UserPresence[D]{
		ID:        presence.ID,
		UserID:    presence.UserID,
		Authentic: presence.Authentic,
		Timestamp: presence.Timestamp,
		TTL:       presence.TTL,
		Kind:      presence.Kind,
		Data:      d,
	}

	return up, err
}
