package structures

type System struct {
	ID ObjectID `json:"id" bson:"_id"`

	AdminUserID ObjectID `json:"admin_user_id" bson:"admin_user_id"`
	EmoteSetID  ObjectID `json:"emote_set_id" bson:"emote_set_id"`
	Config      struct {
		Extension     SystemConfigExtension `json:"extension" bson:"extension"`
		ExtensionBeta SystemConfigExtension `json:"extension_beta" bson:"extension_beta"`
	} `json:"config" bson:"config"`
}

type SystemDefaultObject struct {
	ID         ObjectID `json:"id" bson:"id"`
	Collection string   `json:"collection" bson:"collection"`
}

type SystemConfigExtension struct {
	Version       string                        `json:"version" bson:"version"`
	Overrides     []struct{}                    `json:"overrides" bson:"overrides"`
	Compatibility []SystemConfigExtensionCompat `json:"compatibility" bson:"compatibility"`
}

type SystemConfigExtensionCompat struct {
	ID     []string                           `json:"id" bson:"id"`
	Issues []SystemConfigExtensionCompatIssue `json:"issues" bson:"issues"`
}

type SystemConfigExtensionCompatIssue struct {
	Platforms []UserConnectionPlatform                 `json:"platform" bson:"platform"`
	Severity  SystemConfigExtensionCompatIssueSeverity `json:"severity" bson:"severity"`
	Message   string                                   `json:"message" bson:"message"`
}

type SystemConfigExtensionCompatIssueSeverity string

const (
	SystemConfigExtensionCompatIssueSeverityNote                   SystemConfigExtensionCompatIssueSeverity = "NOTE"
	SystemConfigExtensionCompatIssueSeverityWarning                SystemConfigExtensionCompatIssueSeverity = "WARNING"
	SystemConfigExtensionCompatIssueSeverityBadPerformance         SystemConfigExtensionCompatIssueSeverity = "BAD_PERFORMANCE"
	SystemConfigExtensionCompatIssueSeverityClashing               SystemConfigExtensionCompatIssueSeverity = "CLASHING"
	SystemConfigExtensionCompatIssueSeverityDuplicateFunctionality SystemConfigExtensionCompatIssueSeverity = "DUPLICATE_FUNCTIONALITY"
)
