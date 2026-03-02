package logs

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs/logs"
)

type (
	// Logs defines the logging interface for real-time log display and capture.
	Logs interface {
		// SetOutput sets the terminal for real-time log display.
		SetOutput(show io.Writer) Logs
		// Rest pauses log output.
		Rest()
		// GoOn resumes from pause and continues log output.
		GoOn()
		// EnableStealOne enables or disables log capture copy mode.
		EnableStealOne(bool)
		// StealOne captures log copies in order, returning one at a time; normal indicates whether the logger is closed.
		StealOne() (level int, msg string, normal bool)
		// Close shuts down log output normally.
		Close()
		// Status returns the running status, e.g. 0,"RUN".
		Status() (int, string)
		DelLogger(adaptername string) error
		SetLogger(adaptername string, config map[string]interface{}) error

		// The following methods output logs and, in client/server mode, also send messages over the socket.
		Debug(format string, v ...interface{})
		Informational(format string, v ...interface{})
		App(format string, v ...interface{})
		Notice(format string, v ...interface{})
		Warning(format string, v ...interface{})
		Error(format string, v ...interface{})
		Critical(format string, v ...interface{})
		Alert(format string, v ...interface{})
		Emergency(format string, v ...interface{})
	}
	mylog struct {
		*logs.BeeLogger
	}
)

// Log is the default logger instance.
var Log = func() Logs {
	p, _ := path.Split(config.LOG)
	// Create directory when it does not exist
	statR := result.Ret(os.Stat(p))
	if statR.IsErr() || !statR.Unwrap().IsDir() {
		result.RetVoid(os.MkdirAll(p, 0777)).Unwrap()
	}

	ml := &mylog{
		BeeLogger: logs.NewLogger(config.LOG_CAP, config.LOG_FEEDBACK_LEVEL),
	}

	ml.BeeLogger.EnableFuncCallDepth(config.LOG_LINEINFO)
	ml.BeeLogger.SetLevel(config.LOG_LEVEL)
	ml.BeeLogger.Async(config.LOG_ASYNC)
	ml.BeeLogger.SetLogger("console", map[string]interface{}{
		"level": config.LOG_CONSOLE_LEVEL,
	})

	if config.LOG_SAVE {
		if r := result.RetVoid(ml.BeeLogger.SetLogger("file", map[string]interface{}{
			"filename": config.LOG,
		})); r.IsErr() {
			fmt.Printf("Failed to create log file: %v", r.UnwrapErr())
		}
	}

	return ml
}()

// SetOutput sets the terminal for real-time log display.
func (self *mylog) SetOutput(show io.Writer) Logs {
	self.BeeLogger.SetLogger("console", map[string]interface{}{
		"writer": show,
		"level":  config.LOG_CONSOLE_LEVEL,
	})
	return self
}
