package ringbuffer

import (
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_New(t *testing.T) {
	rb := New[string](5)

	// Add events to the ring buffer
	prev, isFull := rb.Add("Event 1")
	testhelpers.AssertEqual(t, prev, "")
	testhelpers.AssertEqual(t, isFull, false)

	prev, isFull = rb.Add("Event 2")
	testhelpers.AssertEqual(t, prev, "")
	testhelpers.AssertEqual(t, isFull, false)

	testhelpers.AssertEqual(t, rb.All(), []string{
		"Event 1",
		"Event 2",
	})
	testhelpers.AssertEqual(t, rb.Size(), 2)

	// Add more events to the ring buffer
	prev, isFull = rb.Add("Event 3")
	testhelpers.AssertEqual(t, prev, "")
	testhelpers.AssertEqual(t, isFull, false)

	prev, isFull = rb.Add("Event 4")
	testhelpers.AssertEqual(t, prev, "")
	testhelpers.AssertEqual(t, isFull, false)

	prev, isFull = rb.Add("Event 5")
	testhelpers.AssertEqual(t, prev, "")
	testhelpers.AssertEqual(t, isFull, false)

	prev, isFull = rb.Add("Event 6")
	testhelpers.AssertEqual(t, prev, "Event 1")
	testhelpers.AssertEqual(t, isFull, true)

	testhelpers.AssertEqual(t, rb.FirstN(10), []string{
		"Event 2",
		"Event 3",
		"Event 4",
		"Event 5",
		"Event 6",
	})
	testhelpers.AssertEqual(t, rb.LastN(3), []string{
		"Event 4",
		"Event 5",
		"Event 6",
	})

	testhelpers.AssertEqual(t, rb.Size(), 5)
	testhelpers.AssertEqual(t, rb.All(), []string{
		"Event 2",
		"Event 3",
		"Event 4",
		"Event 5",
		"Event 6",
	})
	testhelpers.AssertEqual(t, rb.Size(), 5)

	// Should clear the ring buffer
	rb.Clear()
	testhelpers.AssertEqual(t, rb.All(), nil)
	testhelpers.AssertEqual(t, rb.Size(), 0)
}
