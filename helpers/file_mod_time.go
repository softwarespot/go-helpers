package helpers

import (
	"os"
	"time"
)

// FileModTime gets the modified time of the file path
func FileModTime(name string) time.Time {
	if f, err := os.Stat(name); err == nil && !f.IsDir() {
		return f.ModTime()
	}
	return time.Time{}
}
