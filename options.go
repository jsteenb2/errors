package errors

import (
	"fmt"
)

// FrameSkips marks the number of frames to skip in collecting the stack frame.
// This is helpful when creating helper functions.
// TODO(berg): give example of helper functions here
type FrameSkips int

const (
	// NoFrame marks the error to not have an error frame captured. This is useful
	// when the stack frame is of no use to you the consumer.
	NoFrame FrameSkips = -1

	// SkipCaller skips the immediate caller of the functions. Useful for creating
	// reusable Error constructors in consumer code.
	SkipCaller FrameSkips = 1
)

// JoinFormatFn is the join errors formatter. This allows the user to customize
// the text output when calling Error() on the join error.
type JoinFormatFn func(msg string, errs []error) string

// Kind represents the category of the error type. A few examples of
// error kinds are as follows:
//
//	const (
//		// represents an error for a entity/thing that was not found
//		errKindNotFound = errors.Kind("not found")
//
//		//represents an error for a validation error
//		errKindInvalid = errors.Kind("invalid")
//	)
//
// With the kind, you can write common error handling across the error's
// Kind. This can create a dramatic improvement abstracting errors, allowing
// the behavior (kind) of an error to dictate semantics instead of having to
// rely on N sentinel or custom error types.
//
// Additionally, the Kind can be used to assert that an error is of a
// kind. The following uses the std lib errors.Is to determine if the
// target error is of kind "first":
//
//	err := errors.New("some error", errors.Kind("first"))
//	errors.Is(err, errors.Kind("first")) // output is true
type Kind string

// Error returns the error string indicating the kind's error. This is
// useful for working with std libs errors.Is. It requires an error type.
func (k Kind) Error() string {
	return "error kind: " + string(k)
}

// Is determines if the error's kind matches. To be used with the std
// lib errors.Is function.
func (k Kind) Is(target error) bool {
	ee, ok := target.(*e)
	return ok && ee.kind == k
}

// KV provides context to the error. These can be triggered by different
// formatter options with fmt.*printf calls of the error.
// TODO:
//  1. explore other data structures for passing in the key val pairs
type KV struct {
	K string
	V any
}

// KVs takes a slice of argument kv pairs, where the first of each pair must
// be the key string, and the latter the value. Additionally, a key that is
// a type that implements the strings.Stringer interface is also accepted.
//
// TODO:
//   - I really like the ergonomics of this when working with errors, quite
//     handy to replace the exhaustion of working with the KV type above.
//     similar approach maybe useful for other sorts of metadata as well.
func KVs(fields ...any) []KV {
	out := make([]KV, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		var kv KV
		switch t := fields[i].(type) {
		case string:
			kv.K = t
		case fmt.Stringer:
			kv.K = t.String()
		}
		if valIdx := i + 1; valIdx < len(fields) {
			kv.V = fields[valIdx]
		}
		out = append(out, kv)
	}
	return out
}
