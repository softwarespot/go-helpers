package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"
)

type Stack[T any] struct {
	storage       *Storage
	tableName     string
	lastIterError error
}

// NewStack creates a new stack which is persisted to a SQLite database
func NewStack[T any](s *Storage, name string) (*Stack[T], error) {
	tableName := getNormalizedTableName("stack", name)
	if err := execTransaction(s.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(fmt.Sprintf(
			`
                CREATE TABLE IF NOT EXISTS %s (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    value BLOB NOT NULL,
                    expires_at INTEGER DEFAULT 0,
                    created_at INTEGER NOT NULL
                )
            `,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewStack: create stack table: %w", err)
		}

		_, err = tx.Exec(fmt.Sprintf(
			`
                CREATE INDEX IF NOT EXISTS %s_expires_id_idx ON %s(expires_at, id)
            `,
			tableName,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewStack: create stack composite index: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &Stack[T]{
		storage:       s,
		tableName:     tableName,
		lastIterError: nil,
	}, nil
}

// Push adds a value to the top of the stack
func (s *Stack[T]) Push(value T) error {
	return s.push("Push", value, 0)
}

// PushEx adds a value to the top of the stack with an expiration duration
func (s *Stack[T]) PushEx(value T, expiration time.Duration) error {
	return s.push("PushEx", value, expiration)
}
 
func (s *Stack[T]) push(funcName string, value T, expiration time.Duration) error {
	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("stack.%s: encode value: %w", funcName, err)
	}

	query := fmt.Sprintf(
		`
            INSERT INTO %s (value, expires_at, created_at)
            VALUES (?, ?, ?)
        `,
		s.tableName,
	)
	if _, err = s.storage.db.Exec(query, encValue, getKeyExpirationAsMilli(expiration), nowUnixMilli()); err != nil {
		return fmt.Errorf("stack.%s: push value: %w", funcName, err)
	}
	return nil
}

// Pop deletes and returns the most recently added value from the stack
func (s *Stack[T]) Pop() (T, bool, error) {
	var value T
	if err := execTransaction(s.storage.db, func(tx *sql.Tx) error {
		query := fmt.Sprintf(
			`
                SELECT id, value FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY id DESC
                LIMIT 1
            `,
			s.tableName,
		)

		var id int
		var encValue []byte
		if err := tx.QueryRow(query, nowUnixMilli()).Scan(&id, &encValue); err != nil {
			return fmt.Errorf("stack.Pop: get newest value: %w", err)
		}

		query = fmt.Sprintf(
			`
                DELETE FROM %s
                WHERE id = ?
            `,
			s.tableName,
		)
		if _, err := tx.Exec(query, id); err != nil {
			return fmt.Errorf("stack.Pop: delete value: %w", err)
		}

		decValue, err := decode[T](encValue)
		if err != nil {
			return fmt.Errorf("stack.Pop: decode value: %w", err)
		}
		value = decValue

		return nil
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return value, false, nil
		}
		return value, false, err
	}
	return value, true, nil
}

// Peek returns the most recently added value from the stack without removing it
func (s *Stack[T]) Peek() (T, bool, error) {
	query := fmt.Sprintf(
		`
            SELECT value FROM %s
            WHERE expires_at = 0 OR expires_at > ?
            ORDER BY id DESC
            LIMIT 1
        `,
		s.tableName,
	)
	var encValue []byte
	if err := s.storage.db.QueryRow(query, nowUnixMilli()).Scan(&encValue); err != nil {
		var value T
		if errors.Is(err, sql.ErrNoRows) {
			return value, false, nil
		}
		return value, false, fmt.Errorf("stack.Peek: get newest value: %w", err)
	}

	value, err := decode[T](encValue)
	if err != nil {
		return value, false, fmt.Errorf("stack.Peek: decode value: %w", err)
	}
	return value, true, nil
}

// Entries returns an iterator that iterates over all value entries in the stack (top to bottom)
func (s *Stack[T]) Entries() iter.Seq[T] {
	s.lastIterError = nil
	return func(yield func(T) bool) {
		query := fmt.Sprintf(
			`
                SELECT value FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY id DESC
            `,
			s.tableName,
		)
		rows, err := s.storage.db.Query(query, nowUnixMilli())
		if err != nil {
			s.lastIterError = fmt.Errorf("stack.Entries: query values: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encValue []byte
			if err := rows.Scan(&encValue); err != nil {
				s.lastIterError = fmt.Errorf("stack.Entries: get value: %w", err)
				return
			}

			value, err := decode[T](encValue)
			if err != nil {
				s.lastIterError = fmt.Errorf("stack.Entries: decode value: %w", err)
				return
			}
			if !yield(value) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			s.lastIterError = fmt.Errorf("stack.Entries: iterate values: %w", err)
		}
	}
}

// Values returns an iterator that iterates over all values in the stack (top to bottom)
func (s *Stack[T]) Values() iter.Seq[T] {
	return s.Entries()
}

// IterError returns the first error encountered during the last iteration.
// NOTE: It should be called after iteration has completed
func (s *Stack[T]) IterError() error {
	return s.lastIterError
}

// Size returns the number of values in the stack
func (s *Stack[T]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
            SELECT COUNT(*) FROM %s
            WHERE expires_at = 0 OR expires_at > ?
        `,
		s.tableName,
	)
	if err := s.storage.db.QueryRow(query, nowUnixMilli()).Scan(&size); err != nil {
		return 0, fmt.Errorf("stack.Size: get size: %w", err)
	}
	return size, nil
}

// Clear deletes all values from the stack
func (s *Stack[T]) Clear() error {
	query := fmt.Sprintf(
		`
            DELETE FROM %s
        `,
		s.tableName,
	)
	if _, err := s.storage.db.Exec(query); err != nil {
		return fmt.Errorf("stack.Clear: clear values: %w", err)
	}
	return nil
}
