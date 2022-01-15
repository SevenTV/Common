package structures

import "github.com/SevenTV/Common/mongo"

var (
	CollectionNameEmotes       mongo.CollectionName = "emotes"
	CollectionNameEmoteSets    mongo.CollectionName = "emotes_sets"
	CollectionNameUsers        mongo.CollectionName = "users"
	CollectionNameRoles        mongo.CollectionName = "roles"
	CollectionNameEntitlements mongo.CollectionName = "entitlements"
	CollectionNameReports      mongo.CollectionName = "reports"
	CollectionNameBans         mongo.CollectionName = "bans"
	CollectionNameMessages     mongo.CollectionName = "messages"
	CollectionNameMessagesRead mongo.CollectionName = "messages_read"
)
