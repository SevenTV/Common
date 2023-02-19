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
	Version   string     `json:"version" bson:"version"`
	Overrides []struct{} `json:"overrides" bson:"overrides"`
}
