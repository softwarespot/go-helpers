package helpers

import (
	"errors"
	"slices"
)

func SliceChunkFunc[S ~[]E, E comparable](s S, size int, fn func(S) error) error {
	if size <= 0 {
		return errors.New("size cannot be less than or equal to zero")
	}
	for v := range slices.Chunk(s, size) {
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}
