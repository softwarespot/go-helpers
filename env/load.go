package env

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Load reads environment variables from a specified file and sets them in the current process's environment.
//
// Example usage:
//
//	err := Load(os.DirFS("."), ".env")
//	if err != nil {
//	    log.Fatalf("error loading .env: %+v", err)
//	}
func Load(fs fs.FS, path string) error {
	f, err := fs.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open the env file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for lineNo := 0; scanner.Scan(); {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid env line format at line %d: %s", lineNo, line)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("unable to set env variable at line %d, key: %s: %w", lineNo, key, err)
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("unable to scan the env file: %w", err)
	}
	return nil
}
