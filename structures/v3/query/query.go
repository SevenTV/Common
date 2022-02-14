package query

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/redis"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type defaultRoles struct {
	l sync.Mutex
}

func (x *defaultRoles) Fetch(ctx context.Context, mngo mongo.Instance, rdis redis.Instance) []*structures.Role {
	x.l.Lock()
	defer x.l.Unlock()

	// Find from redis
	result := []*structures.Role{}
	key := rdis.ComposeKey("v3", "default-roles")
	s, err := rdis.RawClient().Get(ctx, key.String()).Result()
	if s != "" {
		if err := multierror.Append(err, json.Unmarshal(utils.S2B(s), &result)).ErrorOrNil(); err != nil {
			logrus.WithError(err).Error("redis, could not unmarshal cached default roles")
		}
		return result
	}

	// Find from mongo
	cur, err := mngo.Collection(mongo.CollectionNameRoles).Find(ctx, bson.M{"default": true})
	if err = multierror.Append(err, cur.All(ctx, &result)).ErrorOrNil(); err != nil {
		logrus.WithError(err).Error("could not fetch default roles")
	}

	// Cache result to redis
	b, err := json.Marshal(result)
	if err != nil {
		logrus.WithError(err).Error("could not marshal default roles for redis cache")
	}
	if _, err = rdis.RawClient().Set(ctx, key.String(), utils.B2S(b), time.Minute*5).Result(); err != nil {
		logrus.WithError(err).Error("redis, could not cache default roles")
	}

	logrus.WithField("default_role_count", len(result)).Debug("loaded default roles")
	return result
}

var DefaultRoles defaultRoles
