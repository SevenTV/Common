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

	if roles, err := qb.q.Roles(qb.ctx, bson.M{}); err == nil && len(roles) > 0 {
		roleMap := make(map[primitive.ObjectID]*structures.Role)
		for _, r := range roles {
			roleMap[r.ID] = r
		}
		for _, u := range m {
			u.RoleIDs = append(u.RoleIDs, m2[u.ID]...)
			for _, roleID := range u.RoleIDs {
				if role, ok := roleMap[roleID]; ok {
					u.Roles = append(u.Roles, role)
				}
			}
		}
	}

	return m
}
