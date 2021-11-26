package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Ban struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`

	// The user who is affected by this ban
	VictimID primitive.ObjectID `json:"victim_id" bson:"victim_id"`
	// The user who created this ban
	ActorID primitive.ObjectID `json:"actor_id" bson:"actor_id"`
	// The reason for the ban
	Reason string `json:"reason" bson:"reason"`
	// The time at which the ban will expire
	ExpireAt time.Time `json:"expire_at" bson:"expire_at"`
	// The effects that this ban will have
	Effects []BanEffect `json:"effects" bson:"effects"`

	// Relational

	Victim *User `json:"victim" bson:"victim,skip,omitempty"`
	Actor  *User `json:"actor" bson:"actor,skip,omitempty"`
}

type BanEffect string

const (
	// Strip the banned user of all permissions
	BanEffectNoPermissions BanEffect = "NO_PERMISSIONS"
	// Prevents the banned user from authenticating
	BanEffectNoAuth BanEffect = "NO_AUTH"
	// Any object owned by the banned user will no longer be returned by the API
	BanEffectNoOwnership BanEffect = "NO_OWNERSHIP"
	// The banned user is never returned by the API to non-privileged users
	BanEffectMemoryHole BanEffect = "MEMORY_HOLE"
	// The banned user's IP will be blocked from accessing all services
	BanEffectBlockedIP BanEffect = "IP_BLOCKED"
)

func (b *Ban) HasEffect(eff BanEffect) bool {
	for _, e := range b.Effects {
		if e == eff {
			return true
		}
	}
	return false
}

type BanBuilder struct {
	Ban    *Ban
	Update UpdateMap
}

// NewRoleBuilder: create a new role builder
func NewBanBuilder(ban *Ban) *BanBuilder {
	return &BanBuilder{
		Update: UpdateMap{},
		Ban:    ban,
	}
}

func (bb *BanBuilder) SetVictimID(id primitive.ObjectID) *BanBuilder {
	bb.Ban.VictimID = id
	bb.Update.Set("victim_id", id)
	return bb
}

func (bb *BanBuilder) SetActorID(id primitive.ObjectID) *BanBuilder {
	bb.Ban.ActorID = id
	bb.Update.Set("actor_id", id)
	return bb
}

func (bb *BanBuilder) SetReason(s string) *BanBuilder {
	bb.Ban.Reason = s
	bb.Update.Set("reason", s)
	return bb
}

func (bb *BanBuilder) SetExpireAt(t time.Time) *BanBuilder {
	bb.Ban.ExpireAt = t
	bb.Update.Set("expire_at", t)
	return bb
}

func (bb *BanBuilder) SetEffects(a []BanEffect) *BanBuilder {
	bb.Ban.Effects = a
	bb.Update.Set("effects", a)
	return bb
}
