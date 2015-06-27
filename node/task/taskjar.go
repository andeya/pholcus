package task

import (
// "log"
)

const (
	OWNERLESS = iota
	READY
	EXECING
	FINISH
)

type TaskJar struct {
	Tasks     map[int]*Task
	Born      []int //所有生成的Task的Id登记
	Ownerless []int //未被客户端领取的Task的Id登记
	Owned     []int //已被客户端领取的Task的Id登记
	Ready     []int //客户端等待执行的Task的Id登记
	Execing   []int //客户端正在执行的Task的Id登记
	Finish    []int //执行完成的Task的Id登记
}

func NewTaskJar() *TaskJar {
	return &TaskJar{
		Tasks:     make(map[int]*Task),
		Born:      []int{},
		Ownerless: []int{},
		Owned:     []int{},
		Ready:     []int{},
		Execing:   []int{},
		Finish:    []int{},
	}
}

// var Jar = NewTaskJar()

// // 向客户端输出一批任务[]Task
// func (self *TaskJar) Out(client string, clientNum int) []Task {
// 	last := len(self.Ownerless)
// 	if clientNum == 0 || last == 0 {
// 		return nil
// 	}

// 	if last == 1 {
// 		same := self.Ownerless
// 		self.Ownerless = nil
// 		self.Owned = append(self.Owned, same[0])
// 		self.Tasks[same[0]].Owner = client
// 		return []Task{*self.Tasks[same[0]]}
// 	}

// 	per := last / clientNum
// 	if per == 0 {
// 		per = 1
// 	}

// 	same := self.Ownerless[0:per]
// 	self.Ownerless = self.Ownerless[per:]
// 	self.Owned = append(self.Owned, same...)
// 	tasks := []Task{}
// 	for _, i := range same {
// 		self.Tasks[same[i]].Owner = client
// 		tasks = append(tasks, *self.Tasks[same[i]])
// 	}
// 	return tasks
// }

// 服务器向仓库添加一个任务
func (self *TaskJar) Push(task *Task) {
	id := len(self.Born)
	task.Id = id
	task.Status = OWNERLESS
	self.Tasks[id] = task
	self.Born = append(self.Born, id)
	self.Ownerless = append(self.Ownerless, id)
}

func (self *TaskJar) Out(client string, clientNum int) (Task, bool) {
	last := len(self.Ownerless)
	if clientNum == 0 || last == 0 {
		return Task{}, false
	}
	one := self.Ownerless[0]
	self.Ownerless = self.Ownerless[1:]
	self.Tasks[one].Owner = client
	return *self.Tasks[one], true
}

// 将从服务器获取的任务加入客户端本地仓库[]Task
// func (self *TaskJar) Into(tasks []Task) {
// 	for _, t := range tasks {
// 		t.Status = READY
// 		self.Tasks[t.Id] = &t
// 		self.joinOne(t.Id)
// 	}
// }

func (self *TaskJar) Into(task *Task) {
	task.Status = READY
	self.Tasks[task.Id] = task
	self.joinOne(task.Id)
}

// 客户端从本地仓库获取一个任务
func (self *TaskJar) Pull() *Task {
	one := self.Ready[0]
	task := self.Tasks[one]
	task.Status = EXECING
	self.Execing = append(self.Execing, one)
	self.Ready = self.Ready[1:]
	// log.Println(task)
	return task
}

// 客户端从本地仓库获取出所有任务
func (self *TaskJar) PullAll() (all []*Task) {
	count := len(self.Ready)
	for i := 0; i < count; i++ {
		all = append(all, self.Pull())
	}
	return
}

func (self *TaskJar) joinOne(id int) {
	self.Owned = append(self.Owned, id)
	self.Ready = append(self.Ready, id)
}
