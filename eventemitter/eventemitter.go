package eventemitter

import (
	"sync"
	"time"

	"github.com/SevenTV/Common/utils"
)

type RawEventEmitter struct {
	mtx  sync.Mutex
	once sync.Once
	done chan struct{}
	mp   map[string]*container
}

type container struct {
	i   uint64
	evt string
	mtx sync.Mutex
	mp  map[uint64]*EventListener
}

func (c *container) listen(l *EventListener) func() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	id := c.i
	c.i++

	c.mp[id] = l

	return func() {
		c.mtx.Lock()
		defer c.mtx.Unlock()
		delete(c.mp, id)
	}
}

func (c *container) publish(payload any) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, v := range c.mp {
		v.publishRaw(c.evt, payload)
	}
}

func New() *RawEventEmitter {
	e := &RawEventEmitter{
		mp:   map[string]*container{},
		done: make(chan struct{}),
	}

	go e.clean()

	return e
}

func (e *RawEventEmitter) Listen(l *EventListener) func() {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	unbindFns := make([]func(), len(l.channels))

	i := 0
	for evt := range l.channels {
		cn, ok := e.mp[evt]
		if !ok {
			cn = &container{
				evt: evt,
				mp:  map[uint64]*EventListener{},
			}
			e.mp[evt] = cn
		}

		unbindFns[i] = cn.listen(l)
		i++
	}

	return func() {
		for _, fn := range unbindFns {
			fn()
		}
	}

}

func (e *RawEventEmitter) PublishRaw(event string, payload any) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if cn, ok := e.mp[event]; ok {
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
			e.mtx.Lock()
			for evt, cn := range e.mp {
				cn.mtx.Lock()
				if len(cn.mp) == 0 {
					delete(e.mp, evt)
				}
				cn.mtx.Unlock()
			}
			e.mtx.Unlock()
		}
	}
}

func (e *RawEventEmitter) Stop() {
	e.once.Do(func() {
		close(e.done)
	})
}
