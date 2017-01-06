package distribute

// 分布式的接口
type Distributer interface {
	// 主节点从仓库发送一个任务
	Send(clientNum int) Task
	// 从节点接收一个任务到仓库
	Receive(task *Task)
	// 返回与之连接的节点数
	CountNodes() int
}
