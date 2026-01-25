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

func (rb *RingBuffer[T]) N(n int) []T {
	if n > 0 {
		return rb.FirstN(n)
	}
	return rb.LastN(-n)
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

func (rb *RingBuffer[T]) IterN(n int) iter.Seq2[int, T] {
	if n > 0 {
		return rb.IterFirstN(n)
	}
	return rb.IterLastN(-n)
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

func (rb *RingBuffer[T]) Push(item T) (T, bool) {
	maxSize := len(rb.items)
	isFull := rb.size == maxSize

	prevItem := rb.items[rb.tail]
	rb.items[rb.tail] = item

	if isFull {
		rb.head = (rb.head + 1) % maxSize
	} else {
		rb.size++
	}
	rb.tail = (rb.tail + 1) % maxSize

	return prevItem, isFull
}

func (rb *RingBuffer[T]) Pop() (T, bool) {
	var zero T
	if rb.size == 0 {
		return zero, false
	}

	maxSize := len(rb.items)
	rb.tail = (rb.tail - 1 + maxSize) % maxSize
	item := rb.items[rb.tail]
	rb.items[rb.tail] = zero
	rb.size--

	return item, true
}

func (rb *RingBuffer[T]) PopN(n int) []T {
	var items []T
	for range n {
		item, ok := rb.Pop()
		if !ok {
			break
		}
		items = append(items, item)
	}
	return items
}

func (rb *RingBuffer[T]) PeekFront() (T, bool) {
	if rb.size == 0 {
		var zero T
		return zero, false
	}
	return rb.items[rb.head], true
}

func (rb *RingBuffer[T]) Unshift(item T) (T, bool) {
	maxSize := len(rb.items)
	isFull := rb.size == maxSize

	rb.head = (rb.head - 1 + maxSize) % maxSize
	prevItem := rb.items[rb.head]
	rb.items[rb.head] = item

	if isFull {
		rb.tail = (rb.tail - 1 + maxSize) % maxSize
	} else {
		rb.size++
	}
	return prevItem, isFull
}

func (rb *RingBuffer[T]) Shift() (T, bool) {
	var zero T
	if rb.size == 0 {
		return zero, false
	}

	maxSize := len(rb.items)
	item := rb.items[rb.head]
	rb.items[rb.head] = zero
	rb.head = (rb.head + 1) % maxSize
	rb.size--

	return item, true
}

func (rb *RingBuffer[T]) ShiftN(n int) []T {
	var items []T
	for range n {
		item, ok := rb.Shift()
		if !ok {
			break
		}
		items = append(items, item)
	}
	return items
}

func (rb *RingBuffer[T]) PeekBack() (T, bool) {
	if rb.size == 0 {
		var zero T
		return zero, false
	}

	maxSize := len(rb.items)
	idx := (rb.tail - 1 + maxSize) % maxSize
	return rb.items[idx], true
}

func (rb *RingBuffer[T]) Capacity() int {
	return len(rb.items)
}

func (rb *RingBuffer[T]) Size() int {
	return rb.size
}

func (rb *RingBuffer[T]) IsEmpty() bool {
	return rb.size == 0
}

func (rb *RingBuffer[T]) IsFull() bool {
	return rb.size == len(rb.items)
}

func (rb *RingBuffer[T]) Reset() {
	clear(rb.items)
	rb.size = 0
	rb.head = 0
	rb.tail = 0
}
