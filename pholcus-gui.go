package main

import (
	"github.com/henrylee2cn/pholcus/gui"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	gui.Run()
}
