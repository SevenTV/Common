package aggregations

import (
	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

var ReportRelationTarget = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameUsers,
			LocalField:   "target_id",
			ForeignField: "_id",
			As:           "targets",
		},
	}},
	{{
		Key: "$set",
		Value: bson.M{
			"target": bson.M{"$first": "$targets"},
		},
	}},
	{{Key: "$unset", Value: bson.A{"targets"}}},
}

var ReportRelationReporter = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameUsers,
			LocalField:   "reporter_id",
			ForeignField: "_id",
			As:           "reporters",
		},
	}},
	{{
		Key: "$set",
		Value: bson.M{
			"reporter": bson.M{"$first": "$reporters"},
		},
	}},
	{{Key: "$unset", Value: bson.A{"reporters"}}},
}
