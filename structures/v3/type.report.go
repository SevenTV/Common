package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Report struct {
	ID     primitive.ObjectID `json:"id" bson:"_id"`
	CaseID string             `json:"case_id" bson:"case_id"`
	// The type of the target
	TargetKind ObjectKind `json:"target_kind" bson:"target_kind"`
	// The ID of the target
	TargetID primitive.ObjectID `json:"target_id" bson:"target_id"`
	// The ID of the user who created the report
	ActorID primitive.ObjectID `json:"actor_id" bson:"actor_id"`
	// The report subject (i.e "Stolen Emote")
	Subject string `json:"subject" bson:"subject"`
	// The report body (a user-generated text field with details)
	Body string `json:"body" bson:"body"`
	// Priority of the report
	Priority int32 `json:"priority" bson:"priority"`
	// Whether or not the report is open
	Status ReportStatus `json:"status" bson:"status"`
	// The date on which the report was created
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	// The date on which the report was closed
	ClosedAt *time.Time `json:"closed_at,omitempty" bson:"closed_at,omitempty"`
	// The date on which the report was last updated
	LastUpdatedAt time.Time `json:"last_updated_at" bson:"last_updated_at"`
	// The IDs of users assigned to this report
	AssigneeIDs []primitive.ObjectID `json:"assignee_ids" bson:"assignee_ids"`
	// Notes (moderator comments)
	Notes []ReportNote `json:"notes" bson:"notes"`

	// Relational

	Target    *User  `json:"target" bson:"target,skip,omitempty"`
	Actor     *User  `json:"reporter" bson:"actor,skip,omitempty"`
	Assignees []User `json:"assignees" bson:"assignees,skip,omitempty"`
}

type ReportStatus string

const (
	ReportStatusOpen     ReportStatus = "OPEN"
	ReportStatusAssigned ReportStatus = "ASSIGNED"
	ReportStatusClosed   ReportStatus = "CLOSED"
)

type ReportNote struct {
	// The time at which the note was created
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	// The ID of the user who wrote this note
	AuthorID primitive.ObjectID `json:"author_id" bson:"author_id"`
	// The text content of the note
	Content string `json:"content" bson:"content"`
	// If true, the note is only visible to other privileged users
	// it will not be visible to the reporter
	Internal bool `json:"internal" bson:"internal"`
	// Whether or not the note was read by the reporter
	Read bool `json:"read" bson:"read"`
	// A reply to the note by the reporter
	Reply string `json:"reply" bson:"reply"`
}
