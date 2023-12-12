package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntitlementBuilder[D EntitlementData] struct {
	Entitlement Entitlement[D]

	User    *User
	initial Entitlement[D]
}

func NewEntitlementBuilder[D EntitlementData](ent Entitlement[D]) *EntitlementBuilder[D] {
	return &EntitlementBuilder[D]{
		Entitlement: ent,
		initial:     ent,
	}
}

// SetKind: Change the entitlement's kind
func (b *EntitlementBuilder[D]) SetKind(kind EntitlementKind) *EntitlementBuilder[D] {
	b.Entitlement.Kind = kind

	return b
}

// SetUserID: Change the entitlement's assigned user
func (b *EntitlementBuilder[D]) SetUserID(id primitive.ObjectID) *EntitlementBuilder[D] {
	b.Entitlement.UserID = id

	return b
}

// SetSubscriptionData: Add a subscription reference to the entitlement
func (b *EntitlementBuilder[D]) SetData(data D) *EntitlementBuilder[D] {
	b.Entitlement.Data = data
	return b
}

func (b *EntitlementBuilder[D]) SetCondition(cond EntitlementCondition) *EntitlementBuilder[D] {
	b.Entitlement.Condition = cond
	return b
}

func (b *EntitlementBuilder[D]) SetApp(app EntitlementApp) *EntitlementBuilder[D] {
	b.Entitlement.App = &app
	return b
}

func (b *EntitlementBuilder[D]) SetClaim(claim EntitlementClaim) *EntitlementBuilder[D] {
	b.Entitlement.Claim = &claim
	return b
}
