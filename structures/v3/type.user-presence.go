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
	// IP is the IP address of the client who initiated this presence
	IP string `json:"ip" bson:"ip"`
	// Authentic is whether or not this presence was authenticated; confirmed that the actor issued this presence
	//
	// If false, such a presence cannot be trusted as truth of a user's location.
	// This may be used for non-sensitive information delivery, for example
	// to allow a client to announce a user as active in a chat room, but not require
	// a full authentication flow to attain this functionality.
	Authentic bool `json:"authentic" bson:"authentic"`
	// Known is whether or not the data in this presence is known
	//
	// If false, it means the data passed wasn't validated to match a real location.
	Known bool `json:"known" bson:"known"`
	// Timestamp is the time at which this presence was issued
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	// TTL is how long this presence is valid for, before it expires
	TTL time.Time `json:"ttl" bson:"ttl"`
	// Kind is the type of presence this is
	Kind         UserPresenceKind          `json:"kind" bson:"kind"`
	Data         T                         `json:"data" bson:"data"`
	Entitlements []UserPresenceEntitlement `json:"entitlements,omitempty" bson:"entitlements,omitempty"`
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
	Platform UserConnectionPlatform        `json:"platform" bson:"platform"`
	ID       string                        `json:"id" bson:"id"`
	Filter   UserPresenceDataChannelFilter `json:"filter" bson:"-"`
}

type UserPresenceDataChannelFilter struct {
	Emotes []string `json:"emotes" bson:"-"`
}

type UserPresenceEntitlement struct {
	Kind         EntitlementKind    `json:"kind" bson:"kind"`
	ID           primitive.ObjectID `json:"id" bson:"id"`
	RefID        primitive.ObjectID `json:"ref" bson:"ref"`
	DispatchHash uint32             `json:"dispatch_hash,omitempty" bson:"dispatch_hash,omitempty"`
}

type UserPresenceLocationDataWebPage struct {
	URL       string `json:"url" bson:"url"`
	UserAgent string `json:"user_agent" bson:"user_agent"`
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
