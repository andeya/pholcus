package collector

import (
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

// 分批输出结果的缓存队列
type DockerQueue struct {
	Curr    int               //当前从通道接收数据的缓存块下标
	Cap     int               //每批数据的容量
	Using   map[int]bool      //正在使用的缓存块下标
	Dockers [][]data.DataCell //缓存块列表
}

var changeMutex sync.Mutex

func NewDocker() []data.DataCell {
	return make([]data.DataCell, 0, cache.Task.DockerCap)
}

func NewDockerQueue() *DockerQueue {
	var queueCap = cache.Task.DockerQueueCap
	if cache.Task.DockerQueueCap < 2 {
		queueCap = 2
	}

	dockerQueue := &DockerQueue{
		Curr:    0,
		Cap:     queueCap,
		Using:   make(map[int]bool, queueCap),
		Dockers: make([][]data.DataCell, 0),
	}

	dockerQueue.Using[0] = true

	dockerQueue.Dockers = append(dockerQueue.Dockers, NewDocker())

	return dockerQueue
}

func (self *DockerQueue) Change() {
	changeMutex.Lock()
	defer changeMutex.Unlock()
	for {
		for k, v := range self.Using {
			if !v {
				self.Curr = k
				self.Using[k] = true
				return
			}
		}
		self.AutoAdd()
		time.Sleep(100 * time.Millisecond)
	}
}

func (self *DockerQueue) Recover(index int) {
	for _, cell := range self.Dockers[index] {
		data.PutDataCell(cell)
	}
	self.Dockers[index] = self.Dockers[index][:0]
	self.Using[index] = false
}

// 根据情况自动动态增加Docker
func (self *DockerQueue) AutoAdd() {
	count := len(self.Dockers)
	if count < self.Cap {
		self.Dockers = append(self.Dockers, NewDocker())
		self.Using[count] = false
	}
}
