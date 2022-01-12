package mongo

import (
	"context"

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

	for _, ind := range opt.Indexes {
		col := ind.Collection
		if name, err := database.Collection(string(col)).Indexes().CreateOne(ctx, ind.Index); err != nil {
			panic(err)
		} else {
			logrus.WithField("collection", col).Infof("Collection index created: %s", name)
		}
	}

	logrus.Info("mongo, ok")

	return &mongoInst{
		client: client,
		db:     database,
	}, nil
}

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

type (
	Pipeline       = mongo.Pipeline
	WriteModel     = mongo.WriteModel
	InsertOneModel = mongo.InsertOneModel
	UpdateOneModel = mongo.UpdateOneModel
	IndexModel     = mongo.IndexModel
)
