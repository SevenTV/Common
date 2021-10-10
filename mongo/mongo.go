package mongo

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Database *mongo.Database

var ErrNoDocuments = mongo.ErrNoDocuments

type Pipeline = mongo.Pipeline

func Setup(opt SetupOptions) {
	clientOptions := options.Client().ApplyURI(opt.URI)
	if opt.Direct {
		clientOptions.SetDirect(true)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	// Send a Ping
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	Database = client.Database(opt.DB)

	for _, ind := range opt.Indexes {
		col := ind.Collection
		if name, err := Database.Collection(string(col)).Indexes().CreateOne(ctx, ind.Index); err != nil {
			panic(err)
		} else {
			log.WithField("collection", col).Infof("Collection index created: %s", name)
		}
	}

	log.Info("mongo, ok")
}

func Collection(name CollectionName) *mongo.Collection {
	return Database.Collection(string(name))
}

type CollectionName string

var (
	CollectionNameEmotes CollectionName = "emotes"
	CollectionNameUsers  CollectionName = "users_v3"
)

type SetupOptions struct {
	URI     string
	DB      string
	Direct  bool
	Indexes []IndexRef
}

type IndexRef struct {
	Collection CollectionName
	Index      mongo.IndexModel
}

type IndexModel = mongo.IndexModel
