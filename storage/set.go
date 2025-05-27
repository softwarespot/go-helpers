package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"time"
)

type Set[T comparable] struct {
	storage   *Storage
	tableName string
}

// NewSet creates a new set which is persisted to a SQLite database
func NewSet[T comparable](s *Storage, name string) (*Set[T], error) {
	tableName := fmt.Sprintf("set_%s", name)
	err := execTransaction(s.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(fmt.Sprintf(
			`
				CREATE TABLE IF NOT EXISTS %s (
					key_hash TEXT PRIMARY KEY,
					value BLOB NOT NULL,
					expires_at INTEGER DEFAULT 0
				)
			`,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewSet: create set table: %w", err)
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
			return fmt.Errorf("storage.NewSet: create set expires at index: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &Set[T]{
		storage:   s,
		tableName: tableName,
	}, nil
}

// Add returns true if the value is added to the set; otherwise, false when it already exists in the set
func (s *Set[T]) Add(value T) (bool, error) {
	return s.add("Add", value, 0)
}

// AddEx returns true if the value is added to the set with an expiration duration; otherwise, false when it already exists in the set
func (s *Set[T]) AddEx(value T, expiration time.Duration) (bool, error) {
	return s.add("AddEx", value, expiration)
}

func (s *Set[T]) add(funcName string, value T, expiration time.Duration) (bool, error) {
	encValue, err := encode(value)
	if err != nil {
		return false, fmt.Errorf("set.%s: encode value: %w", funcName, err)
	}
	hashedKey := getHashedKey[T](encValue)

	var expiresAt int64
	if expiration != 0 {
		expiresAt = now().Add(expiration).Unix()
	}
	query := fmt.Sprintf(
		`
			INSERT INTO %s (key_hash, value, expires_at)
			VALUES (?, ?, ?)
			ON CONFLICT(key_hash) DO NOTHING
		`,
		s.tableName,
	)
	res, err := s.storage.db.Exec(query, hashedKey, encValue, expiresAt)
	if err != nil {
		return false, fmt.Errorf("set.%s: add value: %w", funcName, err)
	}

	affectedCount, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("set.%s: affected count: %w", funcName, err)
	}
	return affectedCount > 0, nil
}

// Has returns true if the value exists in the set; otherwise, false when it doesn't exist
func (s *Set[T]) Has(value T) (bool, error) {
	encValue, err := encode(value)
	if err != nil {
		return false, fmt.Errorf("set.Has: encode value: %w", err)
	}
	hashedKey := getHashedKey[T](encValue)

	var expiresAt int64
	query := fmt.Sprintf(
		`
			SELECT expires_at FROM %s
			WHERE key_hash = ?
			LIMIT 1
		`,
		s.tableName,
	)
	if err := s.storage.db.QueryRow(query, hashedKey).Scan(&expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("set.Has: has value: %w", err)
	}
	if hasKeyExpired(expiresAt) {
		return false, nil
	}
	return true, nil
}

// Delete returns true if the value was deleted from the set; otherwise, false if it wasn't deleted due to not existing
func (s *Set[T]) Delete(value T) (bool, error) {
	encValue, err := encode(value)
	if err != nil {
		return false, fmt.Errorf("set.Delete: encode value: %w", err)
	}
	hashedKey := getHashedKey[T](encValue)

	query := fmt.Sprintf(
		`
			DELETE FROM %s
			WHERE key_hash = ?
		`,
		s.tableName,
	)
	res, err := s.storage.db.Exec(query, hashedKey)
	if err != nil {
		return false, fmt.Errorf("set.Delete: delete value: %w", err)
	}

	affectedCount, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("set.Delete: affected count: %w", err)
	}
	return affectedCount > 0, nil
}

// Entries returns an iterator that iterates over all value entries in the set
func (s *Set[T]) Entries() iter.Seq2[T, T] {
	return func(yield func(T, T) bool) {
		query := fmt.Sprintf(
			`
				SELECT value FROM %s
				WHERE expires_at = 0 OR expires_at > ?
			`,
			s.tableName,
		)
		rows, err := s.storage.db.Query(query, nowUnix())
		if err != nil {
			// Ignore the error
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encValue []byte
			if err := rows.Scan(&encValue); err != nil {
				// Ignore the error
				continue
			}

			var value T
			if err := json.Unmarshal(encValue, &value); err != nil {
				// Ignore the error
				continue
			}
			if !yield(value, value) {
				return
			}
		}
	}
}

// Keys returns an iterator that iterates over all keys in the set
func (s *Set[T]) Keys() iter.Seq[T] {
	return func(yield func(T) bool) {
		for value := range s.Entries() {
			if !yield(value) {
				return
			}
		}
	}
}

// Values returns an iterator that iterates over all values in the set
func (s *Set[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for value := range s.Entries() {
			if !yield(value) {
				return
			}
		}
	}
}

// Size returns the number of values in the set
func (s *Set[T]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
			SELECT COUNT(1) FROM %s
			WHERE expires_at = 0 OR expires_at > ?
        `,
		s.tableName,
	)
	err := s.storage.db.QueryRow(query, nowUnix()).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("set.Size: get size: %w", err)
	}
	return size, nil
}

// Clear deletes all values from the set
func (s *Set[T]) Clear() error {
	query := fmt.Sprintf(
		`
			DELETE FROM %s
		`,
		s.tableName,
	)
	if _, err := s.storage.db.Exec(query); err != nil {
		return fmt.Errorf("set.Clear: clear values: %w", err)
	}
	return nil
}
