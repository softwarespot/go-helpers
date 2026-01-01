package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"
)

type List[T any] struct {
	storage       *Storage
	tableName     string
	lastIterError error
}

// NewList creates a new list which is persisted to a SQLite database
func NewList[T any](s *Storage, name string) (*List[T], error) {
	tableName := getNormalizedTableName("list", name)
	if err := execTransaction(s.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(fmt.Sprintf(
			`
                CREATE TABLE IF NOT EXISTS %s (
                    position INTEGER NOT NULL,
                    value BLOB NOT NULL,
                    expires_at INTEGER DEFAULT 0,
                    created_at INTEGER NOT NULL,
                    PRIMARY KEY (position)
                )
            `,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewList: create list table: %w", err)
		}

		_, err = tx.Exec(fmt.Sprintf(
			`
				CREATE INDEX IF NOT EXISTS %s_expires_pos_idx ON %s(expires_at, position)
			`,
			tableName,
			tableName,
		))
		if err != nil {
			return fmt.Errorf("storage.NewList: create list composite index: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	s.registerTable(tableName)

	return &List[T]{
		storage:       s,
		tableName:     tableName,
		lastIterError: nil,
	}, nil
}

// Append adds a value to the end of the list
func (l *List[T]) Append(value T) error {
	return l.appendEx("Append", value, 0)
}

// AppendEx adds a value to the end of the list with an expiration duration
func (l *List[T]) AppendEx(value T, expiration time.Duration) error {
	return l.appendEx("AppendEx", value, expiration)
}

func (l *List[T]) appendEx(funcName string, value T, expiration time.Duration) error {
	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("list.%s: encode value: %w", funcName, err)
	}

	return execTransaction(l.storage.db, func(tx *sql.Tx) error {
		var nextPos int
		query := fmt.Sprintf(
			`
				SELECT COALESCE(MAX(position) + 1, 0) FROM %s
			`,
			l.tableName,
		)
		if err := tx.QueryRow(query).Scan(&nextPos); err != nil {
			return fmt.Errorf("list.%s: get next position: %w", funcName, err)
		}

		query = fmt.Sprintf(
			`
            INSERT INTO %s (position, value, expires_at, created_at)
            VALUES (?, ?, ?, ?)
        `,
			l.tableName,
		)
		if _, err = tx.Exec(
			query,
			nextPos,
			encValue,
			getKeyExpirationAsMilli(expiration),
			nowUnixMilli(),
		); err != nil {
			return fmt.Errorf("list.%s: append value: %w", funcName, err)
		}
		return nil
	})
}

// Get returns the value at the specified position
func (l *List[T]) Get(position int) (T, bool, error) {
	var value T
	query := fmt.Sprintf(
		`
            SELECT value FROM %s
            WHERE position = ?
				AND (expires_at = 0 OR expires_at > ?)
        `,
		l.tableName,
	)
	var encValue []byte
	if err := l.storage.db.QueryRow(query, position, nowUnixMilli()).Scan(&encValue); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return value, false, nil
		}
		return value, false, fmt.Errorf("list.Get: get value at position: %w", err)
	}

	value, err := decode[T](encValue)
	if err != nil {
		return value, false, fmt.Errorf("list.Get: decode value: %w", err)
	}
	return value, true, nil
}

// Set updates the value at the specified position
func (l *List[T]) Set(position int, value T) error {
	return l.setEx("Set", position, value, 0)
}

// SetEx updates the value at the specified position with an expiration duration
func (l *List[T]) SetEx(position int, value T, expiration time.Duration) error {
	return l.setEx("SetEx", position, value, expiration)
}

func (l *List[T]) setEx(funcName string, position int, value T, expiration time.Duration) error {
	encValue, err := encode(value)
	if err != nil {
		return fmt.Errorf("list.%s: encode value: %w", funcName, err)
	}

	query := fmt.Sprintf(
		`
            UPDATE %s
            SET value = ?, expires_at = ?, created_at = ?
            WHERE position = ?
        `,
		l.tableName,
	)
	result, err := l.storage.db.Exec(
		query,
		encValue,
		getKeyExpirationAsMilli(expiration),
		nowUnixMilli(),
		position,
	)
	if err != nil {
		return fmt.Errorf("list.%s: set value: %w", funcName, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("list.%s: get affected rows: %w", funcName, err)
	}
	if affected == 0 {
		return fmt.Errorf("list.%s: position %d not found", funcName, position)
	}

	return nil
}

// Delete deletes the value at the specified position
func (l *List[T]) Delete(position int) error {
	query := fmt.Sprintf(
		`
			DELETE FROM %s
			WHERE position = ?
		`,
		l.tableName,
	)
	result, err := l.storage.db.Exec(query, position)
	if err != nil {
		return fmt.Errorf("list.Remove: delete value: %w", err)
	}

	affectedCount, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("list.Remove: get affected rows: %w", err)
	}
	if affectedCount == 0 {
		return fmt.Errorf("list.Remove: position %d not found", position)
	}

	// Reindex the remaining items
	err = execTransaction(l.storage.db, func(tx *sql.Tx) error {
		tempTableName := getNormalizedTableName("temp", l.tableName, fmt.Sprintf("%d", time.Now().UnixNano()))
		_, err := tx.Exec(fmt.Sprintf(
			`
                CREATE TEMPORARY TABLE %s AS
                SELECT ROW_NUMBER() OVER (ORDER BY position) - 1 AS new_position, value, expires_at, created_at
                FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY position
            `,
			tempTableName,
			l.tableName,
		), nowUnixMilli())
		if err != nil {
			return fmt.Errorf("list.Remove: create temporary table: %w", err)
		}

		_, err = tx.Exec(fmt.Sprintf(`
			DELETE FROM %s
		`,
			l.tableName,
		))
		if err != nil {
			return fmt.Errorf("list.Remove: clear list table: %w", err)
		}

		_, err = tx.Exec(fmt.Sprintf(
			`
                INSERT INTO %s (position, value, expires_at, created_at)
                SELECT new_position, value, expires_at, created_at
                FROM %s
            `,
			l.tableName, tempTableName,
		))
		if err != nil {
			return fmt.Errorf("list.Remove: reindex values: %w", err)
		}

		_, err = tx.Exec(fmt.Sprintf(`DROP TABLE %s`, tempTableName))
		if err != nil {
			return fmt.Errorf("list.Remove: drop temporary table: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("list.Remove: reindex: %w", err)
	}
	return nil
}

// Entries returns an iterator that iterates over all value entries in position order in the list
func (l *List[T]) Entries() iter.Seq[T] {
	l.lastIterError = nil
	return func(yield func(T) bool) {
		query := fmt.Sprintf(
			`
                SELECT value FROM %s
                WHERE expires_at = 0 OR expires_at > ?
                ORDER BY position ASC
            `,
			l.tableName,
		)
		rows, err := l.storage.db.Query(query, nowUnixMilli())
		if err != nil {
			l.lastIterError = fmt.Errorf("list.Entries: query values: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var encValue []byte
			if err := rows.Scan(&encValue); err != nil {
				l.lastIterError = fmt.Errorf("list.Entries: get value: %w", err)
				return
			}

			value, err := decode[T](encValue)
			if err != nil {
				l.lastIterError = fmt.Errorf("list.Entries: decode value: %w", err)
				return
			}
			if !yield(value) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			l.lastIterError = fmt.Errorf("list.Entries: iterate values: %w", err)
		}
	}
}

// Values returns an iterator that iterates over all values in position order in the list
func (l *List[T]) Values() iter.Seq[T] {
	return l.Entries()
}

// IterError returns the first error encountered during the last iteration.
// NOTE: It should be called after iteration has completed
func (l *List[T]) IterError() error {
	return l.lastIterError
}

// Size returns the number of values in the list
func (l *List[T]) Size() (int, error) {
	var size int
	query := fmt.Sprintf(
		`
            SELECT COUNT(*) FROM %s
            WHERE expires_at = 0 OR expires_at > ?
        `,
		l.tableName,
	)
	if err := l.storage.db.QueryRow(query, nowUnixMilli()).Scan(&size); err != nil {
		return 0, fmt.Errorf("list.Size: get size: %w", err)
	}
	return size, nil
}

// Clear deletes all values from the list
func (l *List[T]) Clear() error {
	query := fmt.Sprintf(
		`
			DELETE FROM %s
		`,
		l.tableName,
	)
	if _, err := l.storage.db.Exec(query); err != nil {
		return fmt.Errorf("list.Clear: clear values: %w", err)
	}
	return nil
}
