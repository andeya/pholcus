package exec

import (
	"flag"
	"runtime"

	"github.com/henrylee2cn/pholcus/cmd"
	"github.com/henrylee2cn/pholcus/web"
)

func init() {
	// 开启最大核心数运行
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func DefaultRun(uiDefault string) {
	flag.String("a . . . . . . . . . . . . .. . . . . . . . . . . common . . . . . . . . . . . . . .. . . . . . . . . . a", "common", "\r\n")
	ui := flag.String("a_ui", uiDefault, "   <选择操作界面> {web} {gui} {cmd}\r\n\r\n")
	web.Flag()
	cmd.Flag()
	flag.String("z", "readme", "   参数设置参考 {value} 提示，参数中包含多个值时以 \",\" 间隔\r\n")
	flag.Parse()
	run(*ui)
}
