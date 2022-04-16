package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

type EntitlementBuilder[D EntitlementData] struct {
	Entitlement Entitlement[D]

	User *User
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
