package web

import (
	ws "github.com/henrylee2cn/pholcus/common/websocket"
	"github.com/henrylee2cn/pholcus/logs"
)

// log发送api
func wsLogHandle(conn *ws.Conn) {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error("%v", p)
		}
	}()
	var err error
	sess, _ := globalSessions.SessionStart(nil, conn.Request())
	sessID := sess.SessionID()
	if Lsc.connPool[sessID] == nil {
		Lsc.Add(sessID, conn)
	}
	go func() {
		defer func() {
			// 关闭web前端log输出并断开websocket连接
			Lsc.Remove(sessID, conn)
		}()
		for {
			if err := ws.JSON.Receive(conn, nil); err != nil {
				// logs.Log.Debug("websocket log接收出错断开 (%v) !", err)
				return
			}
		}
	}()

	for msg := range Lsc.lvPool[sessID].logChan {
		if _, err = ws.Message.Send(conn, msg); err != nil {
			return
		}
	}
}

type LogSocketController struct {
	connPool map[string]*ws.Conn
	lvPool   map[string]*LogView
}

func (self *LogSocketController) Write(p []byte) (int, error) {
	for sessID, lv := range self.lvPool {
		if self.connPool[sessID] != nil {
			lv.Write(p)
		}
	}
	return len(p), nil
}

func (self *LogSocketController) Add(sessID string, conn *ws.Conn) {
	self.connPool[sessID] = conn
	self.lvPool[sessID] = newLogView()
}

func (self *LogSocketController) Remove(sessID string, conn *ws.Conn) {
	defer func() {
		recover()
	}()
	if self.connPool[sessID] == nil {
		return
	}
	lv := self.lvPool[sessID]
	lv.closed = true
	close(lv.logChan)
	conn.Close()
	delete(self.connPool, sessID)
	delete(self.lvPool, sessID)
}

var Lsc = &LogSocketController{
	connPool: make(map[string]*ws.Conn),
	lvPool:   make(map[string]*LogView),
}

// 设置所有log输出位置为Log
type LogView struct {
	closed  bool
	logChan chan string
}

func newLogView() *LogView {
	return &LogView{
		logChan: make(chan string, 1024),
		closed:  false,
	}
}

func (self *LogView) Write(p []byte) (int, error) {
	if self.closed {
		goto end
	}
	defer func() { recover() }()
	self.logChan <- (string(p) + "\r\n")
end:
	return len(p), nil
}

func (self *LogView) Sprint() string {
	return <-self.logChan
}
