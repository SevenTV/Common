package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Report struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	// The type of the target
	TargetKind ReportTargetKind `json:"target_kind" bson:"target_kind"`
	// The ID of the target
	TargetID primitive.ObjectID `json:"target_id" bson:"target_id"`
	// The ID of the user who created the report
	ReporterID primitive.ObjectID `json:"reporter_id" bson:"reporter_id"`
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
	// The IDs of users assigned to this report
	AssigneeIDs []primitive.ObjectID `json:"assignee_ids" bson:"assignee_ids"`
	// Notes (moderator comments)
	Notes []*ReportNote `json:"notes" bson:"notes"`

	// Relational

	Target    *User   `json:"target" bson:"target,skip,omitempty"`
	Reporter  *User   `json:"reporter" bson:"reporter,skip,omitempty"`
	Assignees []*User `json:"assignees" bson:"assignees,skip,omitempty"`
}

// The type of object being reported
type ReportTargetKind string

const (
	ReportTargetKindEmote ReportTargetKind = "EMOTE"
	ReportTargetKindUser  ReportTargetKind = "USER"
)

type ReportStatus string

const (
	ReportStatusOpen     ReportStatus = "OPEN"
	ReportStatusAssigned ReportStatus = "ASSIGNED"
	ReportStatusClosed   ReportStatus = "CLOSED"
)

type ReportBuilder struct {
	Update UpdateMap
	Report *Report
}

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

// NewReportBuilder: create a new report builder
func NewReportBuilder(report *Report) *ReportBuilder {
	return &ReportBuilder{
		Update: map[string]interface{}{},
		Report: report,
	}
}

func (rb *ReportBuilder) SetTargetKind(kind ReportTargetKind) *ReportBuilder {
	rb.Report.TargetKind = kind
	rb.Update.Set("target_kind", kind)
	return rb
}

func (rb *ReportBuilder) SetTargetID(id primitive.ObjectID) *ReportBuilder {
	rb.Report.TargetID = id
	rb.Update.Set("target_id", id)
	return rb
}

func (rb *ReportBuilder) SetReporterID(id primitive.ObjectID) *ReportBuilder {
	rb.Report.ReporterID = id
	rb.Update.Set("reporter_id", id)
	return rb
}

func (rb *ReportBuilder) SetSubject(subject string) *ReportBuilder {
	rb.Report.Subject = subject
	rb.Update.Set("subject", subject)
	return rb
}

func (rb *ReportBuilder) SetBody(body string) *ReportBuilder {
	rb.Report.Body = body
	rb.Update.Set("body", body)
	return rb
}

func (rb *ReportBuilder) SetCreatedAt(t time.Time) *ReportBuilder {
	rb.Report.CreatedAt = t
	rb.Update.Set("created_at", t)
	return rb
}

func (rb *ReportBuilder) SetPriority(p int32) *ReportBuilder {
	rb.Report.Priority = p
	rb.Update.Set("priority", p)
	return rb
}

func (rb *ReportBuilder) SetStatus(s ReportStatus) *ReportBuilder {
	rb.Report.Status = s
	rb.Update.Set("status", s)
	return rb
}

func (rb *ReportBuilder) AddAssignee(id primitive.ObjectID) *ReportBuilder {
	rb.Report.AssigneeIDs = append(rb.Report.AssigneeIDs, id)
	rb.Update.AddToSet("assignee_ids", id)
	return rb
}

func (rb *ReportBuilder) RemoveAssignee(id primitive.ObjectID) *ReportBuilder {
	if len(rb.Report.AssigneeIDs) == 0 {
		return rb
	}

	ind := 0
	for i, a := range rb.Report.AssigneeIDs {
		if a == id {
			ind = i
			break
		}
	}

	rb.Report.AssigneeIDs = append(rb.Report.AssigneeIDs[:ind], rb.Report.AssigneeIDs[ind+1:]...)
	rb.Update.Pull("assignee_ids", id)
	return rb
}

func (rb *ReportBuilder) AddNote(note *ReportNote) *ReportBuilder {
	rb.Report.Notes = append(rb.Report.Notes, note)
	rb.Update.AddToSet("notes", note)
	return rb
}
