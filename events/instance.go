package events

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/seventv/common/redis"
)

type Instance interface {
	Publish(ctx context.Context, msg Message[json.RawMessage]) error
}

type eventsInst struct {
	ctx   context.Context
	redis redis.Instance
}

func NewPublisher(ctx context.Context, redis redis.Instance) Instance {
	return &eventsInst{
		ctx:   ctx,
		redis: redis,
	}
}

func (inst *eventsInst) Publish(ctx context.Context, msg Message[json.RawMessage]) error {
	j, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	k := inst.redis.ComposeKey("events", "op", strings.ToLower(msg.Op.String()))
	if _, err = inst.redis.RawClient().Publish(ctx, k.String(), j).Result(); err != nil {
		return err
	}
	return nil
}
