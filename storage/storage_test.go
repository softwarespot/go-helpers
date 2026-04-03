package storage

import (
	"crypto/rand"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestStorageOperations(t *testing.T) {
	store, err := New("test_demo.sqlite")
	if err != nil {
		t.Fatalf("storage.New() error = %v", err)
	}
	defer store.Close()

	t.Run("Map", func(t *testing.T) {
		userPrefs, err := NewMap[string, string](store, "user_preferences")
		if err != nil {
			t.Fatalf("NewMap[string, string]() error = %v", err)
		}

		if err := userPrefs.Clear(); err != nil {
			t.Fatalf("userPrefs.Clear() error = %v", err)
		}

		initialSize, err := userPrefs.Size()
		if err != nil {
			t.Fatalf("userPrefs.Size() error = %v", err)
		}
		if initialSize != 0 {
			t.Errorf("userPrefs.Size() got = %d, want = 0 after clear", initialSize)
		}

		testMapSet(t, userPrefs, "name1", "value1")
		testMapSet(t, userPrefs, "name2", "value2")
		testMapSet(t, userPrefs, "name3", "value3")

		sizeAfterSets, err := userPrefs.Size()
		if err != nil {
			t.Fatalf("userPrefs.Size() error = %v", err)
		}
		if sizeAfterSets != 3 {
			t.Errorf("userPrefs.Size() got = %d, want = 3 after sets", sizeAfterSets)
		}

		testMapHas(t, userPrefs, "name2", true)
		testMapHas(t, userPrefs, "name4", false)

		testMapGet(t, userPrefs, "name1", "value1", true)

		testMapDelete(t, userPrefs, "name2")

		// Deleting non-existent key should not error
		testMapDelete(t, userPrefs, "name2")

		sizeAfterDelete, err := userPrefs.Size()
		if err != nil {
			t.Fatalf("userPrefs.Size() error = %v", err)
		}
		if sizeAfterDelete != 2 {
			t.Errorf("userPrefs.Size() got = %d, want = 2 after delete", sizeAfterDelete)
		}

		t.Log("\nUser preferences (Entries):")
		entries := make(map[string]string)
		for k, v := range userPrefs.Entries() {
			t.Logf("- %s: %s\n", k, v)
			entries[k] = v
		}
		if err := userPrefs.IterError(); err != nil {
			t.Errorf("userPrefs.IterError() after Entries: %v", err)
		}

		if expectedEntries := (map[string]string{"name1": "value1", "name3": "value3"}); !reflect.DeepEqual(entries, expectedEntries) {
			t.Errorf("userPrefs.Entries() got = %v, want = %v", entries, expectedEntries)
		}

		t.Log("User preference keys:")
		var keys []string
		for k := range userPrefs.Keys() {
			t.Logf("- %s\n", k)
			keys = append(keys, k)
		}
		slices.Sort(keys)

		if err := userPrefs.IterError(); err != nil {
			t.Errorf("userPrefs.IterError() after Keys: %v", err)
		}

		if expectedKeys := ([]string{"name1", "name3"}); !reflect.DeepEqual(keys, expectedKeys) {
			t.Errorf("userPrefs.Keys() got = %v, want = %v", keys, expectedKeys)
		}

		t.Log("User preference values:")
		var values []string
		for v := range userPrefs.Values() {
			t.Logf("- %s\n", v)
			values = append(values, v)
		}
		slices.Sort(values)

		if err := userPrefs.IterError(); err != nil {
			t.Errorf("userPrefs.IterError() after Values: %v", err)
		}

		if expectedValues := ([]string{"value1", "value3"}); !reflect.DeepEqual(values, expectedValues) {
			t.Errorf("userPrefs.Values() got = %v, want = %v", values, expectedValues)
		}
	})

	t.Run("Set", func(t *testing.T) {
		tags, err := NewSet[string](store, "post_tags")
		if err != nil {
			t.Fatalf("NewSet[string]() error = %v", err)
		}

		if err := tags.Clear(); err != nil {
			t.Fatalf("tags.Clear() error = %v", err)
		}

		initialSize, err := tags.Size()
		if err != nil {
			t.Fatalf("tags.Size() error = %v", err)
		}
		if initialSize != 0 {
			t.Errorf("tags.Size() got = %d, want = 0 after clear", initialSize)
		}

		testSetAdd(t, tags, "tag1")
		testSetAdd(t, tags, "tag2")
		testSetAdd(t, tags, "tag3")

		// Adding duplicate should not error and not change size
		testSetAdd(t, tags, "tag1")

		randomTagValue := rand.Text()
		if err := tags.AddEx(randomTagValue, 10*time.Millisecond); err != nil {
			t.Errorf("tags.AddEx(%q) error = %v", randomTagValue, err)
		}

		sizeAfterAdds, err := tags.Size()
		if err != nil {
			t.Fatalf("tags.Size() error = %v", err)
		}
		if sizeAfterAdds != 4 {
			t.Errorf("tags.Size() got = %d, want = 4 after adds", sizeAfterAdds)
		}

		testSetHas(t, tags, "tag2", true)
		testSetHas(t, tags, "tag5", false)

		testSetDelete(t, tags, "tag2")

		// Deleting non-existent should not error
		testSetDelete(t, tags, "tag2")

		sizeAfterDelete, err := tags.Size()
		if err != nil {
			t.Fatalf("tags.Size() error = %v", err)
		}
		if sizeAfterDelete != 3 {
			t.Errorf("tags.Size() got = %d, want = 3 after delete", sizeAfterDelete)
		}

		t.Log("\nTags (Values):")
		var setValues []string
		for tag := range tags.Values() {
			t.Logf("- %s\n", tag)
			setValues = append(setValues, tag)
		}
		slices.Sort(setValues)
		if err := tags.IterError(); err != nil {
			t.Errorf("tags.IterError() after Values: %v", err)
		}
		// Check for tag1, tag3, and randomTagValue
		if !(slices.Contains(setValues, "tag1") && slices.Contains(setValues, "tag3") && slices.Contains(setValues, randomTagValue)) {
			t.Errorf("tags.Values() missing expected tags, got %v", setValues)
		}

		// Wait for expiry and check
		expiry := 10 * time.Millisecond
		time.Sleep(expiry * 2)
		sizeAfterExpiry, err := tags.Size()
		if err != nil {
			t.Fatalf("tags.Size() after expiry error = %v", err)
		}
		if sizeAfterExpiry != 2 {
			t.Errorf("tags.Size() after expiry got = %d, want = 2 (expected %q to expire)",
				sizeAfterExpiry, randomTagValue)
		}

		if hasExpired, err := tags.Has(randomTagValue); err != nil || hasExpired {
			t.Errorf("tags.Has(%q) after expiry got %t, err=%v; want false, nil",
				randomTagValue, hasExpired, err)
		}
	})

	t.Run("Queue", func(t *testing.T) {
		tasks, err := NewQueue[string](store, "background_tasks")
		if err != nil {
			t.Fatalf("NewQueue[string]() error = %v", err)
		}

		testQueueClear(t, tasks)

		testQueueSize(t, tasks, 0, "after clear")

		testQueueEnqueue(t, tasks, "process-email-1")
		testQueueEnqueue(t, tasks, "process-image-2")
		testQueueEnqueue(t, tasks, "send-notification-3")

		tempTask := "temp-task-4"
		tempExpiry := 50 * time.Millisecond
		if err := tasks.EnqueueEx(tempTask, tempExpiry); err != nil {
			t.Errorf("tasks.EnqueueEx(%q, %v) error = %v", tempTask, tempExpiry, err)
		}

		testQueueSize(t, tasks, 4, "after enqueues")

		testQueuePeek(t, tasks, "process-email-1", true)

		// Dequeue multiple items and verify
		expectedDequeues := []string{"process-email-1", "process-image-2"}
		for _, expectedVal := range expectedDequeues {
			testQueueDequeue(t, tasks, expectedVal, true)
		}

		testQueueSize(t, tasks, 2, "after dequeues")

		t.Log("\nRemaining tasks (Entries):")
		var remainingTasks []string
		for task := range tasks.Entries() {
			t.Logf("- %s\n", task)
			remainingTasks = append(remainingTasks, task)
		}
		if err := tasks.IterError(); err != nil {
			t.Errorf("tasks.IterError() after Entries: %v", err)
		}

		if expectedRemaining := ([]string{"send-notification-3", tempTask}); !reflect.DeepEqual(remainingTasks, expectedRemaining) {
			t.Errorf("tasks.Entries() got %v, want %v", remainingTasks, expectedRemaining)
		}

		// Wait for temp task to expire
		t.Log("\nWaiting for expiration...")
		time.Sleep(tempExpiry * 2)

		testQueueSize(t, tasks, 1, "after expiration")

		t.Log("\nRemaining tasks after expiration (Entries):")
		var finalTasks []string
		for task := range tasks.Entries() {
			t.Logf("- %s\n", task)
			finalTasks = append(finalTasks, task)
		}
		if err := tasks.IterError(); err != nil {
			t.Errorf("tasks.IterError() after Entries (post-expiration): %v", err)
		}

		if expectedFinal := ([]string{"send-notification-3"}); !reflect.DeepEqual(finalTasks, expectedFinal) {
			t.Errorf("tasks.Entries() (post-expiration) got %v, want %v", finalTasks, expectedFinal)
		}

		// Dequeue the last item
		testQueueDequeue(t, tasks, "send-notification-3", true)

		// Queue should be empty
		testQueueSize(t, tasks, 0, "at end")

		testQueuePeek(t, tasks, "", false)
	})

	t.Run("Stack", func(t *testing.T) {
		items, err := NewStack[string](store, "history_items")
		if err != nil {
			t.Fatalf("NewStack[string]() error = %v", err)
		}

		testStackClear(t, items)

		testStackSize(t, items, 0, "after clear")

		testStackPush(t, items, "item1")
		testStackPush(t, items, "item2")
		testStackPush(t, items, "item3")

		tempItem := "temp-item"
		tempExpiry := 50 * time.Millisecond
		if err := items.PushEx(tempItem, tempExpiry); err != nil {
			t.Errorf("items.PushEx(%q, %v) error = %v", tempItem, tempExpiry, err)
		}

		testStackSize(t, items, 4, "after pushes")

		// Stack should return items in LIFO order (Last In, First Out)
		testStackPeek(t, items, tempItem, true)

		// Pop items and verify LIFO order
		expectedPops := []string{tempItem, "item3", "item2"}
		for _, expectedVal := range expectedPops {
			testStackPop(t, items, expectedVal, true)
		}

		testStackSize(t, items, 1, "after pops")

		t.Log("\nRemaining items (Entries):")
		var remainingItems []string
		for item := range items.Entries() {
			t.Logf("- %s\n", item)
			remainingItems = append(remainingItems, item)
		}
		if err := items.IterError(); err != nil {
			t.Errorf("items.IterError() after Entries: %v", err)
		}

		if expectedRemaining := ([]string{"item1"}); !reflect.DeepEqual(remainingItems, expectedRemaining) {
			t.Errorf("items.Entries() got %v, want %v", remainingItems, expectedRemaining)
		}

		// Test temporary item with expiry
		tempStack, err := NewStack[string](store, "temp_stack")
		if err != nil {
			t.Fatalf("NewStack[string]() for temp test error = %v", err)
		}

		if err := tempStack.Clear(); err != nil {
			t.Fatalf("tempStack.Clear() error = %v", err)
		}

		permanentItem := "permanent"
		if err := tempStack.Push(permanentItem); err != nil {
			t.Errorf("tempStack.Push(%q) error = %v", permanentItem, err)
		}

		expiringItem := "expiring"
		shortExpiry := 20 * time.Millisecond
		if err := tempStack.PushEx(expiringItem, shortExpiry); err != nil {
			t.Errorf("tempStack.PushEx(%q, %v) error = %v", expiringItem, shortExpiry, err)
		}

		// Both items should be present initially
		testStackSize(t, tempStack, 2, "after pushing permanent and expiring items")

		// Wait for expiring item to expire
		time.Sleep(shortExpiry * 2)

		// After expiry, should only have the permanent item
		testStackSize(t, tempStack, 1, "after expiry")

		// Peek should now show the permanent item
		testStackPeek(t, tempStack, permanentItem, true)

		// Pop the last item
		testStackPop(t, tempStack, permanentItem, true)

		// Stack should be empty
		testStackSize(t, tempStack, 0, "at end")
		testStackPeek(t, tempStack, "", false)
	})
}

// Test helpers for Map operations
func testMapSet[K comparable, V any](t *testing.T, m *Map[K, V], key K, value V) {
	t.Helper()
	if err := m.Set(key, value); err != nil {
		t.Errorf("Map.Set(%v, %v) error = %v", key, value, err)
	}
}

func testMapGet[K, V comparable](t *testing.T, m *Map[K, V], key K, wantValue V, wantFound bool) {
	t.Helper()
	gotValue, gotFound, err := m.Get(key)
	if err != nil {
		t.Errorf("Map.Get(%v) error = %v", key, err)
		return
	}
	if gotFound != wantFound || (wantFound && gotValue != wantValue) {
		t.Errorf("Map.Get(%v) got value=%v, found=%t; want value=%v, found=%t",
			key, gotValue, gotFound, wantValue, wantFound)
	}
}

func testMapHas[K comparable, V any](t *testing.T, m *Map[K, V], key K, want bool) {
	t.Helper()
	got, err := m.Has(key)
	if err != nil {
		t.Errorf("Map.Has(%v) error = %v", key, err)
		return
	}
	if got != want {
		t.Errorf("Map.Has(%v) got = %t, want = %t", key, got, want)
	}
}

func testMapDelete[K comparable, V any](t *testing.T, m *Map[K, V], key K) {
	t.Helper()
	if err := m.Delete(key); err != nil {
		t.Errorf("Map.Delete(%v) error = %v", key, err)
	}
}

// Test helpers for Set operations
func testSetAdd[V comparable](t *testing.T, s *Set[V], value V) {
	t.Helper()
	if err := s.Add(value); err != nil {
		t.Errorf("Set.Add(%v) error = %v", value, err)
	}
}

func testSetHas[V comparable](t *testing.T, s *Set[V], value V, want bool) {
	t.Helper()
	got, err := s.Has(value)
	if err != nil {
		t.Errorf("Set.Has(%v) error = %v", value, err)
		return
	}
	if got != want {
		t.Errorf("Set.Has(%v) got = %t, want = %t", value, got, want)
	}
}

func testSetDelete[V comparable](t *testing.T, s *Set[V], value V) {
	t.Helper()
	if err := s.Delete(value); err != nil {
		t.Errorf("Set.Delete(%v) error = %v", value, err)
	}
}

// Test helpers for Queue operations
func testQueueClear[T any](t *testing.T, q *Queue[T]) {
	t.Helper()
	if err := q.Clear(); err != nil {
		t.Fatalf("Queue.Clear() error = %v", err)
	}
	t.Log("Queue.Clear: success")
}

func testQueueSize[T any](t *testing.T, q *Queue[T], want int, context string) {
	t.Helper()
	got, err := q.Size()
	if err != nil {
		t.Fatalf("Queue.Size() %s error = %v", context, err)
	}
	if got != want {
		t.Errorf("Queue.Size() %s got = %d, want = %d", context, got, want)
	}
}

func testQueueEnqueue[T any](t *testing.T, q *Queue[T], value T) {
	t.Helper()
	if err := q.Enqueue(value); err != nil {
		t.Errorf("Queue.Enqueue(%v) error = %v", value, err)
	}
}

func testQueueDequeue[T comparable](t *testing.T, q *Queue[T], wantValue T, wantFound bool) {
	t.Helper()
	gotValue, gotFound, err := q.Dequeue()
	if err != nil {
		t.Errorf("Queue.Dequeue() error = %v", err)
		return
	}
	if gotFound != wantFound || (wantFound && gotValue != wantValue) {
		t.Errorf("Queue.Dequeue() got value=%v, found=%t; want value=%v, found=%t",
			gotValue, gotFound, wantValue, wantFound)
	}
}

func testQueuePeek[T comparable](t *testing.T, q *Queue[T], wantValue T, wantFound bool) {
	t.Helper()
	gotValue, gotFound, err := q.Peek()
	if err != nil {
		t.Errorf("Queue.Peek() error = %v", err)
		return
	}
	if gotFound != wantFound || (wantFound && gotValue != wantValue) {
		t.Errorf("Queue.Peek() got value=%v, found=%t; want value=%v, found=%t",
			gotValue, gotFound, wantValue, wantFound)
	}
}

func testStackClear[T any](t *testing.T, s *Stack[T]) {
	t.Helper()
	if err := s.Clear(); err != nil {
		t.Fatalf("Stack.Clear() error = %v", err)
	}
	t.Log("Stack.Clear: success")
}

func testStackSize[T any](t *testing.T, s *Stack[T], want int, context string) {
	t.Helper()
	got, err := s.Size()
	if err != nil {
		t.Fatalf("Stack.Size() %s error = %v", context, err)
	}
	if got != want {
		t.Errorf("Stack.Size() %s got = %d, want = %d", context, got, want)
	}
}

func testStackPush[T any](t *testing.T, s *Stack[T], value T) {
	t.Helper()
	if err := s.Push(value); err != nil {
		t.Errorf("Stack.Push(%v) error = %v", value, err)
	}
}

func testStackPop[T comparable](t *testing.T, s *Stack[T], wantValue T, wantFound bool) {
	t.Helper()
	gotValue, gotFound, err := s.Pop()
	if err != nil {
		t.Errorf("Stack.Pop() error = %v", err)
		return
	}
	if gotFound != wantFound || (wantFound && gotValue != wantValue) {
		t.Errorf("Stack.Pop() got value=%v, found=%t; want value=%v, found=%t",
			gotValue, gotFound, wantValue, wantFound)
	}
}

func testStackPeek[T comparable](t *testing.T, s *Stack[T], wantValue T, wantFound bool) {
	t.Helper()
	gotValue, gotFound, err := s.Peek()
	if err != nil {
		t.Errorf("Stack.Peek() error = %v", err)
		return
	}
	if gotFound != wantFound || (wantFound && gotValue != wantValue) {
		t.Errorf("Stack.Peek() got value=%v, found=%t; want value=%v, found=%t",
			gotValue, gotFound, wantValue, wantFound)
	}
}
