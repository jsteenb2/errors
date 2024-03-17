package errors

import (
	"errors"
)

func newE(opts ...any) error {
	var err e

	skipFrames := FrameSkips(3)
	for _, o := range opts {
		switch arg := o.(type) {
		case string:
			err.msg = arg
		case FrameSkips:
			if arg == NoFrame {
				skipFrames = NoFrame
			} else if skipFrames != NoFrame {
				skipFrames += +arg
			}
		case Kind:
			err.kind = arg
		case KV:
			err.kvs = append(err.kvs, arg)
		case []KV:
			err.kvs = append(err.kvs, arg...)
		case error:
			err.wrappedErr = arg
		}
	}
	if frame, ok := getFrame(skipFrames); ok {
		err.frame = frame
	}
	return &err
}

// e represents the internal error. We do not expose this error to avoid
// hyrum's law. We wish for folks to sink their teeth into the behavior
// of this error, via the exported funcs, or w/e interface a consumer
// wishes to create.
// TODO:
//  1. add Formatter implementation
type e struct {
	msg string

	frame      Frame
	kind       Kind
	wrappedErr error

	// TODO:
	//	1. should kvs be a map instead? aka unique by key name?
	//		* if unique by name... what to do with collisions, last write wins? combine values into slice?
	//		  or have some other way to signal what to do with collisions via an additional option?
	//	2. if slice of KVs, do we separate the stack frames from the output when
	//	   calling something like Meta/Fields on the error? Then have a specific
	//	   function for getting the logging fields (i.e. everything to []any)
	kvs []KV
}

func (err *e) Error() string {
	msg := err.msg
	if err.wrappedErr != nil {
		if msg != "" {
			msg += ": "
		}
		msg += err.wrappedErr.Error()
	}
	return msg
}

// Fields represents the meaningful logging fields from the error. These include
// all the KVs, error kind, and stack trace for the error and all wrapped error
// fields. This is an incredibly powerful tool to enhance the logging/observability
// for errors.
//
// TODO:
//   - decide how to handle duplicate kvs, for the time being, allow duplicates
//   - decide on name for this method, Fields is mostly referring to the fields
//     that are useful in the context of logging, where contextual metadata from
//     the error can eliminate large swathes of the DEBUG/Log driven debugging.
func (err *e) Fields() []any {
	var (
		out  []any
		kind Kind
	)
	for err := error(err); err != nil; err = errors.Unwrap(err) {
		ee, ok := err.(*e)
		if !ok {
			continue
		}

		for _, kv := range ee.kvs {
			out = append(out, kv.K, kv.V)
		}
		if kind == "" {
			kind = ee.kind
		}
	}
	if kind != "" {
		out = append(out, "err_kind", string(kind))
	}
	if stackFrames := err.stackTrace(); len(stackFrames) > 0 {
		var simplified []string
		for _, frame := range stackFrames {
			simplified = append(simplified, frame.String())
		}
		out = append(out, "stack_trace", simplified)
	}

	return out
}

func (err *e) Unwrap() error {
	return err.wrappedErr
}

func (err *e) stackTrace() StackFrames {
	var out StackFrames
	for err := error(err); err != nil; err = errors.Unwrap(err) {
		if ee, ok := err.(*e); ok && ee.frame.FilePath != "" {
			out = append(out, ee.frame)
		}
	}
	return out
}
