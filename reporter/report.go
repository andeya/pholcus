package reporter

import (
	"fmt"
	"log"
)

const (
	STOP = 0
	RUN  = 1
)

type Report struct {
	status int
}

func (self *Report) send(str string) {
	if true {

	}
}

func (self *Report) Printf(format string, v ...interface{}) {
	if self.status == STOP {
		return
	}
	log.Printf(format, v...)
	self.send(fmt.Sprintf(format, v...))
}

func (self *Report) Println(v ...interface{}) {
	if self.status == STOP {
		return
	}
	log.Println(v...)
	self.send(fmt.Sprintln(v...))
}

func (self *Report) Stop() {
	self.status = STOP
}

func (self *Report) Run() {
	self.status = RUN
}

var Log Reporter

func init() {
	Log = &Report{}
}
