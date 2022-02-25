package query

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (q *Query) UserEditorOf(ctx context.Context, id primitive.ObjectID) ([]*structures.UserEditor, error) {
	cur, err := q.mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.M{
				"editors.id": id,
			},
		}},
		{{
			Key: "$project",
			Value: bson.M{
				"editor": bson.M{
					"$mergeObjects": bson.A{
						bson.M{"$first": bson.M{"$filter": bson.M{
							"input": "$editors",
							"as":    "ed",
							"cond": bson.M{
								"$eq": bson.A{"$$ed.id", id},
							},
						}}},
						bson.M{"id": "$_id"},
					},
				},
			},
		}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$editor"}}},
	})
	if err != nil {
		logrus.WithError(err).Error("query, failed to spawn aggregation")
		return nil, err
	}

	v := []*structures.UserEditor{}
	if err = cur.All(ctx, &v); err != nil {
		return nil, err
	}

	return v, nil
}
