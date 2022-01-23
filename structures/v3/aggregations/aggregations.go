package aggregations

import "go.mongodb.org/mongo-driver/bson"

func Combine(s ...[]bson.D) []bson.D {
	result := []bson.D{}
	for _, v := range s {
		result = append(result, v...)
	}
	return result
}
