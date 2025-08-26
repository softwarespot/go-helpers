package helpers

import "os"

// FileExists checks if the specified file path exists and is not a directory
func FileExists(name string) bool {
	f, err := os.Stat(name)
	return err == nil && !f.IsDir()
}
