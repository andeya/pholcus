package exec

import (
	"os"
	"os/signal"
	// "runtime"

	"github.com/henrylee2cn/pholcus/app/scheduler"

	"github.com/henrylee2cn/pholcus/cmd" // cmd版
	"github.com/henrylee2cn/pholcus/gui" // gui版
	"github.com/henrylee2cn/pholcus/web" // web版
)

func Run(which string) {
	// 开启最大核心数运行
	// runtime.GOMAXPROCS(runtime.NumCPU())

	defer func() {
		scheduler.SaveDeduplication()
	}()

	// 选择运行界面
	switch which {
	case "gui":
		gui.Run()

	case "web":
		ctrl := make(chan os.Signal, 1)
		signal.Notify(ctrl, os.Interrupt, os.Kill)
		go web.Run()
		<-ctrl

	case "cmd":
		cmd.Run()
	}
}
