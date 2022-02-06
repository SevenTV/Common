package mongo

type CollectionName string

var (
	CollectionNameSystem       CollectionName = "system"
	CollectionNameEmotes       CollectionName = "emotes"
	CollectionNameEmoteSets    CollectionName = "emote_sets"
	CollectionNameUsers        CollectionName = "users"
	CollectionNameRoles        CollectionName = "roles"
	CollectionNameEntitlements CollectionName = "entitlements"
	CollectionNameReports      CollectionName = "reports"
	CollectionNameBans         CollectionName = "bans"
	CollectionNameMessages     CollectionName = "messages"
	CollectionNameMessagesRead CollectionName = "messages_read"
)
