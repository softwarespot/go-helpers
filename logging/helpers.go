package logging

import (
	"os"
	"runtime/debug"
	"time"
)

func createEntry(msg string, level Level, args ...any) map[string]any {
	entry := map[string]any{}

	if len(args) > 0 && len(args)%2 != 0 {
		args = append(args, "%ARGS NOT DIVISIBLE BY 2%")
	}
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			// Skip keys which are not a string
			continue
		}
		entry[key] = args[i+1]
	}

	entry["@timestamp"] = time.Now().UTC().Format(time.RFC3339)
	entry["@message"] = msg
	entry["@level"] = level
	entry["@env"] = os.Getenv("ENV")

	if level.IsSevere() {
		entry["@stack-trace"] = string(debug.Stack())
	}
	return entry
}
