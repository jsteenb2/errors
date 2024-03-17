package errors

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// Frame is a single step in stack trace.
type Frame struct {
	FilePath string
	Fn       string
	Line     int
}

// String formats Frame to string.
func (f Frame) String() string {
	return fmt.Sprintf("%s:%d[%s]", f.FilePath, f.Line, funcname(f.Fn))
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   function name and path of source file relative to the compile time
//	      GOPATH separated by \n\t (<funcname>\n\t<path>)
//	%+v   equivalent to %+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Fn)
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.FilePath)
		default:
			io.WriteString(s, path.Base(f.FilePath))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Line))
	case 'n':
		io.WriteString(s, funcname(f.Fn))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// StackFrames represents a slice of stack Frames in LIFO order (follows path of code to get
// the original error).
// TODO:
//  1. add String method for this slice of frames so it can be used without fuss in logging
//  2. add Formatter to be able to turn off the way it prints
type StackFrames []Frame

func getFrame(skip FrameSkips) (Frame, bool) {
	if skip == NoFrame {
		return Frame{}, false
	}

	pc, path, line, ok := runtime.Caller(int(skip))
	if !ok {
		return Frame{}, false
	}

	frame := Frame{
		Fn:       runtime.FuncForPC(pc).Name(),
		Line:     line,
		FilePath: path,
	}

	return frame, true
}

// funcname removes the path prefix component of a function's name reported by func.Name().
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}
