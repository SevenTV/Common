package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Activity struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Metadata  ActivityState      `json:"metadata" bson:"metadata"`
	Values    ActivityValues     `json:"values" bson:"values"`
}

type ActivityState struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
}

type ActivityValues struct {
	Type   ActivityType    `json:"type" bson:"type,omitempty"`
	Name   ActivityName    `json:"name" bson:"name,omitempty"`
	Status ActivityStatus  `json:"status" bson:"status"`
	Object *ActivityObject `json:"object" bson:"object,omitempty"`
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
