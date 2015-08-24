package distribute

type subApp interface {
	// 将任务加入仓库
	Into(task *Task)
	// 从仓库取出任务
	Out(clientNum int) Task
	// 返回节点数
	CountNodes() int
}
