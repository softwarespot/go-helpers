package helpers

import "time"

// Taken from URL: https://github.com/matryer/try/blob/master/try.go

// Retry retries a function on error, and continues until successful
// or the maximum number of retries has exceeded. The last function error
// is returned, if the maximum number of retries is exceeded
func Retry(fn func(iter int) error, retries int, retriesWait time.Duration) error {
	if retries <= 0 {
		retries = 1
	}

	for iter := 1; ; iter++ {
		err := fn(iter)
		if err == nil {
			return nil
		}
		if iter >= retries {
			return err
		}
		time.Sleep(retriesWait)
	}
}
