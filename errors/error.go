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

func (e *Error) Unwrap() error {
	return e.wrappedErr
}

func getTrace(e *Error) string {
	return e.funcName + "[" + e.msg + "]" + e.fileName + ":" + strconv.Itoa(e.lineNumber)
}

// As is the same as errors.As() in the Go standard package
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is is the same as errors.Is() in the Go standard package
func Is(err, target error) bool {
	return errors.Is(err, target)
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

func Trace(err error) string {
	var traces []string
	for err != nil {
		var e *Error
		if !As(err, &e) {
			traces = append(traces, "<external-error>["+err.Error()+"]")
			break
		}

		traces = append(traces, getTrace(e))
		err = e.wrappedErr
	}
	return strings.Join(traces, "~>")
}

func Args(err error) []any {
	var args []any
	for err != nil {
		var e *Error
		if !As(err, &e) {
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
	if frame, more := frames.Next(); more {
		e.fileName = frame.File
		e.funcName = filepath.Base(frame.Function)
		e.lineNumber = frame.Line
	}
}
