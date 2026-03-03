// Package logs 提供了实时日志显示与捕获的接口封装。
package logs

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/andeya/gust/result"
	"github.com/andeya/gust/syncutil"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs/logs"
)

type (
	// Logs defines the logging interface for real-time log display and capture.
	Logs interface {
		// SetOutput sets the terminal for real-time log display.
		SetOutput(show io.Writer) Logs
		// PauseOutput pauses log output.
		PauseOutput()
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
		DelLogger(adapterName string) error
		SetLogger(adapterName string, config map[string]interface{}) error

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

var lazyLog = syncutil.NewLazyValueWithFunc(func() result.Result[Logs] {
	return result.Ok[Logs](newLogger())
})

// Log returns the lazily-initialized default logger.
// The first call triggers config loading (via config.Conf()) and logger creation.
func Log() Logs {
	return lazyLog.TryGetValue().Unwrap()
}

func newLogger() *mylog {
	p, _ := path.Split(config.LogPath)
	statR := result.Ret(os.Stat(p))
	if statR.IsErr() || !statR.Unwrap().IsDir() {
		_ = os.MkdirAll(p, 0777)
	}

	ml := &mylog{
		BeeLogger: logs.NewLogger(config.Conf().Log.Cap, config.Conf().Log.FeedbackLevel()),
	}

	ml.BeeLogger.EnableFuncCallDepth(config.Conf().Log.LineInfo)
	ml.BeeLogger.SetLevel(config.Conf().Log.Level())
	ml.BeeLogger.Async(config.LogAsync)
	ml.BeeLogger.SetLogger("console", map[string]interface{}{
		"level": config.Conf().Log.ConsoleLevel(),
	})

	if config.Conf().Log.Save {
		if r := result.RetVoid(ml.BeeLogger.SetLogger("file", map[string]interface{}{
			"filename": config.LogPath,
		})); r.IsErr() {
			fmt.Printf("Failed to create log file: %v", r.UnwrapErr())
		}
	}

	return ml
}

// SetOutput sets the terminal for real-time log display.
func (ml *mylog) SetOutput(show io.Writer) Logs {
	ml.BeeLogger.SetLogger("console", map[string]interface{}{
		"writer": show,
		"level":  config.Conf().Log.ConsoleLevel(),
	})
	return ml
}
