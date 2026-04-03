package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func getNormalizedTableName(names ...string) string {
	var sb strings.Builder
	for i, name := range names {
		for _, r := range strings.ToLower(name) {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				sb.WriteRune(r)
			} else {
				// If an invalid character, then replace with underscore
				sb.WriteRune('_')
			}
		}
		if i < len(names)-1 {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

func nowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

func getKeyExpirationAsMilli(expiration time.Duration) int64 {
	if expiration == 0 {
		return 0
	}
	return time.Now().Add(expiration).UnixMilli()
}

func hasKeyExpired(expiresAt int64) bool {
	if expiresAt == 0 {
		return false
	}
	return expiresAt < nowUnixMilli()
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
	hasher := sha256.New()
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
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
