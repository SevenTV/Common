package mongo

import (
	"time"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collections = []collectionRef{
	{
		Name: "users",
		Indexes: []IndexModel{
			{Keys: bson.M{"username": -1}, Options: options.Index().SetUnique(true)},
		},
		Validator: jsonSchema{
			BSONType: BSONTypeObject,
			Title:    "Users",
			Required: []string{"username", "discriminator"},
			Properties: map[string]*jsonSchema{
				"type": {BSONType: BSONTypeString, Enum: []string{"", "BOT", "SYSTEM"}},
				"username": {
					BSONType:  BSONTypeString,
					MinLength: utils.Int64Pointer(1),
					MaxLength: utils.Int64Pointer(25),
				},
				"dislay_name":   {BSONType: BSONTypeString},
				"discriminator": {BSONType: BSONTypeString, MinLength: utils.Int64Pointer(4), MaxLength: utils.Int64Pointer(4)},
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
		DefaultObjects: []interface{}{
			&structures.User{
				ID:            primitive.NewObjectIDFromTimestamp(time.Date(2021, time.February, 26, 0, 0, 0, 0, time.UTC)),
				UserType:      structures.UserTypeSystem,
				Username:      "admin",
				DisplayName:   "Admin",
				Discriminator: "0001",
				Email:         "system@7tv.app",
				Biography:     "System-generated account",
				TokenVersion:  0,
			},
		},
	},

	{
		Name: "emotes",
		Indexes: []IndexModel{
			{Keys: bson.M{"owner_id": -1}},
			{Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "tags", Value: "text"},
			}, Options: options.Index().SetTextVersion(3)},
		},
		Validator: jsonSchema{
			BSONType: BSONTypeObject,
			Title:    "Emotes",
			Required: []string{"name", "status"},
			Properties: map[string]*jsonSchema{
				"owner_id": {BSONType: BSONTypeObjectId},
				"name":     {BSONType: BSONTypeString, MinLength: utils.Int64Pointer(1)},
				"flags":    {BSONType: BSONTypeInt32},
				"tags":     {BSONType: BSONTypeArray, UniqueItems: utils.BoolPointer(true)},
				"status": {
					BSONType: BSONTypeInt32,
					Minimum:  utils.Int64Pointer(-2),
					Maximum:  utils.Int64Pointer(3),
				},
				"frame_count": {BSONType: BSONTypeInt32},
				"formats": {
					BSONType: BSONTypeArray,
					Items: []*jsonSchema{{
						BSONType: BSONTypeObject,
						Properties: map[string]*jsonSchema{
							"name": {
								BSONType: BSONTypeString,
								Enum:     []string{"image/webp", "image/avif", "image/gif", "image/png"},
							},
							"sizes": {
								BSONType: BSONTypeArray,
								Items: []*jsonSchema{{
									BSONType: BSONTypeObject,
									Properties: map[string]*jsonSchema{
										"scale":    {BSONType: BSONTypeString},
										"width":    {BSONType: BSONTypeInt32},
										"height":   {BSONType: BSONTypeInt32},
										"animated": {BSONType: BSONTypeBoolean},
										"time":     {BSONType: BSONTypeInt64},
										"length":   {BSONType: BSONTypeInt64},
									},
								}},
							},
						},
					}},
				},
				"parent_id":    {BSONType: BSONTypeObjectId},
				"children_ids": {BSONType: BSONTypeArray, Items: []*jsonSchema{{BSONType: BSONTypeObjectId}}},
			},
		},
	},
}
