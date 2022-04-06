package mongo

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ErrNoDocuments = mongo.ErrNoDocuments

type Lookup struct {
	From         CollectionName `bson:"from"`
	LocalField   string         `bson:"localField"`
	ForeignField string         `bson:"foreignField"`
	As           string         `bson:"as"`
}

type LookupWithPipeline struct {
	From     CollectionName  `bson:"from"`
	Let      bson.M          `bson:"let"`
	Pipeline *mongo.Pipeline `bson:"pipeline"`
	As       string          `bson:"as"`
}

func Setup(ctx context.Context, opt SetupOptions) (Instance, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(opt.URI).SetDirect(opt.Direct))
	if err != nil {
		return nil, err
	}

	// Send a Ping
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	database := client.Database(opt.DB)

	logrus.Info("mongo, ok")
	inst := &mongoInst{
		client: client,
		db:     database,
		cache:  cache.New(time.Second*10, time.Second*20),
	}
	if opt.CollSync {
		go collSync(inst)
	}
	return inst, nil
}

type SetupOptions struct {
	URI      string
	DB       string
	Direct   bool
	CollSync bool
}

type IndexRef struct {
	Collection CollectionName
	Index      mongo.IndexModel
}

type (
	Pipeline       = mongo.Pipeline
	WriteModel     = mongo.WriteModel
	InsertOneModel = mongo.InsertOneModel
	UpdateOneModel = mongo.UpdateOneModel
	IndexModel     = mongo.IndexModel
)
