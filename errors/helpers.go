package errors

import "strings"

// Wrap wraps the provided error with an error message and optional arguments,
// in which the error message is not returned when using [err.Error],
// but instead the original error message or an error which has been wrapped with WrapWithMessage().
// NOTE: The message can be an empty string, if only wrapping for the trace and optional arguments
// is needed
func Wrap(err error, msg string, args ...Arg) error {
	if err == nil {
		return nil
	}

	e := &Error{
		msg:        msg,
		wrappedErr: err,
		wrappedAs:  wrappedAsDefault,
		args:       args,
	}
	e.applyCallers()
	return e
}

// WrapWithMessage wraps the provided error with an error message and optional arguments,
// in which the error message is returned when using [err.Error],
// instead of the original error message
func WrapWithMessage(err error, msg string, args ...Arg) error {
	if err == nil {
		return nil
	}

	e := &Error{
		msg:        msg,
		wrappedErr: err,
		wrappedAs:  wrappedAsMessage,
		args:       args,
	}
	e.applyCallers()
	return e
}

// Args returns all structured args attached to the error chain as a single slice
func Args(err error) []Arg {
	var args []Arg
	for err != nil {
		e, ok := AsType[*Error](err)
		if !ok {
			break
		}

		args = append(args, e.args...)
		err = e.wrappedErr
	}
	return args
}

// Cause returns the original error at the start of the error chain
func Cause(err error) error {
	for err != nil {
		wrappedErr := Unwrap(err)
		if wrappedErr == nil {
			break
		}
		err = wrappedErr
	}
	return err
}

// Trace returns a string representation of the error trace chain
func Trace(err error) string {
	var traces []string
	for err != nil {
		e, ok := AsType[*Error](err)
		if !ok {
			traces = append(traces, "<external-error>["+err.Error()+"]")
			break
		}

		traces = append(traces, e.trace())
		err = e.wrappedErr
	}
	return strings.Join(traces, "~>")
}
