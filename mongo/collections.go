package mongo

import (
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collections = []collectionRef{
	// Collection: Users
	{
		Name: string(CollectionNameUsers),
		Indexes: []IndexModel{
			{Keys: bson.M{"username": 1}, Options: options.Index().SetUnique(true)},
			{Keys: bson.M{"connections.emote_set_id": 1}},
			{Keys: bson.M{"metadata.role_position": -1}},
			{Keys: bson.M{"editors.id": -1}},
		},
		Validator: &jsonSchema{
			BSONType: TList{BSONTypeObject},
			Title:    "Users",
			Required: []string{"username", "discriminator"},
			Properties: map[string]*jsonSchema{
				"type": {BSONType: TList{BSONTypeString}, Enum: []string{"", "BOT", "SYSTEM"}},
				"username": {
					BSONType:  TList{BSONTypeString},
					MinLength: utils.PointerOf(int64(1)),
					MaxLength: utils.PointerOf(int64(25)),
				},
				"dislay_name":   {BSONType: TList{BSONTypeString}},
				"discriminator": {BSONType: TList{BSONTypeString}, MinLength: utils.PointerOf(int64(4)), MaxLength: utils.PointerOf(int64(4))},
				"email":         {BSONType: TList{BSONTypeString}},
				"role_ids": {
					BSONType: TList{BSONTypeArray},
					Items:    []*jsonSchema{{BSONType: TList{BSONTypeObjectId}}},
				},
				"editors": {
					BSONType: TList{BSONTypeArray},
					Items: []*jsonSchema{{
						BSONType: TList{BSONTypeObject},
						Required: []string{"id", "permissions"},
						Properties: map[string]*jsonSchema{
							"id":          {BSONType: TList{BSONTypeObjectId}},
							"connections": {BSONType: TList{BSONTypeArray}, Items: []*jsonSchema{{BSONType: TList{BSONTypeObjectId}}}},
							"permissions": {BSONType: TList{BSONTypeInt32}},
							"visible":     {BSONType: TList{BSONTypeBoolean}},
							"added_at":    {BSONType: TList{BSONTypeDate}},
						},
					}},
				},
				"avatar_id":     {BSONType: TList{BSONTypeString}},
				"biography":     {BSONType: TList{BSONTypeString}},
				"token_version": {BSONType: TList{BSONTypeDouble}},
				"connections": {
					BSONType: TList{BSONTypeArray},
					Items: []*jsonSchema{{
						BSONType: TList{BSONTypeObject},
						Properties: map[string]*jsonSchema{
							"id":        {BSONType: TList{BSONTypeString}},
							"platform":  {BSONType: TList{BSONTypeString}, Enum: []string{"TWITCH", "YOUTUBE"}},
							"linked_at": {BSONType: TList{BSONTypeDate}},
						},
					}},
				},
			},
		},
	},

	// Collection: Emotes
	{
		Name: "emotes",
		Indexes: []IndexModel{
			{Keys: bson.M{"owner_id": -1}},
			{
				Keys:    bson.D{{Key: "versions.id", Value: -1}},
				Options: options.Index().SetUnique(true),
			},
			{Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "tags", Value: "text"},
			}, Options: options.Index().SetTextVersion(3)},
			{Keys: bson.M{"versions.state.channel_count": -1}},
		},
		Validator: &jsonSchema{
			BSONType: TList{BSONTypeObject},
			Title:    "Emotes",
			Required: []string{"name", "versions"},
			Properties: map[string]*jsonSchema{
				"owner_id":    {BSONType: TList{BSONTypeObjectId}},
				"name":        {BSONType: TList{BSONTypeString}, MinLength: utils.PointerOf(int64(1))},
				"flags":       {BSONType: TList{BSONTypeInt32}},
				"tags":        {BSONType: TList{BSONTypeArray}},
				"frame_count": {BSONType: TList{BSONTypeInt32}},
				"formats": {
					BSONType: TList{BSONTypeArray},
					Items: []*jsonSchema{{
						BSONType: TList{BSONTypeObject},
						Properties: map[string]*jsonSchema{
							"name": {
								BSONType: TList{BSONTypeString},
								Enum:     []string{"image/webp", "image/avif", "image/gif", "image/png"},
							},
							"sizes": {
								BSONType: TList{BSONTypeArray},
								Items: []*jsonSchema{{
									BSONType: TList{BSONTypeObject},
									Properties: map[string]*jsonSchema{
										"scale":    {BSONType: TList{BSONTypeString}},
										"width":    {BSONType: TList{BSONTypeInt32}},
										"height":   {BSONType: TList{BSONTypeInt32}},
										"animated": {BSONType: TList{BSONTypeBoolean}},
										"time":     {BSONType: TList{BSONTypeInt64}},
										"length":   {BSONType: TList{BSONTypeInt64}},
									},
								}},
							},
						},
					}},
				},
				"versions": {
					BSONType: TList{BSONTypeArray},
					Items: []*jsonSchema{{
						BSONType: TList{BSONTypeObject},
						Required: []string{"id", "state"},
						Properties: map[string]*jsonSchema{
							"name": {BSONType: TList{BSONTypeString}},
							"state": {
								BSONType: TList{BSONTypeObject},
								Properties: map[string]*jsonSchema{
									"lifecycle": {
										BSONType: TList{BSONTypeInt32},
										Minimum:  utils.PointerOf(int64(-2)),
										Maximum:  utils.PointerOf(int64(3)),
									},
								},
							},
						},
					}},
				},
				"parent_id":    {BSONType: TList{BSONTypeObjectId}},
				"children_ids": {BSONType: TList{BSONTypeArray}, Items: []*jsonSchema{{BSONType: TList{BSONTypeObjectId}}}},
			},
		},
	},

	// Collection: Entitlements
	{
		Name: string(CollectionNameEntitlements),
		Indexes: []IndexModel{
			{Keys: bson.M{"data.ref": -1}},
			{Keys: bson.M{"user_id": 1}},
		},
	},

	// Collection: Emote Sets
	{
		Name: string(CollectionNameEmoteSets),
		Indexes: []IndexModel{
			{Keys: bson.M{"emotes.id": -1}},
			{Keys: bson.M{"owner_id": -1}},
		},
	},

	// Collection: Roles
	{
		Name: string(CollectionNameRoles),
		Indexes: []IndexModel{
			{Keys: bson.M{"position": 1}},
		},
	},

	// Collection: Message Read States
	{
		Name: string(CollectionNameMessagesRead),
		Indexes: []IndexModel{
			{Keys: bson.M{"message_id": -1}},
		},
	},
	// Collection: Messages
	{
		Name: string(CollectionNameMessages),
		Indexes: []IndexModel{
			{ // Partial Index: Mod Requests Only
				Keys: bson.M{"data.target_id": -1},
				Options: options.Index().SetPartialFilterExpression(bson.M{
					"kind": structures.MessageKindModRequest,
				}),
			},
		},
	},
}
