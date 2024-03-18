package errors

// New creates a new error.
func New(msg string, opts ...any) error {
	passedOpts := make([]any, 1, len(opts)+1)
	passedOpts[0] = msg
	passedOpts = append(passedOpts, opts...)
	return newE(passedOpts...)
}

// Wrap wraps the provided error and includes any additional options on this
// entry of the error. Note, a msg is not required. A new stack frame will
// be captured when calling Wrap. It is useful for that alone. This function
// will not wrap a nil error, rather, it'll return with a nil.
func Wrap(err error, opts ...any) error {
	if err == nil {
		return nil
	}

	passedOpts := make([]any, 1, len(opts)+1)
	passedOpts[0] = err
	passedOpts = append(passedOpts, opts...)
	return newE(passedOpts...)
}

// Fields returns logging fields for a given error.
func Fields(err error) []any {
	if err == nil {
		return nil
	}

	fielder, ok := err.(interface{ Fields() []any })
	if !ok {
		return nil
	}

	return fielder.Fields()
}

// StackTrace returns the StackFrames for an error. See StackFrames for more info.
// TODO:
//  1. make this more robust with Is
//  2. determine if its even worth exposing an accessor for this private method
func StackTrace(err error) StackFrames {
	ee, ok := err.(*e)
	if !ok {
		return nil
	}
	return ee.stackTrace()
}

// V returns a typed value for the kvs of an error. Type conversion
// can be used to convert the output value. We do not distinguish
// between a purposeful <nil> value and key not found. With the
// single return param, we can do the following to convert it to a
// more specific type:
//
//	err := errors.New("simple msg", errors.KVs("int", 1))
//	i, ok := errors.V(err, "int).(int)
//
// Note: this will take the first matching key. If you are interested
// in obtaining a key's value from a wrapped error collides with a
// parent's key value, then you can manually unwrap the error and call V
// on it to skip the parent field.
//
// TODO:
//   - food for thought, we could change the V funcs signature
//     to allow for a generic type to be provided, however... it
//     feels both premature and limiting in the event you don't
//     care about the type. If we get an ask for that, we can provide
//     guidance for this via the comment above and perhaps some example
//     code.
func V(err error, key string) any {
	if err == nil {
		return nil
	}

	fielder, ok := err.(interface{ V(key string) (any, bool) })
	if !ok {
		return nil
	}

	raw, _ := fielder.V(key)
	return raw
}
