package structures

type System struct {
	ID ObjectID `json:"id" bson:"_id"`

	AdminUserID ObjectID `json:"admin_user_id" bson:"admin_user_id"`
	EmoteSetID  ObjectID `json:"emote_set_id" bson:"emote_set_id"`
}

type SystemDefaultObject struct {
	ID         ObjectID `json:"id" bson:"id"`
	Collection string   `json:"collection" bson:"collection"`
}
