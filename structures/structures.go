package structures

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/sirupsen/logrus"
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

var (
	ErrUnknownEmote          error = fmt.Errorf("unknown emote")
	ErrUnknownUser           error = fmt.Errorf("unknown user")
	ErrInsufficientPrivilege error = fmt.Errorf("insufficient privilege")
	ErrInternalError         error = fmt.Errorf("internal error occured")
	ErrIncompleteMutation    error = fmt.Errorf("the mutation struct was not set up properly")
)

type ObjectID = primitive.ObjectID

var (
	DefaultRoles     []*Role = nil
	defaultRolesLock sync.Mutex
)

func FetchDefaultRoles(ctx context.Context, mngo mongo.Instance) []*Role {
	defaultRolesLock.Lock()
	defer defaultRolesLock.Unlock()
	if DefaultRoles != nil {
		return DefaultRoles
	}
	go func() {
		time.Sleep(time.Minute)
		DefaultRoles = nil
	}()

	roles := []*Role{}
	cur, err := mngo.Collection(mongo.CollectionNameRoles).Find(ctx, bson.M{"default": true})
	if err != nil {
		logrus.WithError(err).Error("mongo, could not fetch default roles")
	}
	if err = cur.All(ctx, &roles); err != nil {
		logrus.WithError(err).Error("mongo, could not fetch default roles")
	}

	DefaultRoles = roles
	return roles
}
