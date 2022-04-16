package eventemitter

import (
	"reflect"

	"github.com/SevenTV/Common/sync_map"
)

type EventListener struct {
	channels *sync_map.Map[string, reflect.Value]
}

func NewEventListener(channels map[string]reflect.Value) *EventListener {
	return &EventListener{channels: sync_map.FromStdMap(channels)}
}

func (e *EventListener) publishRaw(event string, payload any) bool {
	if ch, ok := e.channels.Load(event); ok && ch.Kind() == reflect.Chan {
		ch.Send(reflect.ValueOf(payload))

		return true
	}

	return false
}
