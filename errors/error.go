package errors

import (
	"path/filepath"
	"runtime"
	"strconv"
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
	args       []Arg

	fileName   string
	funcName   string
	lineNumber int
}

func New(msg string, args ...Arg) error {
	e := &Error{
		msg:        msg,
		wrappedErr: nil,
		wrappedAs:  wrappedAsMessage,
		args:       args,

		fileName:   "",
		funcName:   "",
		lineNumber: 0,
	}
	e.applyCallers()
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

func (e *Error) applyCallers() {
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
