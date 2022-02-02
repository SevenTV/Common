package aggregations

import (
	"fmt"

	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func Combine(s ...mongo.Pipeline) mongo.Pipeline {
	result := mongo.Pipeline{}
	for _, v := range s {
		result = append(result, v...)
	}
	return result
}

func CombinePtr(s ...mongo.Pipeline) *mongo.Pipeline {
	v := Combine(s...)
	return &v
}

// MergeArrays combines two arrays and adds a new field from one to the other on corresponding elements
func MergeArrays(key, mergeWith, mergePath, mergeKey string) bson.D {
	return bson.D{{
		Key: "$set",
		Value: bson.M{
			key: bson.M{"$map": bson.M{
				"input": fmt.Sprintf("$%s", key),
				"in": bson.M{"$mergeObjects": bson.A{
					"$$this",
					bson.M{
						mergeKey: bson.M{"$arrayElemAt": bson.A{
							mergeWith,
							bson.M{"$indexOfArray": bson.A{fmt.Sprintf("%s.%s", mergeWith, mergeKey), "$$this.id"}},
						}},
					},
				}},
			}},
		},
	}}
}
