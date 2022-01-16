package mongo

import (
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collections = []collectionRef{
	{
		Name: "users",
		Validator: jsonSchema{
			BSONType: BSONTypeObject,
			Title:    "users",
			Required: []string{"username", "discriminator"},
			Properties: map[string]*jsonSchema{
				"type": {BSONType: BSONTypeString, Enum: []string{"", "BOT", "SYSTEM"}},
				"username": {
					BSONType:    BSONTypeString,
					Title:       "Username",
					Description: "The user's username",
					MinLength:   utils.Int64Pointer(1),
					MaxLength:   utils.Int64Pointer(25),
				},
				"dislay_name":   {BSONType: BSONTypeString},
				"discriminator": {BSONType: BSONTypeString},
				"email":         {BSONType: BSONTypeString, Pattern: `^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`},
				"editors": {
					BSONType: BSONTypeArray,
					Items: []*jsonSchema{{
						BSONType: BSONTypeObject,
						Required: []string{"id", "permissions"},
						Properties: map[string]*jsonSchema{
							"id":          {BSONType: BSONTypeObjectId},
							"connections": {BSONType: BSONTypeArray, Items: []*jsonSchema{{BSONType: BSONTypeObjectId}}},
							"permissions": {BSONType: BSONTypeInt32},
							"visible":     {BSONType: BSONTypeBoolean},
							"added_at":    {BSONType: BSONTypeDate},
						},
					}},
				},
				"avatar_id":     {BSONType: BSONTypeString},
				"biography":     {BSONType: BSONTypeString},
				"token_version": {BSONType: BSONTypeDouble},
				"connections": {
					BSONType: BSONTypeArray,
					Items: []*jsonSchema{{
						BSONType: BSONTypeObject,
						Properties: map[string]*jsonSchema{
							"id":        {BSONType: BSONTypeString},
							"platform":  {BSONType: BSONTypeString, Enum: []string{"TWITCH", "YOUTUBE"}},
							"linked_at": {BSONType: BSONTypeDate},
							"grant": {
								BSONType: BSONTypeObject,
								Properties: map[string]*jsonSchema{
									"access_token":  {BSONType: BSONTypeString},
									"refresh_token": {BSONType: BSONTypeString},
									"scope":         {BSONType: BSONTypeArray, Items: []*jsonSchema{{BSONType: BSONTypeString}}},
									"expires_at":    {BSONType: BSONTypeDate},
								},
							},
						},
					}},
				},
			},
		},
		Indexes: []IndexModel{
			{Keys: bson.M{"username": -1}, Options: options.Index().SetUnique(true)},
		},
	},
}
