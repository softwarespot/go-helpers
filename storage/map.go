package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"strings"
	"time"
)

type Map[K comparable, V any] struct {
	storage       *Storage
	tableName     string
	lastIterError error
}

// NewMap creates a new map which is persisted to a SQLite database
func NewMap[K comparable, V any](s *Storage, name string) (*Map[K, V], error) {
	tableName := getNormalizedTableName("map", name)
	if err := execTransaction(s.db, func(tx *sql.Tx) error {
		if _, err := tx.Exec(fmt.Sprintf(
			`
				CREATE TABLE IF NOT EXISTS %s (
					key_hash TEXT PRIMARY KEY,
					key BLOB NOT NULL,
					value BLOB NOT NULL,
					expires_at INTEGER DEFAULT 0,
					updated_at INTEGER NOT NULL
				)
			`,
			tableName,
		)); err != nil {
			return fmt.Errorf("storage.NewMap: create map table: %w", err)
		}

		if _, err := tx.Exec(fmt.Sprintf(
			`
				CREATE INDEX IF NOT EXISTS %s_expires_idx ON %s(expires_at)
			`,
			tableName,
			tableName,
		)); err != nil {
			return fmt.Errorf("storage.NewMap: create map expires at index: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &Map[K, V]{
		storage:       s,
		tableName:     tableName,
		lastIterError: nil,
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

	query := fmt.Sprintf(
		`
			INSERT INTO %s (key_hash, key, value, expires_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(key_hash) DO UPDATE SET
				value = excluded.value,
				expires_at = excluded.expires_at,
				updated_at = excluded.updated_at
		`,
		m.tableName,
	)
	if _, err = m.storage.db.Exec(query, hashedKey, encKey, encValue, getKeyExpirationAsMilli(expiration), nowUnixMilli()); err != nil {
		return fmt.Errorf("map.%s: set key/value: %w", funcName, err)
	}
	return nil
}

// MSet adds or updates multiple key/value pairs in the map
func (m *Map[K, V]) MSet(pairs map[K]V) error {
	return m.mset("MSet", pairs, 0)
}

// MSetEx adds or updates multiple key/value pairs with an expiration duration
func (m *Map[K, V]) MSetEx(pairs map[K]V, expiration time.Duration) error {
	return m.mset("MSetEx", pairs, expiration)
}

// SQLite default limit is 999 parameters, each row uses 5 parameters
const defaultSetChunkSize = 199

func (m *Map[K, V]) mset(funcName string, pairs map[K]V, expiration time.Duration) error {
	if len(pairs) == 0 {
		return nil
	}

	return execTransaction(m.storage.db, func(tx *sql.Tx) error {
		currCount := 0
		expiresAt := getKeyExpirationAsMilli(expiration)
		now := nowUnixMilli()

		var placeholders []string
		var args []any
		for k, v := range pairs {
			encKey, err := encode(k)
			if err != nil {
				return fmt.Errorf("map.%s: encode key: %w", funcName, err)
			}
			hashedKey := getHashedKey[K](encKey)

			encValue, err := encode(v)
			if err != nil {
				return fmt.Errorf("map.%s: encode value: %w", funcName, err)
			}

			placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
			args = append(args, hashedKey, encKey, encValue, expiresAt, now)
			currCount++

			if currCount == defaultSetChunkSize {
				if err := execSetBatch(tx, m.tableName, funcName, placeholders, args); err != nil {
					return err
				}

				placeholders = nil
				args = nil
				currCount = 0
			}
		}
		if currCount > 0 {
			if err := execSetBatch(tx, m.tableName, funcName, placeholders, args); err != nil {
				return err
			}
		}
		return nil
	})
}

func execSetBatch(tx *sql.Tx, tableName, funcName string, placeholders []string, args []any) error {
	query := fmt.Sprintf(
		`
			INSERT INTO %s (key_hash, key, value, expires_at, updated_at)
         	VALUES %s
         	ON CONFLICT(key_hash) DO UPDATE SET
             	value = excluded.value,
             	expires_at = excluded.expires_at,
             	updated_at = excluded.updated_at
		`,
		tableName,
		strings.Join(placeholders, ","),
	)
	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("map.%s: batch insert: %w", funcName, err)
	}
	return nil
}

// Get returns the value for the key in the map.
// If the key does not exist, it returns false and no error
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
			LIMIT 1
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

// MGet returns a map of values for the specified keys in the map.
// NOTE: If a key does not exist, it will not be returned in the map
func (m *Map[K, V]) MGet(keys ...K) (map[K]V, error) {
	if len(keys) == 0 {
		return map[K]V{}, nil
	}

	var placeholdersBuilder strings.Builder
	var args []any
	for i, key := range keys {
		encKey, err := encode(key)
		if err != nil {
			return nil, fmt.Errorf("map.MGet: encode key %v: %w", key, err)
		}
		hashedKey := getHashedKey[K](encKey)

		if i > 0 {
			placeholdersBuilder.WriteByte(',')
		}
		placeholdersBuilder.WriteByte('?')
		args = append(args, hashedKey)
	}
	args = append(args, nowUnixMilli())

	query := fmt.Sprintf(
		`
			SELECT key, value FROM %s
        	WHERE key_hash IN (%s)
				AND (expires_at = 0 OR expires_at > ?)
		`,
		m.tableName,
		placeholdersBuilder.String(),
	)
	rows, err := m.storage.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("map.MGet: query key/values: %w", err)
	}
	defer rows.Close()

	res := map[K]V{}
	for rows.Next() {
		var encKey, encValue []byte
		if err := rows.Scan(&encKey, &encValue); err != nil {
			return nil, fmt.Errorf("map.MGet: get key/value: %w", err)
		}

		key, err := decode[K](encKey)
		if err != nil {
			return nil, fmt.Errorf("map.MGet: decode key: %w", err)
		}

		value, err := decode[V](encValue)
		if err != nil {
			return nil, fmt.Errorf("map.MGet: decode value: %w", err)
		}
		res[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("map.MGet: iterate key/values: %w", err)
	}
	return res, nil
}

// Has returns true if the key exists in the map; otherwise, false
func (m *Map[K, V]) Has(key K) (bool, error) {
	encKey, err := encode(key)
	if err != nil {
		return false, fmt.Errorf("map.Has: encode key: %w", err)
	}
	hashedKey := getHashedKey[K](encKey)

	query := fmt.Sprintf(
		`
			SELECT EXISTS(
				SELECT 1 FROM %s
				WHERE key_hash = ?
					AND (expires_at = 0 OR expires_at > ?)
			)
		`,
		m.tableName,
	)
	var exists bool
	if err := m.storage.db.QueryRow(query, hashedKey, nowUnixMilli()).Scan(&exists); err != nil {
		return false, fmt.Errorf("map.Has: has key: %w", err)
	}
	return exists, nil
}

// Delete deletes a key/value pair from the map
func (m *Map[K, V]) Delete(key K) error {
	encKey, err := encode(key)
	if err != nil {
		return fmt.Errorf("map.Delete: encode key: %w", err)
	}
	hashedKey := getHashedKey[K](encKey)

	query := fmt.Sprintf(
		`
			DELETE FROM %s
			WHERE key_hash = ?
		`,
		m.tableName,
	)
	if _, err := m.storage.db.Exec(query, hashedKey); err != nil {
		return fmt.Errorf("map.Delete: delete key: %w", err)
	}
	return nil
}

// Entries returns an iterator that iterates over all key/value pair entries in the map
func (m *Map[K, V]) Entries() iter.Seq2[K, V] {
	m.lastIterError = nil
	return func(yield func(K, V) bool) {
		query := fmt.Sprintf(
			`
				SELECT key, value FROM %s
				WHERE expires_at = 0 OR expires_at > ?
				ORDER BY updated_at DESC
			`,
			m.tableName,
		)
		rows, err := m.storage.db.Query(query, nowUnixMilli())
		if err != nil {
			m.lastIterError = fmt.Errorf("map.Entries: query key/values: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encKey, encValue []byte
			if err := rows.Scan(&encKey, &encValue); err != nil {
				m.lastIterError = fmt.Errorf("map.Entries: get key/value: %w", err)
				return
			}

			key, err := decode[K](encKey)
			if err != nil {
				m.lastIterError = fmt.Errorf("map.Entries: decode key: %w", err)
				return
			}

			value, err := decode[V](encValue)
			if err != nil {
				m.lastIterError = fmt.Errorf("map.Entries: decode value: %w", err)
				return
			}
			if !yield(key, value) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			m.lastIterError = fmt.Errorf("map.Entries: iterate key/values: %w", err)
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

// IterError returns the first error encountered during the last iteration.
// NOTE: It should be called after iteration has completed
func (m *Map[K, V]) IterError() error {
	return m.lastIterError
}

// Size returns the number of key/value pairs in the map
func (m *Map[K, V]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
			SELECT COUNT(*) FROM %s
			WHERE expires_at = 0 OR expires_at > ?
        `,
		m.tableName,
	)
	if err := m.storage.db.QueryRow(query, nowUnixMilli()).Scan(&size); err != nil {
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
