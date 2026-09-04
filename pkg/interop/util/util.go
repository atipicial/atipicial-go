/*
Package util contains some special useful functions that are provided by compiler and VM.
*/
package util

import (
	"github.com/atipicial/atipicial-go/pkg/interop/atipicialgointernal"
)

// Abort terminates current execution, unlike exception throwing with panic() it
// can't be recovered from.
func Abort() {
	atipicialgointernal.Opcode0NoReturn("ABORT")
}

// AbortMsg terminates current execution with the specified message. Unlike
// exception throwing with panic() it can't be recovered from.
func AbortMsg(msg string) {
	atipicialgointernal.Opcode1NoReturn("ABORTMSG", msg)
}

// Assert terminates current execution if the condition specified is false. Unlike
// exception throwing with panic() it can't be recovered from.
func Assert(ok bool) {
	atipicialgointernal.Opcode1NoReturn("ASSERT", ok)
}

// AssertMsg terminates current execution with the specified message if the
// condition specified is false. Unlike exception throwing with panic() it can't
// be recovered from.
func AssertMsg(ok bool, msg string) {
	atipicialgointernal.Opcode2NoReturn("ASSERTMSG", ok, msg)
}

// Equals compares a with b and will return true when a and b are equal. It's
// implemented as an EQUAL VM opcode, so the rules of comparison are those
// of EQUAL.
func Equals(a, b any) bool {
	return atipicialgointernal.Opcode2("EQUAL", a, b).(bool)
}

// Remove removes element with index i from slice.
// This is done in place and slice must have type other than `[]byte`.
func Remove(slice any, i int) {
	atipicialgointernal.Opcode2NoReturn("REMOVE", slice, i)
}
