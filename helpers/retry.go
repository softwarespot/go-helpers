package helpers

import "time"

// Taken from URL: https://github.com/matryer/try/blob/master/try.go

// Retry retries a function on error, and continues until successful
// or the maximum number of retries has exceeded. The last function error
// is returned, if the maximum number of retries is exceeded
func Retry(fn func(attempt int) error, retries int, retriesWait time.Duration) error {
	for attempt := 1; ; {
		err := fn(attempt)
		if err == nil {
			return nil
		}
		if retries <= 0 || attempt == retries {
			return err
		}

		time.Sleep(retriesWait)
		attempt++
	}
}
