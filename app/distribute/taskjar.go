package distribute

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

func (self *TaskJar) Out(clientNum int) Task {
	return *<-self.Tasks
}

func (self *TaskJar) Into(task *Task) {
	self.Tasks <- task
}

// 客户端从本地仓库获取一个任务
func (self *TaskJar) Pull() *Task {
	return <-self.Tasks
}

func (self *TaskJar) Len() int {
	return len(self.Tasks)
}
