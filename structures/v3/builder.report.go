package structures

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportBuilder struct {
	Update UpdateMap
	Report Report
}

// NewReportBuilder: create a new report builder
func NewReportBuilder(report Report) *ReportBuilder {
	return &ReportBuilder{
		Update: map[string]interface{}{},
		Report: report,
	}
}

func (rb *ReportBuilder) SetTargetKind(kind ObjectKind) *ReportBuilder {
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
	rb.Report.ActorID = id
	rb.Update.Set("actor_id", id)
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
	if len(rb.Report.AssigneeIDs) == 0 {
		rb.Update.Set("assignee_ids", []primitive.ObjectID{id})
	} else {
		rb.Update.AddToSet("assignee_ids", id)
	}
	rb.Report.AssigneeIDs = append(rb.Report.AssigneeIDs, id)
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

func (rb *ReportBuilder) AddNote(note ReportNote) *ReportBuilder {
	rb.Report.Notes = append(rb.Report.Notes, note)
	rb.Update.AddToSet("notes", note)
	return rb
}
