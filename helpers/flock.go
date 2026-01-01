package helpers

import (
	"errors"
	"fmt"
	"time"

	goflock "github.com/gofrs/flock"
)

// ErrFlockTimeout is returned when obtaining advisory lock on a file descriptor times out
var ErrFlockTimeout = errors.New("flock timeout")

type Flock struct {
	path string

	// The time elapsed between consecutive file locking attempts
	retryBackoff time.Duration
	flock        *goflock.Flock
}

// Original implementation was taken from URL: https://github.com/etcd-io/bbolt/blob/master/bolt_unix.go

// NewFlock creates a new Flock instance for the specified file path.
// It initializes the Flock with a default retry backoff of 64 milliseconds
func NewFlock(path string) *Flock {
	return &Flock{
		path:         path,
		retryBackoff: 64 * time.Millisecond,
		flock:        goflock.New(path),
	}
}

// Lock attempts to acquire a lock on the file, either exclusive or shared.
// It retries until the lock is acquired or the timeout is reached
func (f *Flock) Lock(exclusive bool, timeout time.Duration) error {
	var (
		expiry = timeout - f.retryBackoff
		now    = time.Now()
	)
	for {
		var (
			ok  bool
			err error
		)
		if exclusive {
			ok, err = f.flock.TryLock()
		} else {
			ok, err = f.flock.TryRLock()
		}
		if ok {
			return nil
		}
		if err != nil {
			return fmt.Errorf("unable to lock the path %q: %w", f.path, err)
		}

		if timeout > 0 && time.Since(now) > expiry {
			return ErrFlockTimeout
		}
		time.Sleep(f.retryBackoff)
	}
}

// Unlock releases the acquired lock on the file
func (f *Flock) Unlock() error {
	if err := f.flock.Unlock(); err != nil {
		return fmt.Errorf("unable to unlock the path %q: %w", f.path, f.flock.Unlock())
	}
	return nil
}
