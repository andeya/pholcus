package reporter

import (
	"fmt"
	"log"
)

type Report struct{}

func (self *Report) send(str string) {
	if true {

	}
}

func (self *Report) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
	self.send(fmt.Sprintf(format, v...))
}

func (self *Report) Println(v ...interface{}) {
	log.Println(v...)
	self.send(fmt.Sprintln(v...))
}

var Log Reporter

func init() {
	Log = &Report{}
}
