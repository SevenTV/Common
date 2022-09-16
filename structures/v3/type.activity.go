package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Activity struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	State     ActivityState      `json:"state" bson:"state"`
	Type      ActivityType       `json:"type" bson:"type,omitempty"`
	Name      ActivityName       `json:"name" bson:"name,omitempty"`
	Status    ActivityStatus     `json:"status" bson:"status"`
	Object    *ActivityObject    `json:"object" bson:"object,omitempty"`
}

type ActivityState struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	Timespan ActivityTimespan   `json:"timespan" bson:"timespan"`
}

type ActivityObject struct {
	Kind ObjectKind         `json:"kind" bson:"kind"`
	ID   primitive.ObjectID `json:"id" bson:"id"`
}

type ActivityName string

type ActivityType uint8

const (
	ActivityTypeViewing ActivityType = iota + 1
	ActivityTypeEditing
	ActivityTypeWatching
	ActivityTypeListening
	ActivityTypeChatting
	ActivityTypeCreating
	ActivityTypeUpdating
)

type ActivityStatus uint8

const (
	ActivityStatusOffline ActivityStatus = iota
	ActivityStatusIdle
	ActivityStatusDnd
	ActivityStatusOnline
)

type ActivityTimespan struct {
	Start time.Time  `json:"start" bson:"start"`
	End   *time.Time `json:"end" bson:"end"`
}
