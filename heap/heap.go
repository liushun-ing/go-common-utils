package heap

import "container/heap"

// Element 自定义 Element 接口，支持自定义排序规则和可用规则
type Element interface {
	Less(t Element) bool
	IsUsable() bool
}

type Heap[T Element] []T

func (h *Heap[T]) Init() {
	heap.Init(h)
}

func (h *Heap[T]) PushOne(one T) {
	heap.Push(h, one)
}

func (h *Heap[T]) PopOne() (T, bool) {
	if h.Len() == 0 {
		var zero T
		return zero, false
	}

	top := (*h)[0]
	if top.IsUsable() {
		return heap.Pop(h).(T), true
	}

	var zero T
	return zero, false
}

func (h *Heap[T]) Len() int {
	return len(*h)
}

func (h *Heap[T]) Less(i, j int) bool {
	return (*h)[i].Less((*h)[j])
}

func (h *Heap[T]) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *Heap[T]) Push(x interface{}) {
	(*h) = append(*h, x.(T))
}

func (h *Heap[T]) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
