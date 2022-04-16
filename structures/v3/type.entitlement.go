package structures

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntitlementData interface {
	bson.Raw | EntitlementDataSubscription | EntitlementDataBadge | EntitlementDataRole | EntitlementDataSet
}

// Entitlement is a binding between a resource and a user
// It grants the user access to the bound resource
// and may define some additional properties on top.
type Entitlement[D EntitlementData] struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// Kind represents what item this entitlement grants
	Kind EntitlementKind `json:"kind" bson:"kind"`
	// Data referencing the entitled item
	Data D `json:"data" bson:"data"`
	// The user who is entitled to the item
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	// Wether this entitlement is currently inactive
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
}

func ConvertEntitlement[D EntitlementData](c Entitlement[bson.Raw]) (Entitlement[D], error) {
	var d D
	err := bson.Unmarshal(c.Data, &d)
	c2 := Entitlement[D]{
		ID:       c.ID,
		Kind:     c.Kind,
		Data:     d,
		UserID:   c.UserID,
		Disabled: c.Disabled,
	}

	return c2, err
}

// EntitlementKind A string representing a kind of entitlement
type EntitlementKind string

const (
	EntitlementKindSubscription = EntitlementKind("SUBSCRIPTION") // Subscription Entitlement
	EntitlementKindBadge        = EntitlementKind("BADGE")        // Badge Entitlement
	EntitlementKindPaint        = EntitlementKind("PAINT")        // Badge Entitlement
	EntitlementKindRole         = EntitlementKind("ROLE")         // Role Entitlement
	EntitlementKindEmoteSet     = EntitlementKind("EMOTE_SET")    // Emote Set Entitlement
)

// EntitledSubscription Subscription binding in an Entitlement
type EntitlementDataSubscription struct {
	ID string `json:"id" bson:"-"`
	// The ID of the subscription
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// EntitledBadge Badge binding in an Entitlement
type EntitlementDataBadge struct {
	ID              string             `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
	// The role required for the badge to show up
	RoleBindingID *string             `json:"role_binding_id" bson:"-"`
	RoleBinding   *primitive.ObjectID `json:"role_binding,omitempty" bson:"role_binding,omitempty"`
	Selected      bool                `json:"selected,omitempty" bson:"selected,omitempty"`
}

// EntitledRole Role binding in an Entitlement
type EntitlementDataRole struct {
	ID              string             `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// EntitledEmoteSet Emote Set binding in an Entitlement
type EntitlementDataSet struct {
	ID              string             `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}
