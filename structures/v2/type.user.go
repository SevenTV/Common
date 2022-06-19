package structures

import (
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Email        string               `json:"email" bson:"email"`
	Rank         int32                `json:"rank" bson:"rank"`
	EmoteIDs     []primitive.ObjectID `json:"emote_ids" bson:"emotes"`
	EditorIDs    []primitive.ObjectID `json:"editor_ids" bson:"editors"`
	RoleID       *primitive.ObjectID  `json:"role_id" bson:"role"`
	TokenVersion string               `json:"token_version" bson:"token_version"`

	// Twitch Data
	TwitchID         string              `json:"twitch_id" bson:"id"`
	YouTubeID        string              `json:"yt_id,omitempty" bson:"yt_id,omitempty"`
	DisplayName      string              `json:"display_name" bson:"display_name"`
	Login            string              `json:"login" bson:"login"`
	BroadcasterType  string              `json:"broadcaster_type" bson:"broadcaster_type"`
	ProfileImageURL  string              `json:"profile_image_url" bson:"profile_image_url"`
	OfflineImageURL  string              `json:"offline_image_url" bson:"offline_image_url"`
	Description      string              `json:"description" bson:"description"`
	CreatedAt        time.Time           `json:"twitch_created_at" bson:"twitch_created_at"`
	ViewCount        int32               `json:"view_count" bson:"view_count"`
	ProfilePictureID string              `json:"profile_picture_id,omitempty" bson:"profile_picture_id,omitempty"`
	EmoteAlias       map[string]string   `json:"-" bson:"emote_alias"`           // Emote Alias - backend only
	Badge            *primitive.ObjectID `json:"badge" bson:"badge"`             // User's badge, if any
	EmoteSlots       int32               `json:"emote_slots" bson:"emote_slots"` // User's maximum channel emote slots

	// Relational Data
	Emotes            []Emote              `json:"emotes" bson:"-"`
	OwnedEmotes       []Emote              `json:"owned_emotes" bson:"-"`
	Editors           []User               `json:"editors" bson:"-"`
	Role              Role                 `json:"role" bson:"-"`
	EditorIn          []User               `json:"editor_in" bson:"-"`
	AuditEntries      []AuditLog           `json:"audit_entries" bson:"-"`
	Reports           []Report             `json:"reports" bson:"-"`
	Bans              []Ban                `json:"bans" bson:"-"`
	Cosmetics         []Cosmetic[bson.Raw] `json:"cosmetics" bson:"-"`
	Notifications     []Notification       `json:"-" bson:"-"`
	NotificationCount int64                `json:"-" bson:"-"`
}

// Test whether a User has a permission flag
func (u User) HasPermission(flag int64) bool {
	return utils.BitField.HasBits(u.Role.Allowed, flag) || utils.BitField.HasBits(u.Role.Allowed, RolePermissionAdministrator)
}
