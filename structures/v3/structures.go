package structures

import (
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	RegExpEmoteName               = regexp.MustCompile(`^[-_A-Za-zÀ-ÖØ-öø-įĴ-őŔ-žǍ-ǰǴ-ǵǸ-țȞ-ȟȤ-ȳɃɆ-ɏḀ-ẞƀ-ƓƗ-ƚƝ-ơƤ-ƥƫ-ưƲ-ƶẠ-ỿ(!?&)$+:0-9]{2,100}$`)
	RegExpEmoteVersionName        = regexp.MustCompile(`^[A-Za-zÀ-ÖØ-öø-įĴ-őŔ-žǍ-ǰǴ-ǵǸ-țȞ-ȟȤ-ȳɃɆ-ɏḀ-ẞƀ-ƓƗ-ƚƝ-ơƤ-ƥƫ-ưƲ-ƶẠ-ỿ0-9\s]{2,40}$`)
	RegExpEmoteVersionDescription = regexp.MustCompile(`^[-_*=/\\"'\]\[}{@&~!?;:A-Za-zÀ-ÖØ-öø-įĴ-őŔ-žǍ-ǰǴ-ǵǸ-țȞ-ȟȤ-ȳɃɆ-ɏḀ-ẞƀ-ƓƗ-ƚƝ-ơƤ-ƥƫ-ưƲ-ƶẠ-ỿ0-9\s]{3,240}$`)
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

func (u UpdateMap) Has(operator, key string) bool {
	if m, ok := u[operator]; ok {
		if _, ok := m.(bson.M)[key]; ok {
			return true
		}
	}

	return false
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
	ObjectKindPresence    ObjectKind = 9
	ObjectKindCosmetic    ObjectKind = 10
)

type Object interface {
	AuditLog | Ban | Cosmetic[bson.Raw] | Emote | EmoteSet | Entitlement[bson.Raw] | Message[bson.Raw] | Report | Role | User
}

func (k ObjectKind) String() string {
	switch k {
	case ObjectKindUser:
		return "USER"
	case ObjectKindEmote:
		return "EMOTE"
	case ObjectKindEmoteSet:
		return "EMOTE_SET"
	case ObjectKindRole:
		return "ROLE"
	case ObjectKindEntitlement:
		return "ENTITLEMENT"
	case ObjectKindBan:
		return "BAN"
	case ObjectKindMessage:
		return "MESSAGE"
	case ObjectKindReport:
		return "REPORT"
	case ObjectKindPresence:
		return "PRESENCE"
	case ObjectKindCosmetic:
		return "COSMETIC"
	default:
		return ""
	}
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

type ListItemAction string

const (
	ListItemActionAdd    ListItemAction = "ADD"
	ListItemActionUpdate ListItemAction = "UPDATE"
	ListItemActionRemove ListItemAction = "REMOVE"
)

type BitField[T ~int64 | ~int32] int64

func (b BitField[T]) Has(flag T) bool {
	return int64(b)&int64(flag) != 0
}

func (b BitField[T]) Set(flag T) BitField[T] {
	return BitField[T](int64(b) | int64(flag))
}

func (b BitField[T]) Unset(flag T) BitField[T] {
	return BitField[T](int64(b) &^ int64(flag))
}

func (b BitField[T]) Value() T {
	return T(b)
}
