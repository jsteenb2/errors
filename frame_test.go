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
	eq(t, "frame_test.go", sVal)

	sVal = fmt.Sprintf("%s", frames[0])
	eq(t, "frame_test.go", sVal)

	wantSPlusVal := `github.com/jsteenb2/errors_test.TestStackTrace_SimpleError
	github.com/jsteenb2/errors/frame_test.go`

	sPlusVal := fmt.Sprintf("%+s", frames)
	eq(t, wantSPlusVal, sPlusVal)

	sPlusVal = fmt.Sprintf("%+s", frames[0])
	eq(t, wantSPlusVal, sPlusVal)

	wantLine := "11"
	dVal := fmt.Sprintf("%d", frames)
	eq(t, wantLine, dVal)
	dVal = fmt.Sprintf("%d", frames[0])
	eq(t, wantLine, dVal)

	wantFile := "frame_test.go:" + wantLine
	vVal := fmt.Sprintf("%v", frames)
	eq(t, wantFile, vVal)
	vVal = fmt.Sprintf("%v", frames[0])
	eq(t, wantFile, vVal)

	wantVPlusVal := `github.com/jsteenb2/errors_test.TestStackTrace_SimpleError
	github.com/jsteenb2/errors/` + wantFile
	vPlusVal := fmt.Sprintf("%+v", frames)
	eq(t, wantVPlusVal, vPlusVal)
	vPlusVal = fmt.Sprintf("%+v", frames[0])
	eq(t, wantVPlusVal, vPlusVal)
}

func TestStackTrace_WrappedError(t *testing.T) {
	err := errors.Wrap(
		errors.New("some error"),
	)

	frames := errors.StackTrace(err)
	must(t, eqLen(t, 2, frames))

	wantStr := "[ github.com/jsteenb2/errors/frame_test.go:54[TestStackTrace_WrappedError], github.com/jsteenb2/errors/frame_test.go:55[TestStackTrace_WrappedError] ]"
	eq(t, wantStr, frames.String())

	eq(t, "frame_test.go\n\nframe_test.go", fmt.Sprintf("%s", frames))
	eq(t, "frame_test.go", fmt.Sprintf("%s", frames[0]))
	eq(t, "frame_test.go", fmt.Sprintf("%s", frames[1]))

	wantSPlusVal := `github.com/jsteenb2/errors_test.TestStackTrace_WrappedError
	github.com/jsteenb2/errors/frame_test.go

github.com/jsteenb2/errors_test.TestStackTrace_WrappedError
	github.com/jsteenb2/errors/frame_test.go`
	eq(t, wantSPlusVal, fmt.Sprintf("%+s", frames))

	wantSPlusVal = `github.com/jsteenb2/errors_test.TestStackTrace_WrappedError
	github.com/jsteenb2/errors/frame_test.go`
	eq(t, wantSPlusVal, fmt.Sprintf("%+s", frames[0]))
	eq(t, wantSPlusVal, fmt.Sprintf("%+s", frames[1]))

	eq(t, "54\n\n55", fmt.Sprintf("%d", frames))
	eq(t, "54", fmt.Sprintf("%d", frames[0]))
	eq(t, "55", fmt.Sprintf("%d", frames[1]))

	wantFile := `frame_test.go:54

frame_test.go:55`
	eq(t, wantFile, fmt.Sprintf("%v", frames))
	eq(t, "frame_test.go:54", fmt.Sprintf("%v", frames[0]))
	eq(t, "frame_test.go:55", fmt.Sprintf("%v", frames[1]))

	wantVPlusFrame := `github.com/jsteenb2/errors_test.TestStackTrace_WrappedError
	github.com/jsteenb2/errors/frame_test.go`
	eq(t, wantVPlusFrame+":54\n\n"+wantVPlusFrame+":55", fmt.Sprintf("%+v", frames))
	eq(t, wantVPlusFrame+":54", fmt.Sprintf("%+v", frames[0]))
	eq(t, wantVPlusFrame+":55", fmt.Sprintf("%+v", frames[1]))
}
