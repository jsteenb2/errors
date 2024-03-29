package errors

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

const fmtInline = 'i'

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
			io.WriteString(s, f.String())
		default:
			io.WriteString(s, path.Base(f.FilePath))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Line))
	case 'n':
		io.WriteString(s, funcname(f.Fn))
	case fmtInline:
		io.WriteString(s, f.String())
	case 'v':
		if s.Flag('+') {
			f.Format(s, 's')
			return
		}
		io.WriteString(s, path.Base(f.FilePath))
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

// String formats Frame to string.
func (f StackFrames) String() string {
	var sb strings.Builder
	sb.WriteString("[ ")

	for i, frame := range f {
		sb.WriteString(frame.String())
		if i < len(f)-1 {
			sb.WriteString(", ")
		}
	}
	if sb.Len() > 2 {
		sb.WriteString(" ")
	}
	sb.WriteString("]")
	return sb.String()
}

// Format formats the frame according to the fmt.Formatter interface.
// See Frame.Format for the formatting rules.
func (f StackFrames) Format(s fmt.State, verb rune) {
	io.WriteString(s, "[ ")
	defer func() { io.WriteString(s, " ]") }()
	for i, frame := range f {
		frame.Format(s, verb)
		if i < len(f)-1 {
			io.WriteString(s, ", ")
		}
	}
}

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
