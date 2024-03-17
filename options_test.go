package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/jsteenb2/errors"
)

func TestKind_Is(t *testing.T) {
	err := errors.New("some error", errors.Kind("first"))

	matches := stderrors.Is(errors.Kind("first"), err)
	eq(t, true, matches)
}
