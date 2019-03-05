package goutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// DEFAULT_PID_FILE the default PID file name
var DEFAULT_PID_FILE = "log/PID"

// WritePidFile writes the current PID to the specified file.
func WritePidFile(pidFile ...string) {
	fname := DEFAULT_PID_FILE
	if len(pidFile) > 0 {
		fname = pidFile[0]
	}
	abs, err := filepath.Abs(fname)
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(abs)
	os.MkdirAll(dir, 0777)
	pid := os.Getpid()
	f, err := os.OpenFile(abs, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%d\n", pid))
}
