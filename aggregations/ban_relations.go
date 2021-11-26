package aggregations

import (
	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// Ban Relations
//
// Input: Ban
// Adds Field: "victim" as User
// Output: Ban
var BanRelationVictim = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameUsers,
			LocalField:   "victim_id",
			ForeignField: "_id",
			As:           "victims",
		},
	}},
	{{
		Key: "$set",
		Value: bson.M{
			"victim": bson.M{"$first": "$victims"},
		},
	}},
}

// Ban Relations
//
// Input: Ban
// Adds Field: "actor" as User
// Output: Ban
var BanRelationActor = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameUsers,
			LocalField:   "actor_id",
			ForeignField: "_id",
			As:           "actors",
		},
	}},
	{{
		Key: "$set",
		Value: bson.M{
			"actor": bson.M{"$first": "$actors"},
		},
	}},
}
