package mongo

import (
	"context"
	"time"

	"github.com/SevenTV/Common/structures/v3"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Instance interface {
	Collection(CollectionName) *mongo.Collection
	Ping(ctx context.Context) error
	RawClient() *mongo.Client
	RawDatabase() *mongo.Database
	System(ctx context.Context) structures.System
}

type mongoInst struct {
	client *mongo.Client
	db     *mongo.Database
	cache  *cache.Cache
}

func (i *mongoInst) Collection(name CollectionName) *mongo.Collection {
	return i.db.Collection(string(name))
}

func (i *mongoInst) Ping(ctx context.Context) error {
	return i.db.Client().Ping(ctx, nil)
}

func (i *mongoInst) RawClient() *mongo.Client {
	return i.client
}

func (i *mongoInst) RawDatabase() *mongo.Database {
	return i.db
}

func (i *mongoInst) System(ctx context.Context) structures.System {
	v, ok := i.cache.Get("SYSTEM")
	if ok {
		return v.(structures.System)
	}
	result := structures.System{}
	if err := i.Collection(CollectionNameSystem).FindOne(ctx, bson.M{}).Decode(&result); err != nil {
		if err != ErrNoDocuments {
			logrus.WithError(err).Error("mongo, couldn't fetch system info")
		}
		return result
	}

	i.cache.Set("SYSTEM", result, time.Second*30)
	return result
}
