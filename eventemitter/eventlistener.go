package eventemitter

import "reflect"

type EventListener struct {
	channels map[string]reflect.Value
}

func NewEventListener(channels map[string]reflect.Value) *EventListener {
	return &EventListener{channels: channels}
}

func (e *EventListener) publishRaw(event string, payload any) bool {
	if ch, ok := e.channels[event]; ok && ch.Kind() == reflect.Chan {
		ch.Send(reflect.ValueOf(payload))

		return true
	}

	return false
}
