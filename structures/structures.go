package structures

import (
	"go.mongodb.org/mongo-driver/bson"
)

type UpdateMap bson.M

type UpdateValue interface{}

func (u UpdateMap) Set(key string, value UpdateValue) UpdateMap {
	if _, ok := u["$set"]; !ok {
		u["$set"] = bson.M{
			key: value,
		}
	} else {
		m := u["$set"].(bson.M)
		m[key] = value
	}

	return u
}

func (u UpdateMap) AddToSet(key string, value UpdateValue) UpdateMap {
	if _, ok := u["$addToSet"]; !ok {
		u["$addToSet"] = bson.M{
			key: value,
		}
	} else {
		m := u["$addToSet"].(bson.M)
		m[key] = value
	}

	return u
}
