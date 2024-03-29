package mongo

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
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
	uri := options.Client().ApplyURI(opt.URI)

	if opt.Username != "" && opt.Password != "" {
		uri.SetAuth(options.Credential{
			Username: opt.Username,
			Password: opt.Password,
		})
	}

	var readPref *readpref.ReadPref = readpref.Primary()

	if opt.HedgedReads {
		readPref = readpref.PrimaryPreferred(readpref.WithHedgeEnabled(true))
	}

	client, err := mongo.Connect(ctx, uri.SetDirect(opt.Direct).SetReadPreference(readPref).SetRetryReads(true))
	if err != nil {
		return nil, err
	}

	// Send a Ping
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	database := client.Database(opt.DB)

	inst := &mongoInst{
		client: client,
		db:     database,
		cache:  cache.New(time.Second*10, time.Second*20),
	}
	return inst, nil
}

type SetupOptions struct {
	URI         string
	DB          string
	Direct      bool
	Username    string
	Password    string
	HedgedReads bool
}

type (
	Pipeline        = mongo.Pipeline
	WriteModel      = mongo.WriteModel
	InsertOneModel  = mongo.InsertOneModel
	UpdateOneModel  = mongo.UpdateOneModel
	DeleteOneModel  = mongo.DeleteOneModel
	UpdateManyModel = mongo.UpdateManyModel
	IndexModel      = mongo.IndexModel
)
