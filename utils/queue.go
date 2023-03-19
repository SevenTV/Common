package utils

import "sync"

type Queue[T any] interface {
	Add(val T)
	Pop() T
	Items() []T
	Clear()
	IsFull() bool
	IsEmpty() bool
}

type queue[T any] struct {
	mx       sync.Mutex
	values   []T
	capacity int
}

func NewQueue[T any](capacity int) Queue[T] {
	return &queue[T]{
		mx:       sync.Mutex{},
		values:   make([]T, 0),
		capacity: capacity,
	}
}

func (q *queue[T]) Add(val T) {
	q.mx.Lock()
	defer q.mx.Unlock()

	q.values = append(q.values, val)
}

func (q *queue[T]) Pop() T {
	var val T

	q.mx.Lock()
	defer q.mx.Unlock()

	if len(q.values) == 0 {
		return val
	}

	val = q.values[0]
	q.values = q.values[1:]

	return val
}

func (q *queue[T]) Items() []T {
	return q.values
}

func (q *queue[T]) Clear() {
	q.mx.Lock()
	defer q.mx.Unlock()

	q.values = make([]T, 0)
}

func (q *queue[T]) IsFull() bool {
	return len(q.values) >= q.capacity
}

func (q *queue[T]) IsEmpty() bool {
	return len(q.values) == 0
}
