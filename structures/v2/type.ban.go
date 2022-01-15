package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Ban struct {
	ID         primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	UserID     *primitive.ObjectID `json:"user_id" bson:"user_id"`
	Reason     string              `json:"reason" bson:"reason"`
	IssuedByID *primitive.ObjectID `json:"issued_by_id" bson:"issued_by_id"`
	ExpireAt   time.Time           `json:"expire_at" bson:"expire_at"`
}
