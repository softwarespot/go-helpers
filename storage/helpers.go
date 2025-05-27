package storage

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func now() time.Time {
	return time.Now()
}

func nowUnix() int64 {
	return time.Now().Unix()
}

func hasKeyExpired(expiresAt int64) bool {
	if expiresAt == 0 {
		return false
	}
	return expiresAt < nowUnix()
}

func encode[T any](v T) ([]byte, error) {
	encValue, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}
	return encValue, nil
}

func decode[T any](encValue []byte) (T, error) {
	var v T
	if err := json.Unmarshal(encValue, &v); err != nil {
		return v, fmt.Errorf("decoding: %w", err)
	}
	return v, nil
}

func getHashedKey[T comparable](ev []byte) string {
	hasher := sha1.New()
	hasher.Write(ev)
	return hex.EncodeToString(hasher.Sum(nil))
}

func execTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err = fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback transaction: %w; commit transaction: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback transaction: %w; commit transaction: %w", rbErr, err)
		}
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
