package structures

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLogBuilder struct {
	Update   UpdateMap
	AuditLog AuditLog
}

// NewAuditLogBuilder creates a new Builder utility for an Audit Log
func NewAuditLogBuilder(log AuditLog) *AuditLogBuilder {
	return &AuditLogBuilder{
		Update:   UpdateMap{},
		AuditLog: log,
	}
}

// SetKind sets the kind of the audit log
func (alb *AuditLogBuilder) SetKind(kind AuditLogKind) *AuditLogBuilder {
	alb.AuditLog.Kind = kind
	alb.Update.Set("kind", kind)
	return alb
}

// SetActor defines the ID of the responsible ('actor') user in the audit log
func (alb *AuditLogBuilder) SetActor(id primitive.ObjectID) *AuditLogBuilder {
	alb.AuditLog.ActorID = id
	alb.Update.Set("actor_id", id)
	return alb
}

// SetTargetKind sets the object kind of the resource targeted in the audit log
func (alb *AuditLogBuilder) SetTargetKind(kind ObjectKind) *AuditLogBuilder {
	alb.AuditLog.TargetKind = kind
	alb.Update.Set("target_kind", kind)
	return alb
}

// SetTargetID sets the id of the resource targeted in the audit log
func (alb *AuditLogBuilder) SetTargetID(id primitive.ObjectID) *AuditLogBuilder {
	alb.AuditLog.TargetID = id
	alb.Update.Set("target_id", id)
	return alb
}

// AddChanges adds one or more changes in the audit log
func (alb *AuditLogBuilder) AddChanges(changes ...*AuditLogChange) *AuditLogBuilder {
	alb.AuditLog.Changes = append(alb.AuditLog.Changes, changes...)
	alb.Update.Push("changes", changes)
	return alb
}

// SetExtra defines arbitrary extraneous data that may be helpful in some cases
// where changes cannot be explained with just old and new values
func (alb *AuditLogBuilder) SetExtra(key string, value interface{}) *AuditLogBuilder {
	alb.AuditLog.Extra[key] = value
	alb.Update.Set(fmt.Sprintf("extra.%s", key), value)
	return alb
}
