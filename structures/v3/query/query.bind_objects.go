package query

import (
	"context"

	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QueryBinder struct {
	ctx context.Context
	q   *Query
}

func (q *Query) NewBinder(ctx context.Context) *QueryBinder {
	return &QueryBinder{ctx, q}
}

func (qb *QueryBinder) mapUsers(users []*structures.User, roleEnts ...*structures.Entitlement) map[primitive.ObjectID]*structures.User {
	m := make(map[primitive.ObjectID]*structures.User)
	for _, v := range users {
		m[v.ID] = v
	}
	m2 := make(map[primitive.ObjectID][]primitive.ObjectID)
	for _, ent := range roleEnts {
		ref := ent.GetData().ReadRole()
		if ref == nil {
			continue
		}
		m2[ent.UserID] = append(m2[ent.UserID], ref.ObjectReference)
	}

	roles, _ := qb.q.Roles(qb.ctx, bson.M{})
	if len(roles) > 0 {
		roleMap := make(map[primitive.ObjectID]*structures.Role)
		var defaultRole *structures.Role
		for _, r := range roles {
			if r.Default {
				defaultRole = r
			}
			roleMap[r.ID] = r
		}
		for _, u := range m {
			roleIDs := make([]primitive.ObjectID, len(m2[u.ID])+len(u.RoleIDs)+1)
			switch defaultRole != nil { // add default role, or if no default role add nil role
			case true:
				roleIDs[0] = defaultRole.ID
			case false:
				roleIDs[0] = structures.NilRole.ID
			}
			roleIDs[0] = defaultRole.ID
			copy(roleIDs[1:], u.RoleIDs)
			copy(roleIDs[len(u.RoleIDs)+1:], m2[u.ID])

			u.Roles = make([]*structures.Role, len(roleIDs)) // allocate space on the user's roles slice
			for i, roleID := range roleIDs {
				if role, ok := roleMap[roleID]; ok { // add role if exists
					u.Roles[i] = role
				} else {
					u.Roles[i] = structures.NilRole // set nil role if role wasn't found
				}
			}
		}
	}

	return m
}
