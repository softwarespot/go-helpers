package helpers

import (
	"errors"
	"fmt"
)

// RecoverFunc is a helper function that recovers from panics in a given function.
// It takes a function `fn` that accepts an error parameter. If a panic occurs,
// it captures the panic value, wraps it as an error, and passes it to the provided function.
//
// If the panic value is not an error, it converts it to a string and creates a new error.
// If the recovered error is nil, it creates a default "unexpected nil error".
//
// Example usage:
//
//	defer RecoverFunc(func(err error) {
//		if err != nil {
//		    fmt.Println("Recovered from panic:", err)
//	    }
//	})
func RecoverFunc(fn func(err error)) {
	if r := recover(); r != nil {
		var err error
		switch e := r.(type) {
		case error:
			err = fmt.Errorf("recovered panic: %w", e)
		default:
			err = fmt.Errorf("%v", e)
		}
		if err == nil {
			err = errors.New("unexpected nil error")
		}
		fn(err)
	}
}
