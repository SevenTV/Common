package redis

import (
	"time"

	"github.com/go-redsync/redsync/v4"
)

func (inst *redisInst) Mutex(name Key, ex time.Duration) *redsync.Mutex {
	if inst.sync == nil {
		return nil // sync is disabled
	}

	return inst.sync.NewMutex(name.String(), redsync.WithExpiry(ex))
}
