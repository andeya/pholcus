package collector

import (
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

// 分批输出结果的缓存队列
type DockerQueue struct {
	curr     int               //当前从通道接收数据的缓存块下标
	capacity int               //每批数据的容量
	using    map[int]bool      //正在使用的缓存块下标
	Dockers  [][]data.DataCell //缓存块列表
	lock     sync.RWMutex
}

func NewDocker() []data.DataCell {
	return make([]data.DataCell, 0, cache.Task.DockerCap)
}

func newDockerQueue(queueCap int) *DockerQueue {
	dockerQueue := &DockerQueue{
		curr:     0,
		capacity: queueCap,
		using:    make(map[int]bool, queueCap),
		Dockers:  make([][]data.DataCell, 0),
	}

	dockerQueue.using[0] = true

	dockerQueue.Dockers = append(dockerQueue.Dockers, NewDocker())

	return dockerQueue
}

func (self *DockerQueue) Curr() int {
	return self.curr
}

func (self *DockerQueue) Change() {
	for {
		self.lock.Lock()
		for k, v := range self.using {
			if !v {
				self.curr = k
				self.using[k] = true
				self.lock.Unlock()
				return
			}
		}
		if self.autoAdd() {
			self.lock.Unlock()
			continue
		}
		self.lock.Unlock()
		time.Sleep(time.Second)
		// println("::::self.DockerQueue.Change()++++++++++++++++++++++++++++++++++++++++")
	}
}

func (self *DockerQueue) Recover(index int) {
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, cell := range self.Dockers[index] {
		data.PutDataCell(cell)
	}
	self.Dockers[index] = self.Dockers[index][:0]
	self.using[index] = false
}

// 根据情况自动动态增加Docker
func (self *DockerQueue) autoAdd() bool {
	count := len(self.Dockers)
	if count < self.capacity {
		self.Dockers = append(self.Dockers, NewDocker())
		self.using[count] = false
		return true
	}
	return false
}
