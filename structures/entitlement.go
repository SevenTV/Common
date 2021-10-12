package structures

import (
	"context"
	"fmt"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/configure"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EntitlementBuilder struct {
	Entitlement Entitlement

	User *User
}

func (b EntitlementBuilder) Write(ctx context.Context, inst mongo.Instance) (EntitlementBuilder, error) {
	// Create new Object ID if this is a new entitlement
	if b.Entitlement.ID.IsZero() {
		b.Entitlement.ID = primitive.NewObjectID()
	}

	if _, err := inst.Collection(mongo.CollectionNameEntitlements).UpdateByID(ctx, b.Entitlement.ID, bson.M{
		"$set": b.Entitlement,
	}, &options.UpdateOptions{
		Upsert: utils.BoolPointer(true),
	}); err != nil {
		logrus.WithError(err).Error("mongo")
		return b, err
	}

	return b, nil
}

// GetUser: Fetch the user data from the user ID assigned to the entitlement
func (b EntitlementBuilder) GetUser(ctx context.Context, inst mongo.Instance) (*UserBuilder, error) {
	if b.Entitlement.UserID.IsZero() {
		return nil, fmt.Errorf("Entitlement does not have a user assigned")
	}

	ub, err := UserBuilder{}.FetchByID(ctx, inst, b.Entitlement.UserID)
	if err != nil {
		return nil, err
	}

	// role := datastructure.GetRole(ub.User.RoleID)
	// ub.User.Role = &role
	return ub, nil
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
func (b EntitlementBuilder) ReadSubscriptionData() EntitledSubscription {
	var e EntitledSubscription
	if err := bson.Unmarshal(b.Entitlement.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return e
	}
	return e
}

// ReadBadgeData: Read the data as an Entitled Badge
func (b EntitlementBuilder) ReadBadgeData() EntitledBadge {
	var e EntitledBadge
	if err := bson.Unmarshal(b.Entitlement.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return e
	}
	return e
}

// ReadRoleData: Read the data as an Entitled Role
func (b EntitlementBuilder) ReadRoleData() EntitledRole {
	var e EntitledRole
	if err := bson.Unmarshal(b.Entitlement.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return e
	}
	return e
}

// ReadEmoteSetData: Read the data as an Entitled Emote Set
func (b EntitlementBuilder) ReadEmoteSetData() EntitledEmoteSet {
	var e EntitledEmoteSet
	if err := bson.Unmarshal(b.Entitlement.Data, &e); err != nil {
		logrus.WithError(err).Error("bson")
		return e
	}
	return e
}

// FetchEntitlements: gets entitlement of specified kind
func FetchEntitlements(ctx context.Context, inst mongo.Instance, opts struct {
	Kind            *EntitlementKind
	ObjectReference primitive.ObjectID
}) ([]EntitlementBuilder, error) {
	// Make a request to get the user's entitlements
	var entitlements []*entitlementWithUser
	query := bson.M{
		"kind":     opts.Kind,
		"disabled": bson.M{"$not": bson.M{"$eq": true}},
	}
	if !opts.ObjectReference.IsZero() {
		query["data.ref"] = opts.ObjectReference
	}

	pipeline := mongo.Pipeline{
		bson.D{bson.E{
			Key:   "$match",
			Value: query,
		}},
		bson.D{bson.E{
			Key:   "$addFields",
			Value: bson.M{"entitlement": "$$ROOT"},
		}},
		bson.D{bson.E{
			Key: "$lookup",
			Value: bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		}},
		bson.D{bson.E{
			Key:   "$unwind",
			Value: "$user",
		}},
	}

	cur, err := inst.Collection(configure.CollectionNameEntitlements).Aggregate(ctx, pipeline)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		logrus.WithError(err).Error("actions, UserBuilder, FetchEntitlements")
		return nil, err
	}

	// Get all entitlements
	if err := cur.All(ctx, &entitlements); err != nil {
		return nil, err
	}

	// Wrap into Entitlement Builders
	builders := make([]EntitlementBuilder, len(entitlements))
	for i, e := range entitlements {
		builders[i] = EntitlementBuilder{
			Entitlement: *e.Entitlement,
			User:        e.User,
		}
	}

	return builders, nil
}

type entitlementWithUser struct {
	*Entitlement
	User *User
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

// A string representing an Entitlement Kind
type EntitlementKind string

var (
	EntitlementKindSubscription = EntitlementKind("SUBSCRIPTION") // Subscription Entitlement
	EntitlementKindBadge        = EntitlementKind("BADGE")        // Badge Entitlement
	EntitlementKindRole         = EntitlementKind("ROLE")         // Role Entitlement
	EntitlementKindEmoteSet     = EntitlementKind("EMOTE_SET")    // Emote Set Entitlement
)

// (Data) Subscription binding in an Entitlement
type EntitledSubscription struct {
	ID string `json:"id" bson:"-"`
	// The ID of the subscription
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// (Data) Badge binding in an Entitlement
type EntitledBadge struct {
	ID              string             `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
	Selected        bool               `json:"selected" bson:"selected"`
	// The role required for the badge to show up
	RoleBinding   *primitive.ObjectID `json:"role_binding" bson:"role_binding"`
	RoleBindingID *string             `json:"role_binding_id" bson:"-"`
}

// (Data) Role binding in an Entitlement
type EntitledRole struct {
	ID              string             `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID `json:"-" bson:"ref"`
}

// (Data) Emote Set binding in an Entitlement
type EntitledEmoteSet struct {
	ID              string               `json:"id" bson:"-"`
	ObjectReference primitive.ObjectID   `json:"-" bson:"ref"`
	UnicodeTag      string               `json:"unicode_tag" bson:"unicode_tag"`
	EmoteIDs        []primitive.ObjectID `json:"emote_ids" bson:"emotes"`

	// Relational

	// A list of emotes for this emote set entitlement
	Emotes []*Emote `json:"emotes" bson:"-"`
}
