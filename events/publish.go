package events

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/seventv/common/redis"
)

func Publish[D AnyPayload](ctx context.Context, msg Message[D], redis redis.Instance) error {
	j, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	k := redis.ComposeKey("events", "op", strings.ToLower(msg.Op.String()))
	if _, err = redis.RawClient().Publish(ctx, k.String(), j).Result(); err != nil {
		return err
	}
	return nil
}
