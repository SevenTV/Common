package structures

import (
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntitlementData interface {
	bson.Raw | EntitlementDataBase | EntitlementDataBaseSelectable | EntitlementDataSubscription | EntitlementDataBadge | EntitlementDataPaint | EntitlementDataRole | EntitlementDataEmoteSet
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
	Condition EntitlementCondition `json:"condition" bson:"condition"`
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
		ID:        e.ID,
		Kind:      e.Kind,
		Data:      raw,
		UserID:    e.UserID,
		Disabled:  e.Disabled,
		Condition: e.Condition,
	}
}

func ConvertEntitlement[D EntitlementData](c Entitlement[bson.Raw]) (Entitlement[D], error) {
	var d D
	err := bson.Unmarshal(c.Data, &d)
	c2 := Entitlement[D]{
		ID:        c.ID,
		Kind:      c.Kind,
		Data:      d,
		UserID:    c.UserID,
		Disabled:  c.Disabled,
		Condition: c.Condition,
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
	RefID primitive.ObjectID `json:"-" bson:"ref"`
}

type EntitlementDataBaseSelectable struct {
	// The ID of the subscription
	RefID    primitive.ObjectID `json:"-" bson:"ref"`
	Selected bool               `json:"selected" bson:"selected"`
}

// EntitledSubscription Subscription binding in an Entitlement
type EntitlementDataSubscription struct {
	Interval     int    `json:"interval" bson:"interval"`
	IntervalUnit string `json:"interval_unit" bson:"interval_unit"`
}

// EntitledBadge Badge binding in an Entitlement
type EntitlementDataBadge struct {
	RefID     primitive.ObjectID           `json:"-" bson:"ref"`
	RefObject *Cosmetic[CosmeticDataBadge] `json:"ref_object" bson:"ref_object,skip,omitempty"`
	Selected  bool                         `json:"selected,omitempty" bson:"selected,omitempty"`
}

type EntitlementDataPaint struct {
	RefID     primitive.ObjectID           `json:"-" bson:"ref"`
	RefObject *Cosmetic[CosmeticDataPaint] `json:"ref_object" bson:"ref_object,skip,omitempty"`
	Selected  bool                         `json:"selected,omitempty" bson:"selected,omitempty"`
}

// EntitledRole Role binding in an Entitlement
type EntitlementDataRole struct {
	RefID     primitive.ObjectID `json:"-" bson:"ref"`
	RefObject *Role              `json:"ref_object" bson:"ref_object,skip,omitempty"`
}

// EntitledEmoteSet Emote Set binding in an Entitlement
type EntitlementDataEmoteSet struct {
	RefID     primitive.ObjectID `json:"-" bson:"ref"`
	RefObject *EmoteSet          `json:"ref_object" bson:"ref_object,skip,omitempty"`
}

type EntitlementCondition struct {
	AnyRoles []primitive.ObjectID `json:"any_roles,omitempty" bson:"any_roles,omitempty"`
	AllRoles []primitive.ObjectID `json:"all_roles,omitempty" bson:"all_roles,omitempty"`
	MinDate  time.Time            `json:"min_date,omitempty" bson:"min_date,omitempty"`
	MaxDate  time.Time            `json:"max_date,omitempty" bson:"max_date,omitempty"`
}

func (e EntitlementCondition) IsMet(roleIDs utils.Set[primitive.ObjectID]) bool {
	for _, roleID := range e.AllRoles {
		if !roleIDs.Has(roleID) {
			return false
		}
	}

	for _, roleID := range e.AnyRoles {
		if roleIDs.Has(roleID) {
			return true
		}
	}

	return len(e.AnyRoles) == 0
}

type EntitlementApp struct {
	Name  string         `json:"name"`
	State map[string]any `json:"state"`
}
