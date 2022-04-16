package eventemitter

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SevenTV/Common/sync_map"
	"github.com/SevenTV/Common/utils"
)

type RawEventEmitter struct {
	once sync.Once
	done chan struct{}
	sMp  sync_map.Map[string, *container]
}

type container struct {
	i   *uint64
	evt string
	sMp sync_map.Map[uint64, *EventListener]
}

func (c *container) listen(l *EventListener) func() {
	id := atomic.AddUint64(c.i, 1)

	c.sMp.Store(id, l)

	return func() {
		c.sMp.Delete(id)
	}
}

func (c *container) publish(payload any) {
	c.sMp.Range(func(key uint64, value *EventListener) bool {
		value.publishRaw(c.evt, payload)
		return true
	})
}

func New() *RawEventEmitter {
	e := &RawEventEmitter{
		done: make(chan struct{}),
	}

	go e.clean()

	return e
}

func (e *RawEventEmitter) Listen(l *EventListener) func() {
	unbindFns := []func(){}

	l.channels.Range(func(evt string, value reflect.Value) bool {
		cn, ok := e.sMp.Load(evt)
		if !ok {
			cn = &container{
				i:   utils.PointerOf(uint64(0)),
				evt: evt,
			}
			e.sMp.Store(evt, cn)
		}

		unbindFns = append(unbindFns, cn.listen(l))

		return true
	})

	return func() {
		for _, fn := range unbindFns {
			fn()
		}
	}

}

func (e *RawEventEmitter) PublishRaw(event string, payload any) {
	if cn, ok := e.sMp.Load(event); ok {
		cn.publish(payload)
	}
}

func (e *RawEventEmitter) clean() {
	tick := time.NewTicker(time.Minute*30 + utils.JitterTime(time.Second, time.Minute))
	defer tick.Stop()

	for {
		select {
		case <-e.done:
			return
		case <-tick.C:
			e.sMp.Range(func(evt string, cn *container) bool {
				i := 0
				cn.sMp.Range(func(key uint64, value *EventListener) bool {
					i++
					return false
				})
				if i == 0 {
					e.sMp.Delete(evt)
				}

				return true
			})
		}
	}
}

func (e *RawEventEmitter) Stop() {
	e.once.Do(func() {
		close(e.done)
	})
}
