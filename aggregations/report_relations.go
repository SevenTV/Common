package aggregations

import (
	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

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

var ReportRelationAssignees = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameUsers,
			LocalField:   "assignee_ids",
			ForeignField: "_id",
			As:           "assignees",
		},
	}},
}
