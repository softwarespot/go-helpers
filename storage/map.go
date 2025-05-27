package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"
)

type Map[K comparable, V any] struct {
	storage   *Storage
	tableName string
}

// NewMap creates a new map which is persisted to a SQLite database
func NewMap[K comparable, V any](s *Storage, name string) (*Map[K, V], error) {
	tableName := fmt.Sprintf("map_%s", name)
	err := execTransaction(s.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(fmt.Sprintf(
			`
				CREATE TABLE IF NOT EXISTS %s (
					key_hash TEXT PRIMARY KEY,
					key BLOB NOT NULL,
					value BLOB NOT NULL,
					expires_at INTEGER DEFAULT 0
				)
			`,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewMap: create map table: %w", err)
		}

		// Add an expires_at index for efficently cleaning up
		_, err = tx.Exec(fmt.Sprintf(
			`
				CREATE INDEX IF NOT EXISTS %s_expires_idx ON %s(expires_at)
			`,
			tableName,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewMap: create map expires at index: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &Map[K, V]{
		storage:   s,
		tableName: tableName,
	}, nil
}

// Set adds or updates a key/value pair in the map
func (m *Map[K, V]) Set(key K, value V) error {
	return m.set("Set", key, value, 0)
}

// SetEx adds or updates a key/value pair in the map with an expiration duration
func (m *Map[K, V]) SetEx(key K, value V, expiration time.Duration) error {
	return m.set("SetEx", key, value, expiration)
}

func (m *Map[K, V]) set(funcName string, key K, value V, expiration time.Duration) error {
	encKey, err := encode(key)
	if err != nil {
		return fmt.Errorf("map.%s: encode key: %w", funcName, err)
	}
	hashedKey := getHashedKey[K](encKey)

	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("map.%s: encode value: %w", funcName, err)
	}

	var expiresAt int64
	if expiration != 0 {
		expiresAt = now().Add(expiration).Unix()
	}
	query := fmt.Sprintf(
		`
			INSERT OR REPLACE INTO %s (key_hash, key, value, expires_at)
			VALUES (?, ?, ?, ?)
		`,
		m.tableName,
	)
	_, err = m.storage.db.Exec(query, hashedKey, encKey, encValue, expiresAt)
	if err != nil {
		return fmt.Errorf("map.%s: set key/value: %w", funcName, err)
	}
	return nil
}

// Get returns the value for the key in the map
func (m *Map[K, V]) Get(key K) (V, bool, error) {
	var value V

	encKey, err := encode(key)
	if err != nil {
		return value, false, fmt.Errorf("map.Get: encode key: %w", err)
	}
	hashedKey := getHashedKey[K](encKey)

	query := fmt.Sprintf(
		`
		SELECT value, expires_at FROM %s
		WHERE key_hash = ?
		`,
		m.tableName,
	)
	var encValue []byte
	var expiresAt int64
	if err := m.storage.db.QueryRow(query, hashedKey).Scan(&encValue, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return value, false, nil
		}
		return value, false, fmt.Errorf("map.Get: get value: %w", err)
	}
	if hasKeyExpired(expiresAt) {
		return value, false, nil
	}

	value, err = decode[V](encValue)
	if err != nil {
		return value, false, fmt.Errorf("map.Get: decode value: %w", err)
	}
	return value, true, nil
}

// Has returns true if the key exists in the map; otherwise, false when it doesn't exist
func (m *Map[K, V]) Has(key K) (bool, error) {
	encKey, err := encode(key)
	if err != nil {
		return false, fmt.Errorf("map.Has: encode key: %w", err)
	}
	hashedKey := getHashedKey[K](encKey)

	var expiresAt int64
	query := fmt.Sprintf(
		`
			SELECT expires_at FROM %s
			WHERE key_hash = ?
			LIMIT 1
		`,
		m.tableName,
	)
	if err := m.storage.db.QueryRow(query, hashedKey).Scan(&expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("map.Has: has key: %w", err)
	}
	if hasKeyExpired(expiresAt) {
		return false, nil
	}
	return true, nil
}

// Delete returns true if the key was deleted from the map; otherwise, false if it wasn't deleted i.e. not existing
func (m *Map[K, V]) Delete(key K) (bool, error) {
	encKey, err := encode(key)
	if err != nil {
		return false, fmt.Errorf("map.Delete: encode key: %w", err)
	}
	hashedKey := getHashedKey[K](encKey)

	query := fmt.Sprintf(
		`
			DELETE FROM %s
			WHERE key_hash = ?
		`,
		m.tableName,
	)
	res, err := m.storage.db.Exec(query, hashedKey)
	if err != nil {
		return false, fmt.Errorf("map.Delete: delete key: %w", err)
	}

	affectedCount, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("map.Delete: affected count: %w", err)
	}
	return affectedCount > 0, nil
}

// Entries returns an iterator that iterates over all key/value pair entries in the map
func (m *Map[K, V]) Entries() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		query := fmt.Sprintf(
			`
				SELECT key, value FROM %s
				WHERE expires_at = 0 OR expires_at > ?
			`,
			m.tableName,
		)
		rows, err := m.storage.db.Query(query, nowUnix())
		if err != nil {
			// Ignore the error
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encKey, encValue []byte
			if err := rows.Scan(&encKey, &encValue); err != nil {
				// Ignore the error
				continue
			}

			key, err := decode[K](encKey)
			if err != nil {
				// Ignore the error
				continue
			}

			value, err := decode[V](encValue)
			if err != nil {
				// Ignore the error
				continue
			}
			if !yield(key, value) {
				return
			}
		}
	}
}

// Keys returns an iterator that iterates over all keys in the map
func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for key := range m.Entries() {
			if !yield(key) {
				return
			}
		}
	}
}

// Values returns an iterator that iterates over all values in the map
func (m *Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, value := range m.Entries() {
			if !yield(value) {
				return
			}
		}
	}
}

// Size returns the number of key/value pairs in the map
func (m *Map[K, V]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
			SELECT COUNT(1) FROM %s
			WHERE expires_at = 0 OR expires_at > ?
        `,
		m.tableName,
	)
	err := m.storage.db.QueryRow(query, nowUnix()).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("map.Size: get size: %w", err)
	}
	return size, nil
}

// Clear deletes all key/value pairs from the map
func (m *Map[K, V]) Clear() error {
	query := fmt.Sprintf(
		`
			DELETE FROM %s
		`,
		m.tableName,
	)
	if _, err := m.storage.db.Exec(query); err != nil {
		return fmt.Errorf("map.Clear: clear key/values: %w", err)
	}
	return nil
}
