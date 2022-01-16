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
				"username": {
					BSONType:    BSONTypeString,
					Title:       "Username",
					Description: "The user's username",
					MaxLength:   utils.Int64Pointer(25),
				},
			},
		},
		Indexes: []IndexModel{
			{Keys: bson.M{"username": -1}, Options: options.Index().SetUnique(true)},
		},
	},
}
