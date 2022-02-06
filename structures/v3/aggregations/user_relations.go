package aggregations

import (
	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// User Relations
//
// Input: User
// Adds Field: "roles" as []Role
// Output: User
var UserRelationRoles = []bson.D{
	// Step 1: Lookup ROLE entitlements matching the input user
	{{
		Key: "$lookup",
		Value: mongo.LookupWithPipeline{
			From: mongo.CollectionNameEntitlements,
			Let:  bson.M{"user_id": "$_id"},
			Pipeline: &mongo.Pipeline{
				bson.D{{
					Key: "$match",
					Value: bson.M{
						"disabled": bson.M{"$not": bson.M{"$eq": true}},
						"kind":     "ROLE",
						"$expr": bson.M{
							"$eq": bson.A{"$user_id", "$$user_id"},
						},
					},
				}},
			},
			As: "role_entitlements",
		},
	}},
	// Step 2: Update the "role_ids" field combining the original value + entitled roles
	{{
		Key: "$set",
		Value: bson.M{
			"role_ids": bson.M{
				"$concatArrays": bson.A{"$role_ids", "$role_entitlements.data.ref"},
			},
		},
	}},
	// Step 3: Unset the temporary "role_entitlements" field
	{{Key: "$unset", Value: bson.A{"role_entitlements"}}},
	// Step 4: Lookup roles matching the newly defined role IDs and output them as "roles", an array of Role
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameRoles,
			LocalField:   "role_ids",
			ForeignField: "_id",
			As:           "roles",
		},
	}},
}

// User Relations
//
// Input: User
// Adds Field: "editors" as []UserEditor with the "user" field added to each UserEditor object
// Output: User
var UserRelationEditors = []bson.D{
	// Step 1: Lookup user editors
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameUsers,
			LocalField:   "editors.id",
			ForeignField: "_id",
			As:           "_ed",
		},
	}},
	{{
		Key: "$set",
		Value: bson.M{
			"editors": bson.M{
				"$map": bson.M{
					"input": "$editors",
					"in": bson.M{
						"$mergeObjects": bson.A{
							"$$this",
							bson.M{
								"user": bson.M{
									"$arrayElemAt": bson.A{
										"$_ed",
										bson.M{"$indexOfArray": bson.A{"$_ed._id", "$$this.id"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}},
}

// User Relations
//
// Input: User
// Add Fields: "editor_of" as []UserEditor with the "user" field added to each UserEditor object
// Output: User
var UserRelationEditorOf = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.LookupWithPipeline{
			From: mongo.CollectionNameUsers,
			Let:  bson.M{"user_id": "$_id"},
			Pipeline: &mongo.Pipeline{
				{{
					Key: "$match",
					Value: bson.M{
						"$expr": bson.M{"$in": bson.A{"$$user_id", "$editors.id"}},
					},
				}},
				{{
					Key: "$project",
					Value: bson.M{
						"user": "$$ROOT",
						"as_editor": bson.M{
							"$mergeObjects": bson.M{
								"$filter": bson.M{
									"input": "$editors",
									"as":    "u_editor_of",
									"cond":  bson.M{"$eq": bson.A{"$$u_editor_of.id", "$$user_id"}},
								},
							},
						},
					},
				}},
				{{
					Key: "$project",
					Value: bson.M{
						"id":          "$_id", // Replace the _id field with "id"
						"permissions": "$as_editor.permissions",
						"connections": "$as_editor.connections",
						"visible":     "$as_editor.visible",
						"user":        "$user",
					},
				}},
				{{Key: "$unset", Value: bson.A{"_id"}}}, // Remove the "_id" field (it's id)
			},
			As: "editor_of",
		},
	}},
}

// User Owned Emote Relations
//
// Input: User
// Adds Field: "owned_emotes" as []Emote
// Output: User
var UserRelationOwnedEmotes = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         mongo.CollectionNameEmotes,
			LocalField:   "_id",
			ForeignField: "owner_id",
			As:           "owned_emotes",
		},
	}},
}

var UserRelationConnections = []bson.D{
	{{
		Key: "$lookup",
		Value: mongo.Lookup{
			From:         "user_connections",
			LocalField:   "connection_ids",
			ForeignField: "_id",
			As:           "connections",
		},
	}},
}

// User Relations
// Input: User
// Adds Field: "emote_set" of []connections as EmoteSet
func UserConnectionEmoteSetRelation() mongo.Pipeline {
	return mongo.Pipeline{
		{{
			Key: "$lookup",
			Value: mongo.LookupWithPipeline{
				From: mongo.CollectionNameEmoteSets,
				Let:  bson.M{"set_id": "$connections.emote_set_id"},
				Pipeline: &mongo.Pipeline{
					{{
						Key: "$match",
						Value: bson.M{
							"$expr": bson.M{
								"$in": bson.A{"$_id", "$$set_id"},
							},
						},
					}},
					{{
						Key: "$lookup",
						Value: mongo.LookupWithPipeline{
							From: mongo.CollectionNameEmotes,
							Let:  bson.M{"emote_ids": "$emotes.id"},
							Pipeline: CombinePtr(
								mongo.Pipeline{
									{{
										Key: "$match",
										Value: bson.M{"$expr": bson.M{
											"$in": bson.A{"$_id", "$$emote_ids"},
										}},
									}},
									{{
										Key: "$lookup",
										Value: mongo.LookupWithPipeline{
											From: mongo.CollectionNameUsers,
											Let:  bson.M{"owner_id": "$owner_id"},
											Pipeline: CombinePtr(
												mongo.Pipeline{{{
													Key:   "$match",
													Value: bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$owner_id"}}}},
												}},
												UserRelationRoles,
											),
											As: "owner_user",
										},
									}},
									{{
										Key:   "$set",
										Value: bson.M{"owner_user": bson.M{"$first": "$owner_user"}},
									}},
								},
							),
							As: "_emotes",
						},
					}},
					MergeArrays("emotes", "$_emotes", "_id", "emote"),
					{{Key: "$unset", Value: "_emotes"}},
				},
				As: "_sets",
			},
		}},
		MergeArrays("connections", "$_sets", "_id", "emote_set"),
		{{Key: "$unset", Value: "_sets"}},
	}
}

// Emote Relations
//
// Input: Emote
// Adds Field: "owner" as User
// Output: Emote
func GetEmoteRelationshipOwner(opt UserRelationshipOptions) []bson.D {
	up := mongo.Pipeline{
		bson.D{{
			Key: "$match",
			Value: bson.M{
				"$expr": bson.M{"$eq": bson.A{"$_id", "$$owner_id"}},
			},
		}},
	}
	if opt.Editors {
		up = append(up, UserRelationEditors...)
	}
	if opt.Roles {
		up = append(up, UserRelationRoles...)
	}

	p := mongo.Pipeline{
		// Step 1: Lookup emote owners
		{{
			Key: "$lookup",
			Value: mongo.LookupWithPipeline{
				From:     mongo.CollectionNameUsers,
				Let:      bson.M{"owner_id": "$owner_id"},
				Pipeline: &up,
				As:       "owner_user",
			},
		}},
		{{
			Key: "$set",
			Value: bson.M{
				"owner_user": bson.M{
					"$first": "$owner_user",
				},
			},
		}},
	}

	return p
}

type UserRelationshipOptions struct {
	Editors       bool
	Roles         bool
	ChannelEmotes bool
}
