package helpers

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

// ID generates a unique identifier of type T, which must be a string type i.e. string alias types can be used.
// The identifier is created by combining a random byte sequence with the current
// Unix timestamp in milliseconds, ensuring that each ID is unique.
//
// The function returns the generated ID and any error encountered during the process.
//
// Example usage:
//
//	id, err := helpers.ID()
//	if err != nil {
//		log.Fatalf("Error generating ID: %v", err)
//	}
//	fmt.Println("Generated ID:", id)
func ID[T ~string]() (T, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b[:24]); err != nil {
		return "", fmt.Errorf("creating ID: %w", err)
	}

	// 8 bytes used for the timestamp i.e. 32 - 8 = 24
	binary.BigEndian.PutUint64(b[24:], uint64(time.Now().UnixMilli()))
	return T(hex.EncodeToString(b)), nil
}
