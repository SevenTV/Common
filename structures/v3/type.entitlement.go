package structures

import (
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntitlementData interface {
	bson.Raw | EntitlementDataBase | EntitlementDataSubscription | EntitlementDataBadge | EntitlementDataPaint | EntitlementDataRole | EntitlementDataEmoteSet
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
	// Eligibility conditions for this entitlement
	Condition EntitlementCondition `json:"condition,omitempty" bson:"condition,omitempty"`
	// Whether this entitlement is currently inactive
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
	// Information about the app that created this entitlement
	App *EntitlementApp `json:"app,omitempty" bson:"app,omitempty"`
}

func (e Entitlement[D]) ToRaw() Entitlement[bson.Raw] {
	switch x := utils.ToAny(e.Data).(type) {
	case bson.Raw:
		return Entitlement[bson.Raw]{
			ID:       e.ID,
			Kind:     e.Kind,
			Data:     x,
			UserID:   e.UserID,
			Disabled: e.Disabled,
		}
	}

	raw, _ := bson.Marshal(e.Data)
	return Entitlement[bson.Raw]{
		ID:       e.ID,
		Kind:     e.Kind,
		Data:     raw,
		UserID:   e.UserID,
		Disabled: e.Disabled,
	}
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

type EntitlementDataBase struct {
	// The ID of the subscription
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// EntitledSubscription Subscription binding in an Entitlement
type EntitlementDataSubscription struct {
	// The ID of the subscription
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// EntitledBadge Badge binding in an Entitlement
type EntitlementDataBadge struct {
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
	// DEPRECATED: use Entitlement.Condition
	RoleBinding *primitive.ObjectID `json:"role_binding,omitempty" bson:"role_binding,omitempty"`
	Selected    bool                `json:"selected,omitempty" bson:"selected,omitempty"`
}

type EntitlementDataPaint struct {
	ObjectReference primitive.ObjectID  `json:"-" bson:"ref"`
	RoleBinding     *primitive.ObjectID `json:"role_binding,omitempty" bson:"role_binding,omitempty"`
	Selected        bool                `json:"selected,omitempty" bson:"selected,omitempty"`
}

// EntitledRole Role binding in an Entitlement
type EntitlementDataRole struct {
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// EntitledEmoteSet Emote Set binding in an Entitlement
type EntitlementDataEmoteSet struct {
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

type EntitlementCondition struct {
	AnyRoles []primitive.ObjectID `json:"any_roles,omitempty" bson:"any_roles,omitempty"`
	AllRoles []primitive.ObjectID `json:"all_roles,omitempty" bson:"all_roles,omitempty"`
	MinDate  time.Time            `json:"min_date,omitempty" bson:"min_date,omitempty"`
	MaxDate  time.Time            `json:"max_date,omitempty" bson:"max_date,omitempty"`
}

type EntitlementApp struct {
	Name  string         `json:"name"`
	State map[string]any `json:"state"`
}
