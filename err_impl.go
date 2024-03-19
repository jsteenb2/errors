package errors

import (
	"cmp"
	"fmt"
	"io"
)

func newE(opts ...any) error {
	var err e

	skipFrames := FrameSkips(3)
	for _, o := range opts {
		if o == nil {
			continue
		}
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

func (err *e) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		fallthrough
	case 's':
		io.WriteString(s, err.Error()+" ")
		err.stackTrace().Format(s, fmtInline)
	case 'q':
		fmt.Fprintf(s, "%q", err.Error())
	}
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
	for err := error(err); err != nil; err = Unwrap(err) {
		em := getErrMeta(err)
		for _, kv := range em.kvs {
			out = append(out, kv.K, kv.V)
		}
		kind = cmp.Or(kind, em.kind)
		if ej, ok := err.(*joinE); ok {
			innerKind, multiErrFields := ej.subErrFields()
			kind = cmp.Or(kind, innerKind)
			if len(multiErrFields) > 0 {
				out = append(out, "multi_err", multiErrFields)
			}
			break
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

func (err *e) Is(target error) bool {
	kind, ok := target.(Kind)
	return ok && err.kind == kind
}

func (err *e) Unwrap() error {
	return err.wrappedErr
}

func (err *e) V(key string) (any, bool) {
	for err := error(err); err != nil; err = Unwrap(err) {
		for _, kv := range getErrMeta(err).kvs {
			if kv.K == key {
				return kv.V, true
			}
		}
	}
	return nil, false
}

func (err *e) stackTrace() StackFrames {
	var out StackFrames
	for err := error(err); err != nil; err = Unwrap(err) {
		em := getErrMeta(err)
		if em.frame.FilePath == "" {
			continue
		}
		out = append(out, em.frame)
		if em.errType == errTypeJoin {
			break
		}
	}
	return out
}

type errMeta struct {
	kind    Kind
	frame   Frame
	kvs     []KV
	errType string
}

const (
	errTypeE    = "e"
	errTypeJoin = "j"
)

func getErrMeta(err error) errMeta {
	var em errMeta
	switch err := err.(type) {
	case *e:
		em.kind, em.frame, em.kvs, em.errType = err.kind, err.frame, err.kvs, errTypeE
	case *joinE:
		em.kind, em.frame, em.kvs, em.errType = err.kind, err.frame, err.kvs, errTypeJoin
	}
	return em
}

func getKind(err error) Kind {
	for ; err != nil; err = Unwrap(err) {
		if em := getErrMeta(err); em.kind != "" {
			return em.kind
		}
	}
	return ""
}
