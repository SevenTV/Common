package structures

import (
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	RegExpEmoteName               = regexp.MustCompile(`^[-_A-Za-z(!?&)$:0-9]{2,100}$`)
	RegExpEmoteVersionName        = regexp.MustCompile(`^[A-Za-z0-9\s]{2,40}$`)
	RegExpEmoteVersionDescription = regexp.MustCompile(`^[-_*=/\\"'\]\[}{@&~!?;:A-Za-z0-9\s]{3,240}$`)
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

func (u UpdateMap) Push(key string, value UpdateValue) UpdateMap {
	if _, ok := u["$push"]; !ok {
		u["$push"] = bson.M{
			key: value,
		}
	} else {
		m := u["$push"].(bson.M)
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

func (u UpdateMap) UndoSet(key string) UpdateMap {
	if m, ok := u["$set"]; ok {
		delete(m.(bson.M), key)
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

type ObjectKind int8

const (
	ObjectKindUser        ObjectKind = 1
	ObjectKindEmote       ObjectKind = 2
	ObjectKindEmoteSet    ObjectKind = 3
	ObjectKindRole        ObjectKind = 4
	ObjectKindEntitlement ObjectKind = 5
	ObjectKindBan         ObjectKind = 6
	ObjectKindMessage     ObjectKind = 7
	ObjectKindReport      ObjectKind = 8
)

type Object interface {
	AuditLog | Ban | Cosmetic[bson.Raw] | Emote | EmoteSet | Entitlement[bson.Raw] | Message[bson.Raw] | Report | Role | User
}

func (k ObjectKind) CollectionName() string {
	switch k {
	case ObjectKindUser:
		return "users"
	case ObjectKindEmote:
		return "emotes"
	case ObjectKindEmoteSet:
		return "emote_sets"
	case ObjectKindRole:
		return "roles"
	case ObjectKindEntitlement:
		return "entitlements"
	case ObjectKindBan:
		return "bans"
	case ObjectKindMessage:
		return "messages"
	default:
		return ""
	}
}
