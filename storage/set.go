package storage

import (
	"database/sql"
	"fmt"
	"iter"
	"strings"
	"time"
)

type Set[T comparable] struct {
	storage       *Storage
	tableName     string
	lastIterError error
}

// NewSet creates a new set which is persisted to a SQLite database
func NewSet[T comparable](s *Storage, name string) (*Set[T], error) {
	tableName := getNormalizedTableName("set", name)
	if err := execTransaction(s.db, func(tx *sql.Tx) error {
		if _, err := tx.Exec(fmt.Sprintf(
			`
				CREATE TABLE IF NOT EXISTS %s (
					key_hash TEXT PRIMARY KEY,
					value BLOB NOT NULL,
					expires_at INTEGER DEFAULT 0,
					updated_at INTEGER NOT NULL
				)
			`,
			tableName,
		)); err != nil {
			return fmt.Errorf("storage.NewSet: create set table: %w", err)
		}

		if _, err := tx.Exec(fmt.Sprintf(
			`
				CREATE INDEX IF NOT EXISTS %s_expires_idx ON %s(expires_at)
			`,
			tableName,
			tableName,
		)); err != nil {
			return fmt.Errorf("storage.NewSet: create set expires at index: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &Set[T]{
		storage:       s,
		tableName:     tableName,
		lastIterError: nil,
	}, nil
}

// Add adds a value to the set
func (s *Set[T]) Add(value T) error {
	return s.add("Add", value, 0)
}

// AddEx adds a value to the set with an expiration duration
func (s *Set[T]) AddEx(value T, expiration time.Duration) error {
	return s.add("AddEx", value, expiration)
}

func (s *Set[T]) add(funcName string, value T, expiration time.Duration) error {
	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("set.%s: encode value: %w", funcName, err)
	}
	hashedKey := getHashedKey[T](encValue)

	query := fmt.Sprintf(
		`
			INSERT INTO %s (key_hash, value, expires_at, updated_at)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(key_hash) DO UPDATE SET
				expires_at = excluded.expires_at,
				updated_at = excluded.updated_at
		`,
		s.tableName,
	)
	if _, err := s.storage.db.Exec(query, hashedKey, encValue, getKeyExpirationAsMilli(expiration), nowUnixMilli()); err != nil {
		return fmt.Errorf("set.%s: add value: %w", funcName, err)
	}
	return nil
}

// MAdd adds multiple values to the set
func (s *Set[T]) MAdd(values ...T) error {
	return s.madd("MAdd", values, 0)
}

// MAddEx adds multiple values to the set with an expiration duration
func (s *Set[T]) MAddEx(values []T, expiration time.Duration) error {
	return s.madd("MAddEx", values, expiration)
}

// SQLite default limit is 999 parameters, each row uses 4 parameters
const defaultAddChunkSize = 249

func (s *Set[T]) madd(funcName string, values []T, expiration time.Duration) error {
	if len(values) == 0 {
		return nil
	}

	return execTransaction(s.storage.db, func(tx *sql.Tx) error {
		currCount := 0
		expiresAt := getKeyExpirationAsMilli(expiration)
		now := nowUnixMilli()

		var placeholders []string
		var args []any
		for _, v := range values {
			encValue, err := encode(v)
			if err != nil {
				return fmt.Errorf("set.%s: encode value: %w", funcName, err)
			}
			hashedKey := getHashedKey[T](encValue)

			placeholders = append(placeholders, "(?, ?, ?, ?)")
			args = append(args, hashedKey, encValue, expiresAt, now)
			currCount++

			if currCount == defaultAddChunkSize {
				if err := execAddBatch(tx, s.tableName, funcName, placeholders, args); err != nil {
					return err
				}

				placeholders = nil
				args = nil
				currCount = 0
			}
		}
		if currCount > 0 {
			if err := execAddBatch(tx, s.tableName, funcName, placeholders, args); err != nil {
				return err
			}
		}
		return nil
	})
}

func execAddBatch(tx *sql.Tx, tableName, funcName string, placeholders []string, args []any) error {
	query := fmt.Sprintf(
		`
            INSERT INTO %s (key_hash, value, expires_at, updated_at)
            VALUES %s
            ON CONFLICT(key_hash) DO UPDATE SET
                expires_at = excluded.expires_at,
                updated_at = excluded.updated_at
        `,
		tableName,
		strings.Join(placeholders, ","),
	)
	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("set.%s: batch insert: %w", funcName, err)
	}
	return nil
}

// Has returns true if the value exists in the set; otherwise, false
func (s *Set[T]) Has(value T) (bool, error) {
	encValue, err := encode(value)
	if err != nil {
		return false, fmt.Errorf("set.Has: encode value: %w", err)
	}
	hashedKey := getHashedKey[T](encValue)

	query := fmt.Sprintf(
		`
			SELECT EXISTS(
				SELECT 1 FROM %s
				WHERE key_hash = ?
					AND (expires_at = 0 OR expires_at > ?)
			)
		`,
		s.tableName,
	)
	var exists bool
	if err := s.storage.db.QueryRow(query, hashedKey, nowUnixMilli()).Scan(&exists); err != nil {
		return false, fmt.Errorf("set.Has: has value: %w", err)
	}
	return exists, nil
}

// MHas checks if multiple values exist in the set and returns a map of values to their existence status.
// NOTE. The returned map contains an entry for each requested value with a boolean indicating existence.
func (s *Set[T]) MHas(values ...T) (map[T]bool, error) {
	if len(values) == 0 {
		return map[T]bool{}, nil
	}

	var placeholdersBuilder strings.Builder
	var args []any
	for i, value := range values {
		encValue, err := encode(value)
		if err != nil {
			return nil, fmt.Errorf("set.MHas: encode value %v: %w", value, err)
		}
		hashedKey := getHashedKey[T](encValue)

		if i > 0 {
			placeholdersBuilder.WriteByte(',')
		}
		placeholdersBuilder.WriteByte('?')
		args = append(args, hashedKey)
	}
	args = append(args, nowUnixMilli())

	query := fmt.Sprintf(
		`
			SELECT value FROM %s
			WHERE key_hash IN (%s)
				AND (expires_at = 0 OR expires_at > ?)
        `,
		s.tableName,
		placeholdersBuilder.String(),
	)
	rows, err := s.storage.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("set.MHas: query values: %w", err)
	}
	defer rows.Close()

	res := map[T]bool{}
	for _, value := range values {
		res[value] = false
	}
	for rows.Next() {
		var encValue []byte
		if err := rows.Scan(&encValue); err != nil {
			return nil, fmt.Errorf("set.MHas: get value: %w", err)
		}

		value, err := decode[T](encValue)
		if err != nil {
			return nil, fmt.Errorf("set.MHas: decode value: %w", err)
		}
		res[value] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("set.MHas: iterate value: %w", err)
	}
	return res, nil
}

// Delete deletes a value from the set
func (s *Set[T]) Delete(value T) error {
	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("set.Delete: encode value: %w", err)
	}
	hashedKey := getHashedKey[T](encValue)

	query := fmt.Sprintf(
		`
			DELETE FROM %s
			WHERE key_hash = ?
		`,
		s.tableName,
	)
	if _, err := s.storage.db.Exec(query, hashedKey); err != nil {
		return fmt.Errorf("set.Delete: delete value: %w", err)
	}
	return nil
}

// Entries returns an iterator that iterates over all value entries in the set.
// NOTE: As this is a set, the same value is yielded as both the key and value
// for compatibility with map-style iteration patterns
func (s *Set[T]) Entries() iter.Seq2[T, T] {
	s.lastIterError = nil
	return func(yield func(T, T) bool) {
		query := fmt.Sprintf(
			`
				SELECT value FROM %s
				WHERE expires_at = 0 OR expires_at > ?
				ORDER BY updated_at DESC
			`,
			s.tableName,
		)
		rows, err := s.storage.db.Query(query, nowUnixMilli())
		if err != nil {
			s.lastIterError = fmt.Errorf("set.Entries: query values: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encValue []byte
			if err := rows.Scan(&encValue); err != nil {
				s.lastIterError = fmt.Errorf("set.Entries: get value: %w", err)
				return
			}

			value, err := decode[T](encValue)
			if err != nil {
				s.lastIterError = fmt.Errorf("set.Entries: decode value: %w", err)
				return
			}
			if !yield(value, value) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			s.lastIterError = fmt.Errorf("set.Entries: iterate values: %w", err)
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

// IterError returns the first error encountered during the last iteration.
// NOTE: It should be called after iteration has completed
func (s *Set[T]) IterError() error {
	return s.lastIterError
}

// Size returns the number of values in the set
func (s *Set[T]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
			SELECT COUNT(*) FROM %s
			WHERE expires_at = 0 OR expires_at > ?
        `,
		s.tableName,
	)
	if err := s.storage.db.QueryRow(query, nowUnixMilli()).Scan(&size); err != nil {
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
