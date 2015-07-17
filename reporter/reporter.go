// 同时输出报告到子节点。
// 启用顺序：SetOutput(w io.Writer)--->Run()
package reporter

import (
	"io"
)

type Reporter interface {
	// 设置全局log输出目标，不设置或设置为nil则为go语言默认
	SetOutput(w io.Writer)

	// 开启log输出
	Run()

	// 终止一切log输出
	Stop()

	// 以下打印方法除正常log输出外，若为客户端或服务端模式还将进行socket信息发送
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
}
