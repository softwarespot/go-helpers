package logging

import "runtime"

// Memory logs memory statistics using the provided logger.
// It retrieves the current memory allocation statistics and logs them
// along with any additional arguments provided
func Memory(logger Logger, msg string, logArgs []any) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	logger.Log(
		msg,
		LevelNotice,
		append(
			[]any{
				"alloc-mb", m.Alloc / 1024 / 1024,
				"total-alloc-mb", m.TotalAlloc / 1024 / 1024,
				"sys-mb", m.Sys / 1024 / 1024,
				"num-gc", m.NumGC,
			},
			logArgs...,
		)...,
	)
}
