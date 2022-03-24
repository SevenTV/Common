package structures

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntitlementBuilder struct {
	Entitlement Entitlement

	User *User
}

// SetKind: Change the entitlement's kind
func (b EntitlementBuilder) SetKind(kind EntitlementKind) EntitlementBuilder {
	b.Entitlement.Kind = kind

	return b
}

// SetUserID: Change the entitlement's assigned user
func (b EntitlementBuilder) SetUserID(id primitive.ObjectID) EntitlementBuilder {
	b.Entitlement.UserID = id

	return b
}

// SetSubscriptionData: Add a subscription reference to the entitlement
func (b EntitlementBuilder) SetSubscriptionData(data EntitledSubscription) EntitlementBuilder {
	return b.marshalData(data)
}

// SetBadgeData: Add a badge reference to the entitlement
func (b EntitlementBuilder) SetBadgeData(data EntitledBadge) EntitlementBuilder {
	return b.marshalData(data)
}

// SetRoleData: Add a role reference to the entitlement
func (b EntitlementBuilder) SetRoleData(data EntitledRole) EntitlementBuilder {
	return b.marshalData(data)
}

// SetEmoteSetData: Add an emote set reference to the entitlement
func (b EntitlementBuilder) SetEmoteSetData(data EntitledEmoteSet) EntitlementBuilder {
	return b.marshalData(data)
}

func (b EntitlementBuilder) marshalData(data interface{}) EntitlementBuilder {
	d, err := bson.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("bson")
		return b
	}

	b.Entitlement.Data = d
	return b
}

// ReadSubscriptionData: Read the data as an Entitled Subscription
func (e *Entitlement) ReadSubscriptionData() EntitledSubscription {
	var d EntitledSubscription
	if err := bson.Unmarshal(e.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return d
	}
	return d
}

// ReadBadgeData: Read the data as an Entitled Badge
func (e *Entitlement) ReadBadgeData() EntitledBadge {
	var d EntitledBadge
	if err := bson.Unmarshal(e.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return d
	}
	return d
}

// ReadRoleData: Read the data as an Entitled Role
func (e *Entitlement) ReadRoleData() EntitledRole {
	var d EntitledRole
	if err := bson.Unmarshal(e.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return d
	}
	return d
}

// ReadEmoteSetData: Read the data as an Entitled Emote Set
func (e *Entitlement) ReadEmoteSetData() EntitledEmoteSet {
	var d EntitledEmoteSet
	if err := bson.Unmarshal(e.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return d
	}
	return d
}

func (b EntitlementBuilder) Log(str string) {
	logrus.WithFields(logrus.Fields{
		"id":      b.Entitlement.ID,
		"kind":    b.Entitlement.Kind,
		"user_id": b.Entitlement.UserID,
	}).Info(str)
}

func (b EntitlementBuilder) LogError(str string) {
	logrus.WithFields(logrus.Fields{
		"id":      b.Entitlement.ID,
		"kind":    b.Entitlement.Kind,
		"user_id": b.Entitlement.UserID,
	}).Error(str)
}

// Entitlement is a binding between a resource and a user
// It grants the user access to the bound resource
// and may define some additional properties on top.
type Entitlement struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// Kind represents what item this entitlement grants
	Kind EntitlementKind `json:"kind" bson:"kind"`
	// Data referencing the entitled item
	Data bson.Raw `json:"data" bson:"data"`
	// The user who is entitled to the item
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	// Wether this entitlement is currently inactive
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
}

func (e *Entitlement) GetData() EntitlementData {
	return EntitlementData(e.Data)
}

// EntitlementKind A string representing a kind of entitlement
type EntitlementKind string

var (
	EntitlementKindSubscription = EntitlementKind("SUBSCRIPTION") // Subscription Entitlement
	EntitlementKindBadge        = EntitlementKind("BADGE")        // Badge Entitlement
	EntitlementKindPaint        = EntitlementKind("PAINT")        // Badge Entitlement
	EntitlementKindRole         = EntitlementKind("ROLE")         // Role Entitlement
	EntitlementKindEmoteSet     = EntitlementKind("EMOTE_SET")    // Emote Set Entitlement
)

type EntitledItem struct {
	ObjectReference primitive.ObjectID  `json:"-" bson:"ref"`
	RoleBinding     *primitive.ObjectID `json:"role_binding,omitempty" bson:"role_binding,omitempty"`
	Selected        bool                `json:"selected,omitempty" bson:"selected,omitempty"`
}

// EntitledSubscription Subscription binding in an Entitlement
type EntitledSubscription struct {
	ID string `json:"id" bson:"-"`
	// The ID of the subscription
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// EntitledBadge Badge binding in an Entitlement
type EntitledBadge struct {
	ID string `json:"id" bson:"-"`
	*EntitledItem
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
	// The role required for the badge to show up
	RoleBindingID *string `json:"role_binding_id" bson:"-"`
}

// EntitledRole Role binding in an Entitlement
type EntitledRole struct {
	ID              string             `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// EntitledEmoteSet Emote Set binding in an Entitlement
type EntitledEmoteSet struct {
	ID              string             `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

type EntitlementData bson.Raw

func (d EntitlementData) ReadItem() *EntitledItem {
	return d.unmarshal(&EntitledItem{}).(*EntitledItem)
}

func (d EntitlementData) ReadRole() *EntitledRole {
	return d.unmarshal(&EntitledRole{}).(*EntitledRole)
}

func (d EntitlementData) unmarshal(i interface{}) interface{} {
	if err := bson.Unmarshal(d, i); err != nil {
		logrus.WithError(err).Error("message, decoding message data failed")
	}
	return i
}
