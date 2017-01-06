package distribute

// 任务仓库
type TaskJar struct {
	Tasks chan *Task
}

func NewTaskJar() *TaskJar {
	return &TaskJar{
		Tasks: make(chan *Task, 1024),
	}
}

// 服务器向仓库添加一个任务
func (self *TaskJar) Push(task *Task) {
	id := len(self.Tasks)
	task.Id = id
	self.Tasks <- task
}

// 客户端从本地仓库获取一个任务
func (self *TaskJar) Pull() *Task {
	return <-self.Tasks
}

// 仓库任务总数
func (self *TaskJar) Len() int {
	return len(self.Tasks)
}

// 主节点从仓库发送一个任务
func (self *TaskJar) Send(clientNum int) Task {
	return *<-self.Tasks
}

// 从节点接收一个任务到仓库
func (self *TaskJar) Receive(task *Task) {
	self.Tasks <- task
}
