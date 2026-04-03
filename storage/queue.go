package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"
)

type Queue[T any] struct {
	storage       *Storage
	tableName     string
	lastIterError error
}

// NewQueue creates a new queue which is persisted to a SQLite database
func NewQueue[T any](s *Storage, name string) (*Queue[T], error) {
	tableName := getNormalizedTableName("queue", name)
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
			return fmt.Errorf("storage.NewQueue: create queue table: %w", err)
		}

		_, err = tx.Exec(fmt.Sprintf(
			`
                CREATE INDEX IF NOT EXISTS %s_expires_id_idx ON %s(expires_at, id)
            `,
			tableName,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewQueue: create queue composite index: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &Queue[T]{
		storage:       s,
		tableName:     tableName,
		lastIterError: nil,
	}, nil
}

// Enqueue adds a value to the queue
func (q *Queue[T]) Enqueue(value T) error {
	return q.enqueue("Enqueue", value, 0)
}

// EnqueueEx adds a value to the queue with an expiration duration
func (q *Queue[T]) EnqueueEx(value T, expiration time.Duration) error {
	return q.enqueue("EnqueueEx", value, expiration)
}

func (q *Queue[T]) enqueue(funcName string, value T, expiration time.Duration) error {
	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("queue.%s: encode value: %w", funcName, err)
	}

	query := fmt.Sprintf(
		`
            INSERT INTO %s (value, expires_at, created_at)
            VALUES (?, ?, ?)
        `,
		q.tableName,
	)
	if _, err = q.storage.db.Exec(query, encValue, getKeyExpirationAsMilli(expiration), nowUnixMilli()); err != nil {
		return fmt.Errorf("queue.%s: enqueue value: %w", funcName, err)
	}
	return nil
}

// Dequeue deletes and returns the oldest value from the queue
func (q *Queue[T]) Dequeue() (T, bool, error) {
	var value T
	if err := execTransaction(q.storage.db, func(tx *sql.Tx) error {
		query := fmt.Sprintf(
			`
                SELECT id, value FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY id ASC
                LIMIT 1
            `,
			q.tableName,
		)

		var id int
		var encValue []byte
		if err := tx.QueryRow(query, nowUnixMilli()).Scan(&id, &encValue); err != nil {
			return fmt.Errorf("queue.Dequeue: get oldest value: %w", err)
		}

		query = fmt.Sprintf(
			`
                DELETE FROM %s
                WHERE id = ?
            `,
			q.tableName,
		)
		if _, err := tx.Exec(query, id); err != nil {
			return fmt.Errorf("queue.Dequeue: delete value: %w", err)
		}

		decValue, err := decode[T](encValue)
		if err != nil {
			return fmt.Errorf("queue.Dequeue: decode value: %w", err)
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

// Peek returns the oldest value from the queue without removing it
func (q *Queue[T]) Peek() (T, bool, error) {
	query := fmt.Sprintf(
		`
            SELECT value FROM %s
            WHERE expires_at = 0 OR expires_at > ?
            ORDER BY id ASC
            LIMIT 1
        `,
		q.tableName,
	)
	var encValue []byte
	if err := q.storage.db.QueryRow(query, nowUnixMilli()).Scan(&encValue); err != nil {
		var value T
		if errors.Is(err, sql.ErrNoRows) {
			return value, false, nil
		}
		return value, false, fmt.Errorf("queue.Peek: get oldest value: %w", err)
	}

	value, err := decode[T](encValue)
	if err != nil {
		return value, false, fmt.Errorf("queue.Peek: decode value: %w", err)
	}
	return value, true, nil
}

// Entries returns an iterator that iterates over all value entries in the queue
func (q *Queue[T]) Entries() iter.Seq[T] {
	q.lastIterError = nil
	return func(yield func(T) bool) {
		query := fmt.Sprintf(
			`
                SELECT value FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY id ASC
            `,
			q.tableName,
		)
		rows, err := q.storage.db.Query(query, nowUnixMilli())
		if err != nil {
			q.lastIterError = fmt.Errorf("queue.Entries: query values: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encValue []byte
			if err := rows.Scan(&encValue); err != nil {
				q.lastIterError = fmt.Errorf("queue.Entries: get value: %w", err)
				return
			}

			value, err := decode[T](encValue)
			if err != nil {
				q.lastIterError = fmt.Errorf("queue.Entries: decode value: %w", err)
				return
			}
			if !yield(value) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			q.lastIterError = fmt.Errorf("queue.Entries: iterate values: %w", err)
		}
	}
}

// Values returns an iterator that iterates over all values in the queue
func (q *Queue[T]) Values() iter.Seq[T] {
	return q.Entries()
}

// IterError returns the first error encountered during the last iteration.
// NOTE: It should be called after iteration has completed
func (q *Queue[T]) IterError() error {
	return q.lastIterError
}

// Size returns the number of values in the queue
func (q *Queue[T]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
            SELECT COUNT(*) FROM %s
            WHERE expires_at = 0 OR expires_at > ?
        `,
		q.tableName,
	)
	if err := q.storage.db.QueryRow(query, nowUnixMilli()).Scan(&size); err != nil {
		return 0, fmt.Errorf("queue.Size: get size: %w", err)
	}
	return size, nil
}

// Clear deletes all values from the queue
func (q *Queue[T]) Clear() error {
	query := fmt.Sprintf(
		`
            DELETE FROM %s
        `,
		q.tableName,
	)
	if _, err := q.storage.db.Exec(query); err != nil {
		return fmt.Errorf("queue.Clear: clear values: %w", err)
	}
	return nil
}
