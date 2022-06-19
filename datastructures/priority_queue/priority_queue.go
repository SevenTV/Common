package priority_queue

import "github.com/seventv/common/datastructures/heap"

type Item[T any] struct {
	priority int
	idx      int
	value    T
}

func (i Item[T]) Index() int {
	return i.idx
}

func (i *Item[T]) SetIndex(idx int) {
	i.idx = idx
}

func (i Item[T]) Rank() int {
	return -i.priority
}

func (i Item[T]) Priority() int {
	return i.priority
}

func (i Item[T]) Value() T {
	return i.value
}

type PriorityQueue[T any] struct {
	heap.Heap[*Item[T]]
}

func (pq *PriorityQueue[T]) Push(item T, priority int) *Item[T] {
	itm := &Item[T]{
		priority: priority,
		value:    item,
		idx:      -1,
	}

	pq.Heap.Push(itm)

	return itm
}
