// 同时输出报告到子节点。
package reporter

type Reporter interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	send(string)
	Stop()
	Run()
}
