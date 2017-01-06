package status

// 运行模式
const (
	UNSET int = iota - 1
	OFFLINE
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
	STOPPED = iota - 1
	STOP
	RUN
	PAUSE
)
