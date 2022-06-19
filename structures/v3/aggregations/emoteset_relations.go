package aggregations

import (
	"github.com/seventv/common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// Emote Set Relations
//
// Input: EmoteSet
// Adds Field: "emotes" as []ActiveEmote with the "emote" field added to each ActiveEmote object
// Output: User
var EmoteSetRelationActiveEmotes = mongo.Pipeline{
	// Step 1: Lookup user editors
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameEmotes,
			LocalField:   "emotes.id",
			ForeignField: "versions.id",
			As:           "_emotes",
		},
	}},
	// Step 3: Set "emote" property to each UserEmote object in the original emotes array
	{{
		Key: "$set",
		Value: bson.M{
			"emotes": bson.M{
				"$map": bson.M{
					"input": "$emotes",
					"in": bson.M{
						"$mergeObjects": bson.A{
							"$$this",
							bson.M{
								"emote": bson.M{
									"$arrayElemAt": bson.A{
										"$_emotes",
										bson.M{"$indexOfArray": bson.A{"$_emotes._id", "$$this.id"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
}
