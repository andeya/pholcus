package main

import (
	"github.com/henrylee2cn/pholcus/command"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	command.Run()
}
