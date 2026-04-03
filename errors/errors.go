package errors

import "errors"

// As is the same as [errors.As] in the Go standard package
func As(err error, target any) bool {
	return errors.As(err, target)
}

// AsType is the same as [errors.AsType] in the Go standard package
func AsType[E error](err error) (E, bool) {
	return errors.AsType[E](err)
}

// Is is the same as [errors.Is] in the Go standard package
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Join is the same as [errors.Join] in the Go standard package
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// Unwrap is the same as [errors.Unwrap] in the Go standard package
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
