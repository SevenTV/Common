package mongo

type CollectionName string

var (
	CollectionNameSystem       CollectionName = "system"
	CollectionNameActivities   CollectionName = "activities"
	CollectionNameAuditLogs    CollectionName = "audit_logs"
	CollectionNameEmotes       CollectionName = "emotes"
	CollectionNameEmoteSets    CollectionName = "emote_sets"
	CollectionNameUsers        CollectionName = "users"
	CollectionNameRoles        CollectionName = "roles"
	CollectionNameEntitlements CollectionName = "entitlements"
	CollectionNameCosmetics    CollectionName = "cosmetics"
	CollectionNameReports      CollectionName = "reports"
	CollectionNameBans         CollectionName = "bans"
	CollectionNameMessages     CollectionName = "messages"
	CollectionNameMessagesRead CollectionName = "messages_read"
)
