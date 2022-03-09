package query

import (
	"context"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func (q *Query) GlobalEmoteSet(ctx context.Context) (*structures.EmoteSet, error) {
	mx := q.lock("GlobalEmoteSet")
	defer mx.Unlock()
	k := q.key("global_emote_set")

	set := &structures.EmoteSet{}

	// Get cached
	if ok := q.getFromMemCache(ctx, k, set); ok {
		return set, nil
	}

	sys := q.mongo.System(ctx)

	if err := q.mongo.Collection(mongo.CollectionNameEmoteSets).FindOne(ctx, bson.M{"_id": sys.EmoteSetID}).Decode(set); err != nil {
		logrus.WithError(err).Error("mongo, couldn't decode global emote set")
		return nil, err
	}

	// Set cache
	if err := q.setInMemCache(ctx, k, set, time.Second*30); err != nil {
		return nil, err
	}

	return set, nil
}
