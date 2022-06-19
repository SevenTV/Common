package aggregations

import (
	"github.com/seventv/common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// Report Relations
//
// Input: Report
// Adds Field: "reporter" as User
// Output: Report
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

// Report Relations
//
// Input: Report
// Adds Field: "assignees" as []User
// Output: Report
func ReportRelationAssignees() mongo.Pipeline {
	sp := mongo.Pipeline{
		{{
			Key: "$match",
			Value: bson.M{
				"$expr": bson.M{
					"$in": bson.A{"$_id", "$$user_ids"},
				},
			},
		}},
	}
	sp = append(sp, UserRelationRoles...)

	p := []bson.D{
		{{
			Key: "$lookup",
			Value: mongo.LookupWithPipeline{
				From:     mongo.CollectionNameUsers,
				Let:      bson.M{"user_ids": "$assignee_ids"},
				Pipeline: &sp,
				As:       "assignees",
			},
		}},
	}

	return p
}
