package errors_test

import (
	"fmt"
	"testing"

	"github.com/jsteenb2/errors"
)

func TestStackTrace(t *testing.T) {
	err := errors.New("some error")
	isErr(t, err)

	frames := errors.StackTrace(err)
	must(t, eqLen(t, 1, frames))

	sVal := fmt.Sprintf("%s", frames[0])
	eq(t, "frame_test.go", sVal)

	wantSPlusVal := `github.com/jsteenb2/errors_test.TestStackTrace
	github.com/jsteenb2/errors/frame_test.go`

	sPlusVal := fmt.Sprintf("%+s", frames[0])
	eq(t, wantSPlusVal, sPlusVal)

	dVal := fmt.Sprintf("%d", frames[0])
	wantLine := "11"
	eq(t, wantLine, dVal)

	vVal := fmt.Sprintf("%v", frames[0])
	wantFile := "frame_test.go:11"
	eq(t, wantFile, vVal)

	wantVPlusVal := `github.com/jsteenb2/errors_test.TestStackTrace
	github.com/jsteenb2/errors/` + wantFile
	vPlusVal := fmt.Sprintf("%+v", frames[0])
	eq(t, wantVPlusVal, vPlusVal)
}
