package helpers

import "strings"

// Tildify replaces the user's home directory with a tilde ("~")
func Tildify(s string) string {
	dir := UserHomeDir()
	if !strings.HasPrefix(s, dir) {
		return s
	}
	return "~" + s[len(dir):]
}

// Untildify replaces the tilde ("~") with the user's home directory
func Untildify(s string) string {
	if !strings.HasPrefix(s, "~") {
		return s
	}
	return UserHomeDir() + s[1:]
}
