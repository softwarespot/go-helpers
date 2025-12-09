package storage

import (
	"fmt"
	"time"
)

type Cache[K comparable, V any] struct {
	m          *Map[K, V]
	expiration time.Duration
}

// NewCache creates a new cache which is persisted to a SQLite database.
// NOTE: If the expiration duration is 0, the cache behaves like a regular map without expiration
func NewCache[K comparable, V any](s *Storage, name string, expiration time.Duration) (*Cache[K, V], error) {
	tableName := getNormalizedTableName("cache", name)
	m, err := NewMap[K, V](s, tableName)
	if err != nil {
		return nil, err
	}
	return &Cache[K, V]{
		m:          m,
		expiration: expiration,
	}, nil
}

// Set adds or updates a key/value pair in the cache
func (c *Cache[K, V]) Set(key K, value V) error {
	if c.expiration == 0 {
		if err := c.m.Set(key, value); err != nil {
			return fmt.Errorf("cache.Set: %w", err)
		}
		return nil
	}
	if err := c.m.SetEx(key, value, c.expiration); err != nil {
		return fmt.Errorf("cache.Set: %w", err)
	}
	return nil
}

// MSet adds or updates multiple key/value pairs in the cache
func (c *Cache[K, V]) MSet(pairs map[K]V) error {
	if c.expiration == 0 {
		if err := c.m.MSet(pairs); err != nil {
			return fmt.Errorf("cache.MSet: %w", err)
		}
		return nil
	}
	if err := c.m.MSetEx(pairs, c.expiration); err != nil {
		return fmt.Errorf("cache.MSet: %w", err)
	}
	return nil
}

// Get returns the value for the key in the cache
func (c *Cache[K, V]) Get(key K) (V, bool, error) {
	value, ok, err := c.m.Get(key)
	if err != nil {
		return value, ok, fmt.Errorf("cache.Get: %w", err)
	}
	return value, ok, nil
}

// MGet returns a map of values for the specified keys in the cache.
// NOTE: If a key does not exist, it will not be included in the returned map
func (c *Cache[K, V]) MGet(keys ...K) (map[K]V, error) {
	values, err := c.m.MGet(keys...)
	if err != nil {
		return nil, fmt.Errorf("cache.MGet: %w", err)
	}
	return values, nil
}

// GetEx returns the value for the key in the cache and sets its expiration duration (when greater than 0)
func (c *Cache[K, V]) GetEx(key K) (V, bool, error) {
	value, ok, err := c.m.Get(key)
	if err != nil {
		return value, ok, fmt.Errorf("cache.GetEx: %w", err)
	}
	if c.expiration == 0 || !ok {
		return value, ok, nil
	}
	if err := c.m.SetEx(key, value, c.expiration); err != nil {
		return value, true, fmt.Errorf("cache.GetEx: refresh expiry: %w", err)
	}
	return value, true, nil
}

// MGetEx returns a map of values for the specified keys and refreshes their expiration time
func (c *Cache[K, V]) MGetEx(keys ...K) (map[K]V, error) {
	values, err := c.m.MGet(keys...)
	if err != nil {
		return nil, fmt.Errorf("cache.MGetEx: %w", err)
	}

	if c.expiration > 0 && len(values) > 0 {
		if err := c.m.MSetEx(values, c.expiration); err != nil {
			return values, fmt.Errorf("cache.MGetEx: refresh expiry: %w", err)
		}
	}
	return values, nil
}

// Has returns true if the key exists in the cache; otherwise, false
func (c *Cache[K, V]) Has(key K) (bool, error) {
	ok, err := c.m.Has(key)
	if err != nil {
		return false, fmt.Errorf("cache.Has: %w", err)
	}
	return ok, nil
}

// Delete deletes a key/value pair from the cache
func (c *Cache[K, V]) Delete(key K) error {
	if err := c.m.Delete(key); err != nil {
		return fmt.Errorf("cache.Delete: %w", err)
	}
	return nil
}

// Size returns the number of values in the cache
func (c *Cache[K, V]) Size() (int, error) {
	size, err := c.m.Size()
	if err != nil {
		return 0, fmt.Errorf("cache.Size: %w", err)
	}
	return size, nil
}

// Clear deletes all values from the cache
func (c *Cache[K, V]) Clear() error {
	if err := c.m.Clear(); err != nil {
		return fmt.Errorf("cache.Clear: %w", err)
	}
	return nil
}
