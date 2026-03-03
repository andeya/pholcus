package distribute

// TaskJar is the task storage.
type TaskJar struct {
	Tasks chan *Task
}

// NewTaskJar 创建任务存储实例。
func NewTaskJar() *TaskJar {
	return &TaskJar{
		Tasks: make(chan *Task, 1024),
	}
}

// Push adds a task to the jar (server side).
func (tj *TaskJar) Push(task *Task) {
	id := len(tj.Tasks)
	task.ID = id
	tj.Tasks <- task
}

// Pull gets a task from the local jar (client side).
func (tj *TaskJar) Pull() *Task {
	return <-tj.Tasks
}

// Len returns number of tasks in the jar.
func (tj *TaskJar) Len() int {
	return len(tj.Tasks)
}

// Send sends a task from the jar (master side).
func (tj *TaskJar) Send(clientNum int) Task {
	return *<-tj.Tasks
}

// Receive receives a task into the jar (slave side).
func (tj *TaskJar) Receive(task *Task) {
	tj.Tasks <- task
}

// CountNodes returns 0; TaskJar does not track connected nodes.
func (tj *TaskJar) CountNodes() int {
	return 0
}
