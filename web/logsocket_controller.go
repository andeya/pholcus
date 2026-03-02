package web

import (
	"regexp"
	"runtime/debug"

	"github.com/andeya/gust/syncutil"
	ws "github.com/andeya/pholcus/common/websocket"
	"github.com/andeya/pholcus/logs"
)

// wsLogHandle handles WebSocket connections for streaming logs to the client.
func wsLogHandle(conn *ws.Conn) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
		}
	}()
	r := globalSessions.SessionStart(nil, conn.Request())
	if r.IsErr() {
		logs.Log().Error("session start: %v", r.UnwrapErr())
		return
	}
	sess := r.Unwrap()
	sessID := sess.SessionID()
	if LogSocketCtrl.connPool.Load(sessID).IsNone() {
		LogSocketCtrl.Add(sessID, conn)
	}
	defer func() {
		LogSocketCtrl.Remove(sessID)
	}()
	for {
		if err := ws.JSON.Receive(conn, nil); err != nil {
			return
		}
	}
}

// LogSocketController manages WebSocket connections for log streaming.
type LogSocketController struct {
	connPool syncutil.SyncMap[string, *ws.Conn]
}

var (
	// LogSocketCtrl is the global LogSocketController for log streaming.
	LogSocketCtrl = new(LogSocketController)
	colorRegexp   = regexp.MustCompile("\033\\[[0-9;]{1,4}m")
)

func (lsc *LogSocketController) Write(p []byte) (int, error) {
	defer func() {
		if r := recover(); r != nil {
			logs.Log().Error("panic recovered: %v\n%s", r, debug.Stack())
		}
	}()
	p = colorRegexp.ReplaceAll(p, []byte{})
	lsc.connPool.Range(func(sessID string, conn *ws.Conn) bool {
		if _, err := ws.Message.Send(conn, (string(p) + "\r\n")); err != nil {
			lsc.Remove(sessID)
		}
		return true
	})
	return len(p), nil
}

func (lsc *LogSocketController) Add(sessID string, conn *ws.Conn) {
	lsc.connPool.Store(sessID, conn)
}

func (lsc *LogSocketController) Remove(sessID string) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error("panic recovered: %v\n%s", p, debug.Stack())
		}
	}()
	connOpt := lsc.connPool.LoadAndDelete(sessID)
	if connOpt.IsSome() {
		connOpt.Unwrap().Close()
	}
}
