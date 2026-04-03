package ringbuffer

import (
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_New(t *testing.T) {
	rb := New[string](5)

	// Add events to the ring buffer
	prevItem, isFull := rb.Push("Event 2")
	testhelpers.AssertEqual(t, prevItem, "")
	testhelpers.AssertEqual(t, isFull, false)

	prevItem, isFull = rb.Unshift("Event 1")
	testhelpers.AssertEqual(t, prevItem, "")
	testhelpers.AssertEqual(t, isFull, false)

	testhelpers.AssertEqual(t, rb.All(), []string{
		"Event 1",
		"Event 2",
	})
	testhelpers.AssertEqual(t, rb.Capacity(), 5)
	testhelpers.AssertEqual(t, rb.Size(), 2)
	testhelpers.AssertEqual(t, rb.IsEmpty(), false)
	testhelpers.AssertEqual(t, rb.IsFull(), false)

	// Add more events to the ring buffer
	prevItem, isFull = rb.Push("Event 3")
	testhelpers.AssertEqual(t, prevItem, "")
	testhelpers.AssertEqual(t, isFull, false)

	prevItem, isFull = rb.Push("Event 4")
	testhelpers.AssertEqual(t, prevItem, "")
	testhelpers.AssertEqual(t, isFull, false)

	prevItem, isFull = rb.Push("Event 5")
	testhelpers.AssertEqual(t, prevItem, "")
	testhelpers.AssertEqual(t, isFull, false)

	prevItem, isFull = rb.Push("Event 6")
	testhelpers.AssertEqual(t, prevItem, "Event 1")
	testhelpers.AssertEqual(t, isFull, true)

	prevItem, isFull = rb.Unshift("Event 7")
	testhelpers.AssertEqual(t, prevItem, "Event 6")
	testhelpers.AssertEqual(t, isFull, true)

	testhelpers.AssertEqual(t, rb.FirstN(10), []string{
		"Event 7",
		"Event 2",
		"Event 3",
		"Event 4",
		"Event 5",
	})
	testhelpers.AssertEqual(t, rb.LastN(3), []string{
		"Event 3",
		"Event 4",
		"Event 5",
	})

	testhelpers.AssertEqual(t, rb.N(10), []string{
		"Event 7",
		"Event 2",
		"Event 3",
		"Event 4",
		"Event 5",
	})
	testhelpers.AssertEqual(t, rb.N(-3), []string{
		"Event 3",
		"Event 4",
		"Event 5",
	})

	var items []string
	for _, item := range rb.IterN(3) {
		items = append(items, item)
	}
	testhelpers.AssertEqual(t, items, []string{
		"Event 7",
		"Event 2",
		"Event 3",
	})

	items = nil

	for _, item := range rb.IterN(-2) {
		items = append(items, item)
	}
	testhelpers.AssertEqual(t, items, []string{
		"Event 4",
		"Event 5",
	})

	testhelpers.AssertEqual(t, rb.All(), []string{
		"Event 7",
		"Event 2",
		"Event 3",
		"Event 4",
		"Event 5",
	})
	testhelpers.AssertEqual(t, rb.Capacity(), 5)
	testhelpers.AssertEqual(t, rb.Size(), 5)
	testhelpers.AssertEqual(t, rb.IsEmpty(), false)
	testhelpers.AssertEqual(t, rb.IsFull(), true)

	item, ok := rb.PeekBack()
	testhelpers.AssertEqual(t, item, "Event 5")
	testhelpers.AssertEqual(t, ok, true)

	items = rb.PopN(2)
	testhelpers.AssertEqual(t, items, []string{
		"Event 5",
		"Event 4",
	})

	testhelpers.AssertEqual(t, rb.All(), []string{
		"Event 7",
		"Event 2",
		"Event 3",
	})
	testhelpers.AssertEqual(t, rb.Capacity(), 5)
	testhelpers.AssertEqual(t, rb.Size(), 3)
	testhelpers.AssertEqual(t, rb.IsEmpty(), false)
	testhelpers.AssertEqual(t, rb.IsFull(), false)

	item, ok = rb.PeekFront()
	testhelpers.AssertEqual(t, item, "Event 7")
	testhelpers.AssertEqual(t, ok, true)

	items = rb.ShiftN(2)
	testhelpers.AssertEqual(t, items, []string{
		"Event 7",
		"Event 2",
	})

	testhelpers.AssertEqual(t, rb.All(), []string{
		"Event 3",
	})
	testhelpers.AssertEqual(t, rb.Capacity(), 5)
	testhelpers.AssertEqual(t, rb.Size(), 1)
	testhelpers.AssertEqual(t, rb.IsEmpty(), false)
	testhelpers.AssertEqual(t, rb.IsFull(), false)

	// Should reset the ring buffer
	rb.Reset()

	item, ok = rb.PeekBack()
	testhelpers.AssertEqual(t, item, "")
	testhelpers.AssertEqual(t, ok, false)

	items = rb.PopN(2)
	testhelpers.AssertEqual(t, items, nil)

	item, ok = rb.Pop()
	testhelpers.AssertEqual(t, item, "")
	testhelpers.AssertEqual(t, ok, false)

	item, ok = rb.PeekFront()
	testhelpers.AssertEqual(t, item, "")
	testhelpers.AssertEqual(t, ok, false)

	items = rb.ShiftN(2)
	testhelpers.AssertEqual(t, items, nil)

	item, ok = rb.Shift()
	testhelpers.AssertEqual(t, item, "")
	testhelpers.AssertEqual(t, ok, false)

	testhelpers.AssertEqual(t, rb.All(), nil)
	testhelpers.AssertEqual(t, rb.Capacity(), 5)
	testhelpers.AssertEqual(t, rb.Size(), 0)
	testhelpers.AssertEqual(t, rb.IsEmpty(), true)
	testhelpers.AssertEqual(t, rb.IsFull(), false)
}
