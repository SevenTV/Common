package query

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func (q *Query) Roles(ctx context.Context, filter bson.M) ([]*structures.Role, error) {
	mx := q.lock("ManyRoles")
	defer mx.Unlock()

	f, _ := json.Marshal(filter)
	h := sha256.New()
	h.Write(f)
	k := q.key(fmt.Sprintf("roles:%s", hex.EncodeToString(h.Sum(nil))))
	result := []*structures.Role{}

	// Get cached
	if ok := q.getFromMemCache(ctx, k, &result); ok {
		return result, nil
	}

	// Query
	cur, err := q.mongo.Collection(mongo.CollectionNameRoles).Find(ctx, filter)
	if err == nil {
		if err = cur.All(ctx, &result); err != nil {
			return nil, err
		}
	}

	// Set cache
	if err = q.setInMemCache(ctx, k, &result, time.Second*10); err != nil {
		return nil, err
	}
	return result, nil
}

type ManyRolesOptions struct {
	DefaultOnly bool
}
