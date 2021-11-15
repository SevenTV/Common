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

	// Relational

	Target   *User `json:"target" bson:"target,skip"`
	Reporter *User `json:"reporter" bson:"reporter,skip"`
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
