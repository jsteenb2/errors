package errors

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"strings"
)

func newJoinE(opts ...any) error {
	var (
		baseOpts = make([]any, 1, len(opts)+1)
		errs     []error
		formatFn = listFormatFn
	)
	// since we're calling newE from 3 frames away instead of 2
	baseOpts[0] = SkipCaller

	// here we'll make use of a split loop, so that we aren't
	// polluting the newE with multi-err concerns it does not
	// need to be bothered with.
	for _, o := range opts {
		if o == nil {
			continue
		}
		switch v := o.(type) {
		case error:
			errs = append(errs, v)
		case []error:
			errs = append(errs, v...)
		case JoinFormatFn:
			if v != nil {
				formatFn = v
			}
		default:
			baseOpts = append(baseOpts, o)
		}
	}
	if len(errs) == 0 {
		return nil
	}

	ee := newE(baseOpts...).(*e)
	return &joinE{
		msg:      ee.msg,
		formatFn: formatFn,
		frame:    ee.frame,
		kind:     ee.kind,
		errs:     errs,
		kvs:      ee.kvs,
	}
}

type joinE struct {
	msg string

	formatFn JoinFormatFn
	frame    Frame
	kind     Kind
	errs     []error

	// TODO:
	//	1. should kvs be a map instead? aka unique by key name?
	//		* if unique by name... what to do with collisions, last write wins? combine values into slice?
	//		  or have some other way to signal what to do with collisions via an additional option?
	//	2. if slice of KVs, do we separate the stack frames from the output when
	//	   calling something like Meta/Fields on the error? Then have a specific
	//	   function for getting the logging fields (i.e. everything to []any)
	kvs []KV
}

func (err *joinE) Error() string {
	return err.formatFn(err.msg, err.errs)
}

func (err *joinE) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		fallthrough
	case 's':
		io.WriteString(s, err.Error())
		err.stackTrace().Format(s, fmtInline)
	case 'q':
		fmt.Fprintf(s, "%q", err.Error())
	}
}

func (err *joinE) Fields() []any {
	var (
		out  []any
		kind = err.kind
	)
	for _, kv := range err.kvs {
		out = append(out, kv.K, kv.V)
	}

	innerKind, subErrFields := err.subErrFields()
	kind = cmp.Or(kind, innerKind)
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
	for _, v := range subErrFields {
		out = append(out, v)
	}

	return out
}

func (err *joinE) subErrFields() (Kind, []any) {
	var (
		kind         Kind
		subErrFields []any
	)
	for i, err := range err.errs {
		var errFields []any
		switch err := err.(type) {
		case *e:
			errFields = err.Fields()
		case *joinE:
			errFields = err.Fields()
		}
		if len(errFields) > 0 {
			subErrFields = append(subErrFields, fmt.Sprintf("err_%d", i), errFields)
		}
		if innerKind := getKind(err); kind == "" && innerKind != "" {
			kind = innerKind
		}
	}
	return kind, subErrFields
}

func (err *joinE) stackTrace() StackFrames {
	if err.frame.FilePath == "" {
		return nil
	}
	return StackFrames{err.frame}
}

// Unwrap returns an error from Error (or nil if there are no errors).
// This error returned will further support Unwrap to get the next error,
// etc. The order will match the order of errors provided when calling Join.
//
// The resulting error supports errors.As/Is/Unwrap so you can continue
// to use the stdlib errors package to introspect further.
//
// The is borrowed from hashi/go-multierror module.
func (err *joinE) Unwrap() error {
	if err == nil || len(err.errs) == 0 {
		return nil
	}

	if len(err.errs) == 1 {
		return err.errs[0]
	}

	// Shallow copy the slice
	errs := make([]error, len(err.errs))
	copy(errs, err.errs)
	return chain(errs)
}

// chain implements the interfaces necessary for errors.Is/As/Unwrap to
// work in a deterministic way with multierror. A chain tracks a list of
// errors while accounting for the current represented error. This lets
// Is/As be meaningful.
//
// Unwrap returns the next error. In the cleanest form, Unwrap would return
// the wrapped error here but we can't do that if we want to properly
// get access to all the errors. Instead, users are recommended to use
// Is/As to get the correct error type out.
//
// Precondition: []error is non-empty (len > 0)
//
// TODO:
//   - add support for Fields
//   - add support stack trace
//   - question is, do we make these show fields/stack trace for
//     each individual error similar to how the Unwrapping is forcing
//     users to interact with the unwrapped Join error, or make it list
//     all fields/stack traces (not sure what stack trace would look like here)?
type chain []error

// Error implements the error interface
func (e chain) Error() string {
	return e[0].Error()
}

func (e chain) Fields() []any {
	fielder, ok := e[0].(interface{ Fields() []any })
	if !ok {
		return nil
	}
	return fielder.Fields()
}

func (e chain) stackTrace() StackFrames {
	st, ok := e[0].(interface{ stackTrace() StackFrames })
	if !ok {
		return nil
	}
	return st.stackTrace()
}

// Unwrap implements errors.Unwrap by returning the next error in the
// chain or nil if there are no more errors.
func (e chain) Unwrap() error {
	if len(e) == 1 {
		return nil
	}

	return e[1:]
}

// As implements errors.As by attempting to map to the current value.
func (e chain) As(target interface{}) bool {
	return errors.As(e[0], target)
}

// Is implements errors.Is by comparing the current value directly.
func (e chain) Is(target error) bool {
	return errors.Is(e[0], target)
}

// listFormatFn borrowed from hashi go-multierror module.
func listFormatFn(msg string, errs []error) string {
	if msg == "" && len(errs) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n", errs[0])
	}

	points := make([]string, len(errs))
	for i, err := range errs {
		points[i] = fmt.Sprintf("* %s", err)
	}

	if msg == "" {
		msg = fmt.Sprintf("%d errors occurred:\n\t", len(errs))
	}
	return fmt.Sprintf("%s%s\n", msg, strings.Join(points, "\n\t"))
}
