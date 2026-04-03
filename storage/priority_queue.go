package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"
)

type Priority int

type PriorityQueue[T any] struct {
	storage       *Storage
	tableName     string
	lastIterError error
}

// NewPriorityQueue creates a new priority queue which is persisted to a SQLite database
func NewPriorityQueue[T any](s *Storage, name string) (*PriorityQueue[T], error) {
	tableName := getNormalizedTableName("pqueue", name)
	if err := execTransaction(s.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(fmt.Sprintf(
			`
                CREATE TABLE IF NOT EXISTS %s (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    value BLOB NOT NULL,
                    priority INTEGER NOT NULL,
                    expires_at INTEGER DEFAULT 0,
                    created_at INTEGER NOT NULL
                )
            `,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewPriorityQueue: create priority queue table: %w", err)
		}

		_, err = tx.Exec(fmt.Sprintf(
			`
				CREATE INDEX IF NOT EXISTS %s_expires_priority_id_idx ON %s(expires_at, priority DESC, id ASC)
			`,
			tableName,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewPriorityQueue: create priority queue dequeue index: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &PriorityQueue[T]{
		storage:       s,
		tableName:     tableName,
		lastIterError: nil,
	}, nil
}

// Enqueue adds a value to the priority queue with the specified priority
func (pq *PriorityQueue[T]) Enqueue(value T, priority Priority) error {
	return pq.enqueueEx("Enqueue", value, priority, 0)
}

// Enqueue adds a value to the priority queue with the specified priority and expiration duration.
func (pq *PriorityQueue[T]) EnqueueEx(value T, priority Priority, expiration time.Duration) error {
	return pq.enqueueEx("EnqueueEx", value, priority, expiration)
}

func (pq *PriorityQueue[T]) enqueueEx(funcName string, value T, priority Priority, expiration time.Duration) error {
	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("priorityQueue.%s: encode value: %w", funcName, err)
	}

	query := fmt.Sprintf(
		`
            INSERT INTO %s (value, priority, expires_at, created_at)
            VALUES (?, ?, ?, ?)
        `,
		pq.tableName,
	)
	if _, err = pq.storage.db.Exec(
		query,
		encValue,
		priority,
		getKeyExpirationAsMilli(expiration),
		nowUnixMilli(),
	); err != nil {
		return fmt.Errorf("priorityQueue.%s: enqueue value: %w", funcName, err)
	}
	return nil
}

// Dequeue deletes and returns the highest priority value from the priority queue.
// NOTE: When multiple values have the same priority, the oldest value is returned first
func (pq *PriorityQueue[T]) Dequeue() (T, bool, error) {
	var value T
	if err := execTransaction(pq.storage.db, func(tx *sql.Tx) error {
		query := fmt.Sprintf(
			`
                SELECT id, value FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY priority DESC, id ASC
                LIMIT 1
            `,
			pq.tableName,
		)

		var id int
		var encValue []byte
		if err := tx.QueryRow(query, nowUnixMilli()).Scan(&id, &encValue); err != nil {
			return fmt.Errorf("priorityQueue.Dequeue: get highest priority value: %w", err)
		}

		query = fmt.Sprintf(
			`
				DELETE FROM %s
				WHERE id = ?
			`,
			pq.tableName,
		)
		if _, err := tx.Exec(query, id); err != nil {
			return fmt.Errorf("priorityQueue.Dequeue: delete value: %w", err)
		}

		decValue, err := decode[T](encValue)
		if err != nil {
			return fmt.Errorf("priorityQueue.Dequeue: decode value: %w", err)
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

// Peek returns the highest priority value from the priority queue without removing it
func (pq *PriorityQueue[T]) Peek() (T, bool, error) {
	query := fmt.Sprintf(
		`
            SELECT value FROM %s
            WHERE expires_at = 0 OR expires_at > ?
            ORDER BY priority DESC, id ASC
            LIMIT 1
        `,
		pq.tableName,
	)
	var encValue []byte
	if err := pq.storage.db.QueryRow(query, nowUnixMilli()).Scan(&encValue); err != nil {
		var value T
		if errors.Is(err, sql.ErrNoRows) {
			return value, false, nil
		}
		return value, false, fmt.Errorf("priorityQueue.Peek: get highest priority value: %w", err)
	}

	value, err := decode[T](encValue)
	if err != nil {
		return value, false, fmt.Errorf("priorityQueue.Peek: decode value: %w", err)
	}
	return value, true, nil
}

// Entries returns an iterator that iterates over all value entries in priority order in the priority queue
func (pq *PriorityQueue[T]) Entries() iter.Seq[T] {
	pq.lastIterError = nil
	return func(yield func(T) bool) {
		query := fmt.Sprintf(
			`
                SELECT value FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY priority DESC, id ASC
            `,
			pq.tableName,
		)
		rows, err := pq.storage.db.Query(query, nowUnixMilli())
		if err != nil {
			pq.lastIterError = fmt.Errorf("priorityQueue.Entries: query values: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encValue []byte
			if err := rows.Scan(&encValue); err != nil {
				pq.lastIterError = fmt.Errorf("priorityQueue.Entries: get value: %w", err)
				return
			}

			value, err := decode[T](encValue)
			if err != nil {
				pq.lastIterError = fmt.Errorf("priorityQueue.Entries: decode value: %w", err)
				return
			}
			if !yield(value) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			pq.lastIterError = fmt.Errorf("priorityQueue.Entries: iterate values: %w", err)
		}
	}
}

// Values returns an iterator that iterates over all values in priority order in the priority queue
func (pq *PriorityQueue[T]) Values() iter.Seq[T] {
	return pq.Entries()
}

// IterError returns the first error encountered during the last iteration.
// NOTE: It should be called after iteration has completed
func (pq *PriorityQueue[T]) IterError() error {
	return pq.lastIterError
}

// Size returns the number of values in the priority queue
func (pq *PriorityQueue[T]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
            SELECT COUNT(*) FROM %s
            WHERE expires_at = 0 OR expires_at > ?
        `,
		pq.tableName,
	)
	if err := pq.storage.db.QueryRow(query, nowUnixMilli()).Scan(&size); err != nil {
		return 0, fmt.Errorf("priorityQueue.Size: get size: %w", err)
	}
	return size, nil
}

// Clear deletes all values from the priority queue
func (pq *PriorityQueue[T]) Clear() error {
	query := fmt.Sprintf(
		`
			DELETE FROM %s
		`,
		pq.tableName,
	)
	if _, err := pq.storage.db.Exec(query); err != nil {
		return fmt.Errorf("priorityQueue.Clear: clear values: %w", err)
	}
	return nil
}
