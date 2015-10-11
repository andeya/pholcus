package exec

import (
	"os"
	"os/signal"

	"github.com/henrylee2cn/pholcus/app/scheduler"

	"github.com/henrylee2cn/pholcus/cmd" // cmd版
	"github.com/henrylee2cn/pholcus/web" // web版
)

func Run(which string) {
	defer func() {
		scheduler.SaveDeduplication()
	}()

	// 选择运行界面
	switch which {
	case "cmd":
		cmd.Run()

	case "web":
		fallthrough
	default:
		ctrl := make(chan os.Signal, 1)
		signal.Notify(ctrl, os.Interrupt, os.Kill)
		go web.Run()
		<-ctrl
	}
}
