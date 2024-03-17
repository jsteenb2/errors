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
