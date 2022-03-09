package query

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (q *Query) Users(ctx context.Context, filter bson.M) ([]*structures.User, error) {
	items := []*structures.User{}
	cur, err := q.mongo.Collection(mongo.CollectionNameUsers).Aggregate(ctx, mongo.Pipeline{
		{{
			Key:   "$match",
			Value: filter,
		}},
		{{
			Key: "$group",
			Value: bson.M{
				"_id": nil,
				"users": bson.M{
					"$push": "$$ROOT",
				},
			},
		}},
		{{
			Key: "$lookup",
			Value: mongo.Lookup{
				From:         mongo.CollectionNameEntitlements,
				LocalField:   "_id",
				ForeignField: "user_id",
				As:           "role_entitlements",
			},
		}},
		{{
			Key: "$set",
			Value: bson.M{
				"role_entitlements": bson.M{
					"$filter": bson.M{
						"input": "$role_entitlements",
						"as":    "ent",
						"cond": bson.M{
							"$eq": bson.A{"$$ent.kind", structures.EntitlementKindRole},
						},
					},
				},
			},
		}},
	})
	if err != nil {
		logrus.WithError(err).Error("query, failed to execute find")
		return items, err
	}

	// Get roles
	roles, _ := q.Roles(ctx, bson.M{})
	roleMap := make(map[primitive.ObjectID]*structures.Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	// Map all objects
	cur.Next(ctx)
	v := &aggregatedUsersResult{}
	if err = cur.Decode(v); err != nil {
		return items, err
	}

	userMap := make(map[primitive.ObjectID]*structures.User)
	entRoleMap := make(map[primitive.ObjectID][]primitive.ObjectID)
	for _, ent := range v.RoleEntitlements {
		ref := ent.GetData().ReadRole()
		if ref == nil {
			continue
		}
		entRoleMap[ent.UserID] = append(entRoleMap[ent.UserID], ref.ObjectReference)
	}
	for _, user := range v.Users {
		user.RoleIDs = append(user.RoleIDs, entRoleMap[user.ID]...)
		userMap[user.ID] = user
	}

	for _, u := range v.Users { // iterare over users
		// add user roles
		for _, roleID := range u.RoleIDs {
			role, rok := roleMap[roleID]
			if !rok {
				continue
			}
			u.Roles = append(u.Roles, role)
		}
		items = append(items, u)
	}

	if err = multierror.Append(err, cur.Close(ctx)).ErrorOrNil(); err != nil {
		logrus.WithError(err).Error("query, failed to close the cursor")
	}
	return items, nil
}

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

type aggregatedUsersResult struct {
	Users            []*structures.User        `bson:"users"`
	RoleEntitlements []*structures.Entitlement `bson:"role_entitlements"`
}
