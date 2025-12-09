package ringbuffer

import "iter"

type RingBuffer[T any] struct {
	items []T
	size  int
	head  int
	tail  int
}

func New[T any](maxSize int) *RingBuffer[T] {
	if maxSize <= 0 {
		panic("ringbuffer.New: maxSize must be greater than 0")
	}
	return &RingBuffer[T]{
		items: make([]T, maxSize),
		size:  0,
		head:  0,
		tail:  0,
	}
}

func (rb *RingBuffer[T]) All() []T {
	return rb.FirstN(rb.size)
}

func (rb *RingBuffer[T]) Iter() iter.Seq2[int, T] {
	return rb.IterFirstN(rb.size)
}

func (rb *RingBuffer[T]) FirstN(n int) []T {
	var items []T
	for _, item := range rb.IterFirstN(n) {
		items = append(items, item)
	}
	return items
}

func (rb *RingBuffer[T]) LastN(n int) []T {
	var items []T
	for _, item := range rb.IterLastN(n) {
		items = append(items, item)
	}
	return items
}

func (rb *RingBuffer[T]) IterFirstN(n int) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		if n <= 0 {
			return
		}
		if n > rb.size {
			n = rb.size
		}

		maxSize := len(rb.items)
		for i := 0; i < n; i++ {
			idx := (rb.head + i) % maxSize
			if !yield(i, rb.items[idx]) {
				return
			}
		}
	}
}

func (rb *RingBuffer[T]) IterLastN(n int) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		if n <= 0 {
			return
		}
		if n > rb.size {
			n = rb.size
		}

		maxSize := len(rb.items)
		head := (rb.tail - n + maxSize) % maxSize
		for i := 0; i < n; i++ {
			idx := (head + i) % maxSize
			if !yield(i, rb.items[idx]) {
				return
			}
		}
	}
}

func (rb *RingBuffer[T]) Add(item T) (T, bool) {
	maxSize := len(rb.items)
	prev := rb.items[rb.tail]
	isFull := rb.size == maxSize

	rb.items[rb.tail] = item
	rb.tail = (rb.tail + 1) % maxSize

	if isFull {
		rb.head = (rb.head + 1) % maxSize
	} else {
		rb.size++
	}
	return prev, isFull
}

func (rb *RingBuffer[T]) Size() int {
	return rb.size
}

func (rb *RingBuffer[T]) Clear() {
	clear(rb.items)
	rb.size = 0
	rb.head = 0
	rb.tail = 0
}
