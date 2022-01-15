package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

type AuditLog struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type      AuditType          `json:"type" bson:"type"`
	Target    *AuditTarget       `json:"target" bson:"target"`
	Changes   []AuditLogChange   `json:"changes" bson:"changes"`
	Reason    *string            `json:"reason" bson:"reason"`
	CreatedBy primitive.ObjectID `json:"action_user_id" bson:"action_user"`
}

type AuditLogChange struct {
	Key      string      `json:"key" bson:"key"`
	OldValue interface{} `json:"old_value" bson:"old_value"`
	NewValue interface{} `json:"new_value" bson:"new_value"`
}

type AuditTarget struct {
	ID   *primitive.ObjectID `json:"id" bson:"id"`
	Type string              `json:"type" bson:"type"`
}

type AuditType int32

const (
	// Emotes (1-19)
	AuditLogTypeEmoteCreate              AuditType = 1
	AuditLogTypeEmoteDelete              AuditType = 2
	AuditLogTypeEmoteDisable             AuditType = 3
	AuditLogTypeEmoteEdit                AuditType = 4
	AuditLogTypeEmoteUndoDeleteAuditType           = 4
	AuditLogTypeEmoteMerge               AuditType = 5

	// Auth (20-29)
	AuditLogTypeAuthIn  AuditType = 20
	AuditLogTypeAuthOut AuditType = 21

	// Users (30-69)
	AuditLogTypeUserCreate              AuditType = 30
	AuditLogTypeUserDelete              AuditType = 31
	AuditLogTypeUserBan                 AuditType = 32
	AuditLogTypeUserEdit                AuditType = 33
	AuditLogTypeUserChannelEmoteAdd     AuditType = 34
	AuditLogTypeUserChannelEmoteRemove  AuditType = 35
	AuditLogTypeUserUnban               AuditType = 36
	AuditLogTypeUserChannelEditorAdd    AuditType = 37
	AuditLogTypeUserChannelEditorRemove AuditType = 38
	AuditLogTypeUserChannelEmoteEdit    AuditType = 39

	// Admin (70-89)
	AuditLogTypeAppMaintenanceMode AuditType = 70
	AuditLogTypeAppRouteLock       AuditType = 71
	AuditLogTypeAppLogsView        AuditType = 72
	AuditLogTypeAppScale           AuditType = 73
	AuditLogTypeAppNodeCreate      AuditType = 74
	AuditLogTypeAppNodeDelete      AuditType = 75
	AuditLogTypeAppNodeJoin        AuditType = 75
	AuditLogTypeAppNodeUnref       AuditType = 76

	// Reports (90-99)
	AuditLogTypeReport      AuditType = 90
	AuditLogTypeReportClear AuditType = 91
)
