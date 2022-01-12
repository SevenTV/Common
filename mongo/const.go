package mongo

type CollectionName string

var (
	CollectionNameEmotes       CollectionName = "emotes"
	CollectionNameEmoteSets    CollectionName = "emotes_sets"
	CollectionNameUsers        CollectionName = "users"
	CollectionNameRoles        CollectionName = "roles"
	CollectionNameEntitlements CollectionName = "entitlements"
	CollectionNameReports      CollectionName = "reports"
	CollectionNameBans         CollectionName = "bans"
	CollectionNameMessages     CollectionName = "messages"
	CollectionNameMessagesRead CollectionName = "messages_read"
)
