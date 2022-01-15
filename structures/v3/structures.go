package structures

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/redis"
	"github.com/SevenTV/Common/utils"
	"github.com/hashicorp/go-multierror"
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

type defaultRoles struct {
	List []*Role
	l    sync.Mutex
}

func (x *defaultRoles) Fetch(ctx context.Context, mngo mongo.Instance, redis redis.Instance) []*Role {
	x.l.Lock()
	defer x.l.Lock()

	// Find from redis
	result := []*Role{}
	key := redis.ComposeKey("v3", "cache:default-roles")
	s := redis.RawClient().Get(ctx, key.String()).Val()
	if s != "" {
		if err := bson.Unmarshal(utils.S2B(s), result); err != nil {
			logrus.WithError(err).Error("redis, could not unmarshal cached default roles")
		}
		if len(result) > 0 { // return result from cache
			return result
		}
	}

	// Find from mongo
	cur, err := mngo.Collection(CollectionNameRoles).Find(ctx, bson.M{"default": true})
	if err = multierror.Append(err, cur.All(ctx, result)).ErrorOrNil(); err != nil {
		logrus.WithError(err).Error("could not fetch default roles")
	}

	// Cache result to redis
	b, err := bson.Marshal(result)
	if err != nil {
		logrus.WithError(err).Error("could not marshal default roles for redis cache")
	}
	if _, err = redis.RawClient().Set(ctx, key.String(), utils.B2S(b), time.Minute*1).Result(); err != nil {
		logrus.WithError(err).Error("redis, could not cache default roles")
	}

	x.List = result
	return result
}

var DefaultRoles defaultRoles
