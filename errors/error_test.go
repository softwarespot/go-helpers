package errors

import (
	"errors"
	"testing"
)

func Test_New(t *testing.T) {
	e0 := New("non-wrapped",
		"arg0", "value0",
		"arg1", "value1",
	)
	e1 := WrapWithMessage(e0, "wrapped 1",
		"arg2", "value2",
		"arg3", "value3",
	)
	e2 := WrapWithMessage(e1, "wrapped 2 (use this error message)",
		"arg4", "value4",
		"arg5", "value5",
	)
	e3 := Wrap(e2, "wrapped 3",
		"arg6", "value6",
		"arg7", "value7",
	)

	s := e3.Error()
	t.Log("Error:", s)

	cause := Cause(e3)
	t.Log("Cause:", cause)

	trace := Trace(e3)
	t.Log("Trace:", trace)

	args := Args(e3)
	t.Log("Args:", args)

	t.Log("with std pkg error")

	e0 = errors.New("std pkg error")
	e1 = WrapWithMessage(e0, "wrapped 1 (use this error message)",
		"arg2", "value2",
		"arg3", "value3",
	)
	e2 = Wrap(e1, "wrapped 2",
		"arg4", "value4",
		"arg5", "value5",
	)

	s = e2.Error()
	t.Log("Error:", s)

	cause = Cause(e2)
	t.Log("Cause:", cause)

	trace = Trace(e2)
	t.Log("Trace:", trace)

	args = Args(e2)
	t.Log("Args:", args)
}
