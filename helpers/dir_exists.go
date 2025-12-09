package helpers

import "os"

// DirExists checks if the specified directory exists and is a directory
func DirExists(name string) bool {
	f, err := os.Stat(name)
	return err == nil && f.IsDir()
}
