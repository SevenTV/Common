package structures

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Kind    AuditLogKind       `json:"kind" bson:"kind"`
	ActorID primitive.ObjectID `json:"actor_id" bson:"actor_id"`

	TargetID   primitive.ObjectID `json:"target_id" bson:"target_id"`
	TargetKind ObjectKind         `json:"target_kind" bson:"target_kind"`
	Changes    []AuditLogChange   `json:"changes" bson:"changes"`

	Extra  map[string]any `json:"extra,omitempty" bson:"extra,omitempty"`
	Reason string         `json:"reason,omitempty" bson:"reason,omitempty"`
}

type AuditLogKind int8

const (
	// Range: 1-19 (Emote)

	AuditLogKindCreateEmote     AuditLogKind = 1 // emote was created
	AuditLogKindDeleteEmote     AuditLogKind = 2 // emote was deleted
	AuditLogKindDisableEmote    AuditLogKind = 3 // emote was disabled
	AuditLogKindUpdateEmote     AuditLogKind = 4 // emote was updated
	AuditLogKindMergeEmote      AuditLogKind = 5 // emote was merged
	AuditLogKindUndoDeleteEmote AuditLogKind = 6 // deleted emote was restored
	AuditLogKindEnableEmote     AuditLogKind = 7 // emote was enabled

	// Range: 20-29 (Access)

	AuditLogKindSignUserToken  AuditLogKind = 20 // a user token was signed
	AuditLogKindSignCSRFToken  AuditLogKind = 21 // a CSRF token was signed
	AuditLogKindRejectedAccess AuditLogKind = 26 // an attempt to access a privileged area was rejected

	// Range: 30-69 (User)

	AuditLogKindCreateUser       AuditLogKind = 30 // user was created
	AuditLogKindDeleteUser       AuditLogKind = 31 // user was deleted
	AuditLogKindBanUser          AuditLogKind = 32 // user was banned
	AuditLogKindEditUser         AuditLogKind = 33 // user was edited
	AuditLogKindUnban            AuditLogKind = 36 // user was unbanned
	AuditLogKindAddUserEditor    AuditLogKind = 37 // editor added to user
	AuditLogKindRemoveUserEditor AuditLogKind = 38 // editor removed from user

	// Range: 70-79 (Emote Set)

	AuditLogKindCreateEmoteSet AuditLogKind = 70 // emote set was created
	AuditLogKindUpdateEmoteSet AuditLogKind = 71 // emote set was updated
	AuditLogKindDeleteEmoteSet AuditLogKind = 72 // emote set was deleted
)

type AuditLogChange struct {
	Format AuditLogChangeFormat `json:"format" bson:"format"`
	Key    string               `json:"key" bson:"key"`
	Value  bson.Raw             `json:"value" bson:"value"`
}

func (alc *AuditLogChange) WriteSingleValues(old any, new any) *AuditLogChange {
	sv := &AuditLogChangeSingleValue{}
	sv.Old = old
	sv.New = new

	alc.Value, _ = bson.Marshal(sv)
	return alc
}

func (alc *AuditLogChange) WriteArrayAdded(values ...any) *AuditLogChange {
	ac := &AuditLogChangeArrayChange{}
	ac.Added = append(ac.Added, values...)
	alc.Value, _ = bson.Marshal(ac)
	return alc
}

func (alc *AuditLogChange) WriteArrayRemoved(values ...any) *AuditLogChange {
	ac := &AuditLogChangeArrayChange{}
	ac.Removed = append(ac.Added, values...)
	alc.Value, _ = bson.Marshal(ac)
	return alc
}

func (alc *AuditLogChange) WriteArrayUpdated(values ...AuditLogChangeSingleValue) *AuditLogChange {
	ac := &AuditLogChangeArrayChange{}

	ac.Updated = append(ac.Updated, values...)
	alc.Value, _ = bson.Marshal(ac)
	return alc
}

type AuditLogChangeFormat int8

const (
	AuditLogChangeFormatSingleValue AuditLogChangeFormat = 1
	AuditLogChangeFormatArrayChange AuditLogChangeFormat = 2
)

type AuditLogChangeSingleValue struct {
	New      any   `json:"n" bson:"n"`
	Old      any   `json:"o" bson:"o"`
	Position int32 `json:"p,omitempty" bson:"p,omitempty"`
}

type AuditLogChangeArrayChange struct {
	Added   []any                       `json:"added,omitempty" bson:"added,omitempty"`
	Removed []any                       `json:"removed,omitempty" bson:"removed,omitempty"`
	Updated []AuditLogChangeSingleValue `json:"updated,omitempty" bson:"updated,omitempty"`
}
