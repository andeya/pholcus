package main

import (
	"github.com/henrylee2cn/pholcus/exec"
	_ "github.com/henrylee2cn/pholcus_lib" // 此为公开维护的spider规则库
	// _ "github.com/henrylee2cn/pholcus_lib_pte" // 同样你也可以自由添加自己的规则库
)

func main() {
	// 设置运行时默认操作界面，并开始运行
	// 运行软件前，可设置 -a_ui 参数为"web"、"gui"或"cmd"，指定本次运行的操作界面
	// 其中"gui"仅支持Windows系统
	exec.DefaultRun("web")
}
