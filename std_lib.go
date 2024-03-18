package errors

import (
	"errors"
)

// As is a direct callout to the std lib errors.As function. This
// allows users to only ever have to worry about including one
// errors pkg.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is is a direct callout to the std lib errors.Is function. This
// allows users to only ever have to worry about including one
// errors pkg.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap is a direct callout to the std lib errors.Unwrap function.
// This allows users to only ever have to worry about including one
// errors pkg. The use of the std lib errors.Unwrap is a now thing,
// but may change in future releases.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
