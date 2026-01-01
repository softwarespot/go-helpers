package ack

type beforeContextBuffer struct {
	lines []*MatchContext
	size  int
	head  int
	tail  int
}

func newBeforeContextBuffer(maxSize int) *beforeContextBuffer {
	return &beforeContextBuffer{
		lines: make([]*MatchContext, maxSize),
		size:  0,
		head:  0,
		tail:  0,
	}
}

func (b *beforeContextBuffer) add(line *MatchContext) {
	maxSize := len(b.lines)
	if maxSize == 0 {
		return
	}

	b.lines[b.tail] = line
	b.tail = (b.tail + 1) % maxSize

	if b.size == maxSize {
		b.head = (b.head + 1) % maxSize
	} else {
		b.size++
	}
}

func (b *beforeContextBuffer) lastN(n int) []*MatchContext {
	maxSize := len(b.lines)
	if n <= 0 || maxSize == 0 {
		return nil
	}

	if n > b.size {
		n = b.size
	}

	ms := make([]*MatchContext, 0, n)

	head := (b.tail - n + maxSize) % maxSize
	for i := 0; i < n; i++ {
		idx := (head + i) % maxSize
		ms = append(ms, b.lines[idx])
	}
	return ms
}
