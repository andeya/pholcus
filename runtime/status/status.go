package status

//**************************************表示状态的静态变量****************************************\\

// 运行模式
const (
	OFFLINE = iota
	SERVER
	CLIENT
)

// 数据头部信息
const (
	// 任务请求Header
	REQTASK = iota + 1
	// 任务响应流头Header
	TASK
	// 打印Header
	LOG
)

// 运行状态
const (
	STOP = iota
	RUN
)

//**************************************运行时状态记录****************************************\\
var (
	Crawl int
)
