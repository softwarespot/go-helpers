package errors

import (
	"errors"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Ensure interface compatibility
var _ error = &Error{}

type wrappedAs string

const (
	wrappedAsDefault wrappedAs = "wrapped-as-default"
	wrappedAsMessage wrappedAs = "wrapped-as-message"
)

type Error struct {
	msg        string
	wrappedErr error
	wrappedAs  wrappedAs
	args       []any

	fileName   string
	funcName   string
	lineNumber int
}

func New(msg string, args ...any) error {
	e := &Error{
		msg:        msg,
		wrappedErr: nil,
		wrappedAs:  wrappedAsMessage,
		args:       args,

		fileName:   "",
		funcName:   "",
		lineNumber: 0,
	}
	applyCaller(e)
	return e
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	// Not a wrapped error i.e. the original error
	if e.wrappedErr == nil {
		return e.msg
	}

	if e.wrappedAs == wrappedAsMessage {
		return e.msg
	}

	// Recursively call down the error chain to either the next "wrapped as message" or original error
	return e.wrappedErr.Error()
}

func (e *Error) Is(err error) bool {
	return err == e
}

func (e *Error) Unwrap() error {
	return e.wrappedErr
}

func (e *Error) trace() string {
	return e.funcName + "[" + e.msg + "]" + e.fileName + ":" + strconv.Itoa(e.lineNumber)
}

// As is the same as errors.As() in the Go standard package
func As(err error, target any) bool {
	return errors.As(err, target)
}

// AsType is the same as errors.AsType() in the Go standard package
func AsType[T any](err error) (T, bool) {
	// TODO: For Go 1.26, replace with errors.AsType[T](err)
	var target T
	ok := errors.As(err, &target)
	return target, ok
}

// Is is the same as errors.Is() in the Go standard package
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Join is the same as errors.Join() in the Go standard package
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// Wrap wraps the provided error with an error message and optional arguments,
// in which the error message is not returned when using err.Error(),
// but instead the original error message or an error which has been wrapped with WrapWithMessage().
// NOTE: The message can be an empty string, if only wrapping for the trace and optional arguments
// is needed
func Wrap(err error, msg string, args ...any) error {
	if err == nil {
		return nil
	}

	e := &Error{
		msg:        msg,
		wrappedErr: err,
		wrappedAs:  wrappedAsDefault,
		args:       args,
	}
	applyCaller(e)
	return e
}

// WrapWithMessage wraps the provided error with an error message and optional arguments,
// in which the error message is returned when using err.Error(),
// instead of the original error message
func WrapWithMessage(err error, msg string, args ...any) error {
	if err == nil {
		return nil
	}

	e := &Error{
		msg:        msg,
		wrappedErr: err,
		wrappedAs:  wrappedAsMessage,
		args:       args,
	}
	applyCaller(e)
	return e
}

// Unwrap is the same as errors.Unwrap() in the Go standard package
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Cause returns the original error at the start of the error chain
func Cause(err error) error {
	for err != nil {
		wrappedErr := errors.Unwrap(err)
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

// Args returns all the arguments attached to the error chain as a single slice
func Args(err error) []any {
	var args []any
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

func applyCaller(e *Error) {
	stack := make([]uintptr, 4)
	count := runtime.Callers(3, stack)
	if count == 0 {
		return
	}

	frames := runtime.CallersFrames(stack[:count])
	if frame, _ := frames.Next(); frame.Function != "" {
		e.fileName = frame.File
		e.funcName = filepath.Base(frame.Function)
		e.lineNumber = frame.Line
	}
}
