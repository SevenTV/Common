package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Meta struct {
	Announcement      string   `json:"announcement"`
	FeaturedBroadcast string   `json:"featured_broadcast"`
	Roles             []string `json:"roles"`
}

type Broadcast struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	ThumbnailURL string   `json:"thumbnail_url"`
	ViewerCount  int32    `json:"viewer_count"`
	Type         string   `json:"type"`
	GameName     string   `json:"game_name"`
	GameID       string   `json:"game_id"`
	Language     string   `json:"language"`
	Tags         []string `json:"tags"`
	Mature       bool     `json:"mature"`
	StartedAt    string   `json:"started_at"`
	UserID       string   `json:"user_id"`
}

type Notification struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Announcement bool               `json:"announcement" bson:"announcement"` // If true, the notification is global and visible to all users regardless of targets

	Title        string                    `json:"title" bson:"title"`                 // The notification's heading / title
	MessageParts []NotificationMessagePart `json:"message_parts" bson:"message_parts"` // The parts making up the notification's formatted message

	Read   bool      `json:"read" bson:"read,omitempty"`
	ReadAt time.Time `json:"read_at" bson:"read_at,omitempty"`
	Users  []*User   `json:"users" bson:"-"`  // The users mentioned in this notification
	Emotes []*Emote  `json:"emotes" bson:"-"` // The emotesm entioned in this notification
}

type NotificationMessagePart struct {
	Type NotificationContentMessagePartType `json:"part_type" bson:"part_type"` // The type of this part

	Text    *string             `json:"text" bson:"text"`
	Mention *primitive.ObjectID `json:"mention" bson:"mention"`
}

type NotificationReadState struct {
	TargetUser   primitive.ObjectID `json:"target" bson:"target"`                // The user targeted to see the notification
	Notification primitive.ObjectID `json:"notification_id" bson:"notification"` // The notification that can be read
	Read         bool               `json:"read" bson:"read"`                    // Whether the user read the notification
	ReadAt       *time.Time         `json:"read_at" bson:"read_at"`              // When the notification was read
}

const (
	NotificationMessagePartTypeText NotificationContentMessagePartType = 1 + iota
	NotificationMessagePartTypeUserMention
	NotificationMessagePartTypeEmoteMention
	NotificationMessagePartTypeRoleMention
)

type NotificationContentMessagePartType int8
