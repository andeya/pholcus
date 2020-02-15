package goutil

import (
	"bytes"
	"runtime"
	"strconv"
	"strings"
)

// PanicTrace trace panic stack info.
func PanicTrace(kb int) []byte {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, kb<<10) //KB
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	if start == -1 {
		start = 0
	}
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

// GetCallLine gets caller line information.
func GetCallLine(calldepth int) string {
	_, file, line, ok := runtime.Caller(calldepth + 1)
	if !ok {
		return "???:0"
	}
	return file[strings.LastIndex(file, "/src/")+5:] + ":" + strconv.Itoa(line)

}
