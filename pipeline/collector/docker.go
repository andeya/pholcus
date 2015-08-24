// 负责从收集通道接受数据并临时存储
package collector

import (
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"sync"
	"time"
)

type DockerQueue struct {
	Curr    int
	Cap     uint
	Using   map[int]bool
	Dockers [][]DataCell
}

func NewDocker() []DataCell {
	return make([]DataCell, 0, cache.Task.DockerCap)
}

func NewDockerQueue() *DockerQueue {
	var queueCap uint = cache.Task.DockerQueueCap
	if cache.Task.DockerQueueCap < 2 {
		queueCap = 2
	}

	dockerQueue := &DockerQueue{
		Curr:    0,
		Cap:     queueCap,
		Using:   make(map[int]bool, queueCap),
		Dockers: make([][]DataCell, 0),
	}

	dockerQueue.Using[0] = true

	dockerQueue.Dockers = append(dockerQueue.Dockers, NewDocker())

	return dockerQueue
}

var ChangeMutex sync.Mutex

func (self *DockerQueue) Change() {
	ChangeMutex.Lock()
	defer ChangeMutex.Unlock()
getLable:
	for {
		for k, v := range self.Using {
			if !v {
				self.Curr = k
				self.Using[k] = true
				break getLable
			}
		}
		self.AutoAdd()
		time.Sleep(5e8)
	}
}

func (self *DockerQueue) Recover(index int) {
	self.Dockers[index] = NewDocker()
	self.Using[index] = false
}

// 根据情况自动动态增加Docker
func (self *DockerQueue) AutoAdd() {
	count := len(self.Dockers)
	if uint(count) < self.Cap {
		self.Dockers = append(self.Dockers, NewDocker())
		self.Using[count] = false
	}
}
