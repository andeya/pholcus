// 同时输出报告到子节点。
package reporter

import (
	"fmt"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"io"
	"log"
)

type Report struct {
	// 有效输出
	output io.Writer
	// 废弃输出
	rubbish io.Writer
	status  int
}

var Log Reporter

func init() {
	Log = &Report{
		rubbish: &rubbish{},
	}
}

func (self *Report) SetOutput(w io.Writer) {
	if w != nil {
		self.output = w
		log.SetOutput(w)
	}
}

func (self *Report) Run() {
	if self.output != nil {
		log.SetOutput(self.output)
	}
	self.status = status.RUN
}

func (self *Report) Stop() {
	if self.output != nil {
		log.SetOutput(self.rubbish)
	}
	self.status = status.STOP
}

func (self *Report) Printf(format string, v ...interface{}) {
	if self.status == status.STOP {
		return
	}
	log.Printf(format, v...)
	self.send(fmt.Sprintf(format, v...))
}

func (self *Report) Println(v ...interface{}) {
	if self.status == status.STOP {
		return
	}
	log.Println(v...)
	self.send(fmt.Sprintln(v...))
}

func (self *Report) Fatal(v ...interface{}) {
	if self.status == status.STOP {
		return
	}
	self.send(fmt.Sprintln(v...))
	log.Fatal(v...)
}

func (self *Report) send(str string) {
	if cache.Task.RunMode != status.OFFLINE {
		cache.PushNetData(str)
	}
}

// 数据中转
type rubbish struct{}

func (self *rubbish) Write(p []byte) (int, error) {
	return 0, nil
}
