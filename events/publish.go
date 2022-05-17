package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/SevenTV/Common/redis"
)

func PublishDispatch(ctx context.Context, msg Message[DispatchPayload], redis redis.Instance) error {
	t := strings.Split(string(msg.Data.Type), ".")
	if len(t) == 0 {
		return fmt.Errorf("internal: invalid dispatch type")
	}

	k := redis.ComposeKey("events", strconv.Itoa(int(msg.Op)), string(msg.Data.Type))
	j, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err = redis.RawClient().Publish(ctx, k.String(), j).Result(); err != nil {
		return err
	}
	return nil
}
