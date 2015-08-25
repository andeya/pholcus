package web

// 设置所有log输出位置为Log
type LogView struct {
	closed  bool
	logChan chan string
}

var Log = NewLogView()

func NewLogView() *LogView {
	return &LogView{logChan: make(chan string, 1024)}
}

func (self *LogView) Write(p []byte) (int, error) {
	if self.closed {
		goto end
	}
	self.logChan <- (string(p) + "\r\n")
end:
	return len(p), nil
}

func (self *LogView) Sprint() string {
	return <-self.logChan
}

func (self *LogView) Close() {
	self.closed = true
	close(self.logChan)
}

func (self *LogView) Open() {
	self.logChan = make(chan string, 1024)
	self.closed = false
}
