// 设置所有log输出位置为Log，整个项目的log不能用于非Pholcus
package web

type LogView struct {
	logChan chan string
}

func NewLogView() *LogView {
	return &LogView{logChan: make(chan string, 1024)}
}

func (self *LogView) Write(p []byte) (int, error) {
	self.logChan <- (string(p) + "\r\n")
	return len(p), nil
}

func Sprint() string {
	return <-Log.logChan
}

func Close() {
	close(Log.logChan)
}

func Open() {
	Log.logChan = make(chan string, 1024)
}

var Log = NewLogView()
