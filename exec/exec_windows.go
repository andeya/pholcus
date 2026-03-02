//go:build windows

package exec

import (
	"os"
	"os/exec"
	"os/signal"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/config"

	"github.com/andeya/pholcus/cmd" // cmd UI
	"github.com/andeya/pholcus/gui" // gui UI
	"github.com/andeya/pholcus/web" // web UI
)

func run(which string) {
	_ = result.RetVoid(exec.Command("cmd.exe", "/c", "title", config.FullName).Start())

	switch which {
	case "gui":
		gui.Run()

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
