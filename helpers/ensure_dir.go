package helpers

import (
	"fmt"
	"os"
)

// EnsureDir ensures that the specified directory exists, creating it if it does not
func EnsureDir(name string, mode os.FileMode) error {
	if err := os.MkdirAll(name, mode); err != nil && !os.IsExist(err) {
		return fmt.Errorf("unable to create the directory %q: %w", name, err)
	}
	return nil
}
