package heap

type Heapable interface {
	Rank() int
	Index() int

	SetIndex(idx int)
}

type Heap[T Heapable] []T

func (h Heap[T]) Len() int {
	return len(h)
}

func (h *Heap[T]) Pop() T {
	n := h.Len() - 1
	h.Swap(0, n)
	h.down(0, n)

	return h.pop()
}

func (h *Heap[T]) pop() T {
	old := *h
	n := len(old)
	item := old[n-1]
	item.SetIndex(-1)
	*h = old[:n-1]

	return item
}

func (h *Heap[T]) Push(item T) {
	item.SetIndex(h.Len())

	*h = append(*h, item)

	h.up(h.Len() - 1)
}

func (h Heap[T]) Less(i, j int) bool {
	return h[i].Rank() < h[j].Rank()
}

func (h Heap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]

	h[i].SetIndex(i)
	h[j].SetIndex(j)
}

func (h Heap[T]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func (pq Heap[T]) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && pq.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !pq.Less(j, i) {
			break
		}
		pq.Swap(i, j)
		i = j
	}
	return i > i0
}

func (h Heap[T]) Fix(i int) {
	if !h.down(i, h.Len()) {
		h.up(i)
	}
}

func (h Heap[T]) Init() {
	// heapify
	n := h.Len()
	for i := n/2 - 1; i >= 0; i-- {
		h.down(i, n)
	}
}

func (pq *Heap[T]) Remove(i int) T {
	n := pq.Len() - 1
	if n != i {
		pq.Swap(i, n)
		if !pq.down(i, n) {
			pq.up(i)
		}
	}
	return pq.pop()
}
