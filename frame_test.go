package errors_test

import (
	"fmt"
	"testing"

	"github.com/jsteenb2/errors"
)

func TestStackTrace_SimpleError(t *testing.T) {
	err := errors.New("some error")

	frames := errors.StackTrace(err)
	must(t, eqLen(t, 1, frames))

	eq(t, "[ github.com/jsteenb2/errors/frame_test.go:11[TestStackTrace_SimpleError] ]", frames.String())

	sVal := fmt.Sprintf("%s", frames)
	eq(t, "[ frame_test.go ]", sVal)

	sVal = fmt.Sprintf("%s", frames[0])
	eq(t, "frame_test.go", sVal)

	wantSPlusVal := `github.com/jsteenb2/errors/frame_test.go:11[TestStackTrace_SimpleError]`

	eq(t, "[ "+wantSPlusVal+" ]", fmt.Sprintf("%+s", frames))
	eq(t, wantSPlusVal, fmt.Sprintf("%+s", frames[0]))

	wantLine := "11"
	eq(t, "[ "+wantLine+" ]", fmt.Sprintf("%d", frames))
	eq(t, wantLine, fmt.Sprintf("%d", frames[0]))

	wantFile := "frame_test.go:" + wantLine
	eq(t, "[ "+wantFile+" ]", fmt.Sprintf("%v", frames))
	eq(t, wantFile, fmt.Sprintf("%v", frames[0]))

	wantVPlusVal := `github.com/jsteenb2/errors/` + wantFile + `[TestStackTrace_SimpleError]`
	eq(t, "[ "+wantVPlusVal+" ]", fmt.Sprintf("%+v", frames))
	eq(t, wantVPlusVal, fmt.Sprintf("%+v", frames[0]))
}

func TestStackTrace_WrappedError(t *testing.T) {
	err := errors.Wrap(
		errors.New("some error"),
	)

	frames := errors.StackTrace(err)
	must(t, eqLen(t, 2, frames))

	wantStr := "[ github.com/jsteenb2/errors/frame_test.go:43[TestStackTrace_WrappedError], github.com/jsteenb2/errors/frame_test.go:44[TestStackTrace_WrappedError] ]"
	eq(t, wantStr, frames.String())

	eq(t, "[ frame_test.go, frame_test.go ]", fmt.Sprintf("%s", frames))
	eq(t, "frame_test.go", fmt.Sprintf("%s", frames[0]))
	eq(t, "frame_test.go", fmt.Sprintf("%s", frames[1]))

	wantSPlusVal := `[ github.com/jsteenb2/errors/frame_test.go:43[TestStackTrace_WrappedError], github.com/jsteenb2/errors/frame_test.go:44[TestStackTrace_WrappedError] ]`
	eq(t, wantSPlusVal, fmt.Sprintf("%+s", frames))

	eq(t, `github.com/jsteenb2/errors/frame_test.go:43[TestStackTrace_WrappedError]`, fmt.Sprintf("%+s", frames[0]))
	eq(t, `github.com/jsteenb2/errors/frame_test.go:44[TestStackTrace_WrappedError]`, fmt.Sprintf("%+s", frames[1]))

	eq(t, "[ 43, 44 ]", fmt.Sprintf("%d", frames))
	eq(t, "43", fmt.Sprintf("%d", frames[0]))
	eq(t, "44", fmt.Sprintf("%d", frames[1]))

	wantFile := `[ frame_test.go:43, frame_test.go:44 ]`
	eq(t, wantFile, fmt.Sprintf("%v", frames))
	eq(t, "frame_test.go:43", fmt.Sprintf("%v", frames[0]))
	eq(t, "frame_test.go:44", fmt.Sprintf("%v", frames[1]))

	wantVPlusFrame := `[ github.com/jsteenb2/errors/frame_test.go:43[TestStackTrace_WrappedError], github.com/jsteenb2/errors/frame_test.go:44[TestStackTrace_WrappedError] ]`
	eq(t, wantVPlusFrame, fmt.Sprintf("%+v", frames))
	eq(t, "github.com/jsteenb2/errors/frame_test.go:43[TestStackTrace_WrappedError]", fmt.Sprintf("%+v", frames[0]))
	eq(t, "github.com/jsteenb2/errors/frame_test.go:44[TestStackTrace_WrappedError]", fmt.Sprintf("%+v", frames[1]))
}
