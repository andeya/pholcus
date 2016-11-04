package scheduler

import (
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/app/aid/proxy"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

// 调度器
type scheduler struct {
	status       int          // 运行状态
	count        chan bool    // 总并发量计数
	useProxy     bool         // 标记是否使用代理IP
	proxy        *proxy.Proxy // 全局代理IP
	matrices     []*Matrix    // Spider实例的请求矩阵列表
	sync.RWMutex              // 全局读写锁
}

// 定义全局调度
var sdl = &scheduler{
	status: status.RUN,
	count:  make(chan bool, cache.Task.ThreadNum),
	proxy:  proxy.New(),
}

func Init() {
	for sdl.proxy == nil {
		time.Sleep(100 * time.Millisecond)
	}
	sdl.matrices = []*Matrix{}
	sdl.count = make(chan bool, cache.Task.ThreadNum)

	if cache.Task.ProxyMinute > 0 {
		if sdl.proxy.Count() > 0 {
			sdl.useProxy = true
			sdl.proxy.UpdateTicker(cache.Task.ProxyMinute)
			logs.Log.Informational(" *     使用代理IP，代理IP更换频率为 %v 分钟\n", cache.Task.ProxyMinute)
		} else {
			sdl.useProxy = false
			logs.Log.Informational(" *     在线代理IP列表为空，无法使用代理IP\n")
		}
	} else {
		sdl.useProxy = false
		logs.Log.Informational(" *     不使用代理IP\n")
	}

	sdl.status = status.RUN
}

// 注册资源队列
func AddMatrix(spiderName, spiderSubName string, maxPage int64) *Matrix {
	matrix := newMatrix(spiderName, spiderSubName, maxPage)
	sdl.RLock()
	defer sdl.RUnlock()
	sdl.matrices = append(sdl.matrices, matrix)
	return matrix
}

// 暂停\恢复所有爬行任务
func PauseRecover() {
	sdl.Lock()
	defer sdl.Unlock()
	switch sdl.status {
	case status.PAUSE:
		sdl.status = status.RUN
	case status.RUN:
		sdl.status = status.PAUSE
	}
}

// 终止任务
func Stop() {
	// println("scheduler^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	sdl.Lock()
	defer sdl.Unlock()
	sdl.status = status.STOP
	// 清空
	defer func() {
		recover()
	}()
	// for _, matrix := range sdl.matrices {
	// 	matrix.windup()
	// }
	close(sdl.count)
	sdl.matrices = []*Matrix{}
	// println("scheduler$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
}

// 每个spider实例分配到的平均资源量
func (self *scheduler) avgRes() int32 {
	avg := int32(cap(sdl.count) / len(sdl.matrices))
	if avg == 0 {
		avg = 1
	}
	return avg
}

func (self *scheduler) checkStatus(s int) bool {
	self.RLock()
	b := self.status == s
	self.RUnlock()
	return b
}
