package structures

import "github.com/seventv/common/mongo"

var (
	CollectionNameEmotes            = mongo.CollectionName("emotes")
	CollectionNameUsers             = mongo.CollectionName("users")
	CollectionNameBans              = mongo.CollectionName("bans")
	CollectionNameReports           = mongo.CollectionName("reports")
	CollectionNameCosmetics         = mongo.CollectionName("cosmetics")
	CollectionNameRoles             = mongo.CollectionName("roles")
	CollectionNameAudit             = mongo.CollectionName("audit")
	CollectionNameEntitlements      = mongo.CollectionName("entitlements")
	CollectionNameNotifications     = mongo.CollectionName("notifications")
	CollectionNameNotificationsRead = mongo.CollectionName("notifications_read")
)
