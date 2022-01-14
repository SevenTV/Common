package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

type Report struct {
	ID         primitive.ObjectID  `json:"id" bson:"_id"`
	ReporterID *primitive.ObjectID `json:"reporter_id" bson:"reporter_id"`
	Reason     string              `json:"reason" bson:"target"`
	Target     *ReportTarget       `json:"target" bson:"target"`
	Cleared    bool                `json:"cleared" bson:"cleared"`

	// Relational Data
	ETarget      Emote      `json:"e_target" bson:"-"`
	UTarget      User       `json:"u_target" bson:"-"`
	Reporter     User       `json:"reporter" bson:"-"`
	AuditEntries []AuditLog `json:"audit_entries" bson:"-"`
}

type ReportTarget struct {
	ID   *primitive.ObjectID `json:"id" bson:"id"`
	Type string              `json:"type" bson:"type"`
}
