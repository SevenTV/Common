package structures

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (u UpdateMap) Pull(key string, value UpdateValue) UpdateMap {
	if _, ok := u["$pull"]; !ok {
		u["$pull"] = bson.M{
			key: value,
		}
	} else {
		m := u["$pull"].(bson.M)
		m[key] = value
	}

	return u
}

func (u UpdateMap) Clear() {
	for k := range u {
		delete(u, k)
	}
}

var (
	ErrUnknownEmote          error = fmt.Errorf("unknown emote")
	ErrUnknownUser           error = fmt.Errorf("unknown user")
	ErrInsufficientPrivilege error = fmt.Errorf("insufficient privilege")
	ErrInternalError         error = fmt.Errorf("internal error occured")
	ErrIncompleteMutation    error = fmt.Errorf("the mutation struct was not set up properly")
)

type ObjectID = primitive.ObjectID
