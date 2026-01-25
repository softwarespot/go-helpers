package storage

import (
	"sync"
	"time"
)

// LRUCache represents a thread-safe "Least Recently Used (LRU)" cache
type LRUCache[K comparable, V any] struct {
	nodes      map[K]*cacheNode[K, V]
	head       *cacheNode[K, V]
	tail       *cacheNode[K, V]
	expiration time.Duration
	maxSize    int
	size       int

	cleanupDone chan struct{}
	cleanupWg   sync.WaitGroup

	mu sync.Mutex
}

type cacheNode[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
	prev      *cacheNode[K, V]
	next      *cacheNode[K, V]
}

func (n *cacheNode[K, V]) hasExpired(now time.Time) bool {
	if n.expiresAt.IsZero() {
		return false
	}
	return now.After(n.expiresAt)
}

// NewLRUCache creates a new LRU cache with the provided maximum size and optional expiration i.e. if 0, then no expiration
func NewLRUCache[K comparable, V any](maxSize int, expiration time.Duration) *LRUCache[K, V] {
	if maxSize <= 0 {
		panic("lru_cache.NewLRUCache: maxSize must be greater than 0")
	}
	return &LRUCache[K, V]{
		nodes:      map[K]*cacheNode[K, V]{},
		expiration: expiration,
		maxSize:    maxSize,
	}
}

// StartCleanup starts a goroutine that periodically cleans up expired nodes in the cache
func (c *LRUCache[K, V]) StartCleanup(interval time.Duration) {
	c.mu.Lock()
	if c.cleanupDone != nil {
		c.mu.Unlock()
		return
	}

	c.cleanupWg.Add(1)
	c.cleanupDone = make(chan struct{})
	c.mu.Unlock()

	go func() {
		defer c.cleanupWg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-c.cleanupDone:
				return
			case <-ticker.C:
				c.cleanupExpiredNodes()
			}
		}
	}()
}

// StopCleanup stops the periodic cleanup of expired nodes in the cache
func (c *LRUCache[K, V]) StopCleanup() {
	c.mu.Lock()
	if c.cleanupDone == nil {
		c.mu.Unlock()
		return
	}

	cleanupDone := c.cleanupDone
	c.cleanupDone = nil
	c.mu.Unlock()

	close(cleanupDone)
	c.cleanupWg.Wait()
}

// Set adds or updates a key/value pair in the cache
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.expiration)
}

// SetWithTTL adds or updates a key/value pair in the cache with an expiration duration
func (c *LRUCache[K, V]) SetWithTTL(key K, value V, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if expiration > 0 {
		expiresAt = time.Now().Add(expiration)
	}

	if node, ok := c.nodes[key]; ok {
		node.value = value
		node.expiresAt = expiresAt
		c.moveNodeToFront(node)
		return
	}

	node := &cacheNode[K, V]{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}

	c.nodes[key] = node
	c.addNodeToFront(node)
	c.size++

	if c.size > c.maxSize {
		c.deleteNode(c.tail)
	}
}

// Get returns the value for the key in the cache.
// If the key does not exist, it returns false
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var value V
	node, ok := c.nodes[key]
	if !ok {
		return value, false
	}

	if node.hasExpired(time.Now()) {
		c.deleteNode(node)
		return value, false
	}

	c.moveNodeToFront(node)
	return node.value, true
}

// Has returns true if the key exists in the cache; otherwise, false
func (c *LRUCache[K, V]) Has(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.nodes[key]
	if !ok {
		return false
	}

	if node.hasExpired(time.Now()) {
		c.deleteNode(node)
		return false
	}
	return true
}

// Peek returns the value for the key in the cache without updating its position
// in the LRU list. If the key does not exist, it returns false
func (c *LRUCache[K, V]) Peek(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var value V
	node, ok := c.nodes[key]
	if !ok {
		return value, false
	}

	if node.hasExpired(time.Now()) {
		c.deleteNode(node)
		return value, false
	}
	return node.value, true
}

// Delete deletes a key/value pair from the cache
func (c *LRUCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, ok := c.nodes[key]; ok {
		c.deleteNode(node)
	}
}

// Size returns the number of values in the cache
func (c *LRUCache[K, V]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.size
}

// Clear deletes all values from the cache
func (c *LRUCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	clear(c.nodes)
	c.head = nil
	c.tail = nil
	c.size = 0
}

func (c *LRUCache[K, V]) addNodeToFront(node *cacheNode[K, V]) {
	node.next = c.head
	node.prev = nil

	if c.head != nil {
		c.head.prev = node
	}
	c.head = node

	if c.tail == nil {
		c.tail = node
	}
}

func (c *LRUCache[K, V]) moveNodeToFront(node *cacheNode[K, V]) {
	if node == c.head {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}

	if node.next == nil {
		c.tail = node.prev
	} else {
		node.next.prev = node.prev
	}

	node.next = c.head
	node.prev = nil
	c.head.prev = node
	c.head = node
}

func (c *LRUCache[K, V]) deleteNode(node *cacheNode[K, V]) {
	if node == nil {
		return
	}

	if node.prev == nil {
		c.head = node.next
	} else {
		node.prev.next = node.next
	}

	if node.next == nil {
		c.tail = node.prev
	} else {
		node.next.prev = node.prev
	}

	delete(c.nodes, node.key)
	c.size--
}

func (c *LRUCache[K, V]) cleanupExpiredNodes() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	totalExpired := 0
	now := time.Now()

	// Start from the tail and work backwards to delete expired nodes
	node := c.tail
	for node != nil {
		prevNode := node.prev
		if node.hasExpired(now) {
			c.deleteNode(node)
			totalExpired++
		}
		node = prevNode
	}
	return totalExpired
}
