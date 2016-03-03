package scheduler

import (
	"runtime"
	"sync"

	"github.com/henrylee2cn/pholcus/app/aid/history"
	"github.com/henrylee2cn/pholcus/app/aid/proxy"
	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
)

type scheduler struct {
	// Spider实例的请求矩阵列表
	matrices map[int]*Matrix
	// 总并发量计数
	count chan bool
	// 运行状态
	status int
	// 全局代理IP
	proxy *proxy.Proxy
	// 标记是否使用代理IP
	useProxy bool
	// 全局历史记录
	history history.Historier
	// 全局读写锁
	sync.RWMutex
}

// 定义全局调度
var sdl = &scheduler{
	history: history.New(),
	status:  status.RUN,
	count:   make(chan bool, cache.Task.ThreadNum),
	proxy:   proxy.New(),
}

func Init() {
	for sdl.proxy == nil {
		runtime.Gosched()
	}
	sdl.matrices = make(map[int]*Matrix)
	sdl.count = make(chan bool, cache.Task.ThreadNum)
	if cache.Task.Mode == status.OFFLINE {
		sdl.history.ReadSuccess(cache.Task.OutType, cache.Task.SuccessInherit)
		sdl.history.ReadFailure(cache.Task.OutType, cache.Task.FailureInherit)
	}
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
	sdl.Lock()
	defer func() {
		recover()
		sdl.Unlock()
	}()

	sdl.status = status.STOP
	for _, v := range sdl.matrices {
		for _, vv := range v.reqs {
			for _, req := range vv {
				// 删除队列中未执行请求的去重记录
				DeleteSuccess(req)
			}
		}
		v.reqs = make(map[int][]*request.Request)
		v.priorities = []int{}
	}

	// 清空
	close(sdl.count)
	sdl.matrices = make(map[int]*Matrix)
}

func UpsertSuccess(record history.Record) bool {
	return sdl.history.UpsertSuccess(record)
}

func DeleteSuccess(record history.Record) {
	sdl.history.DeleteSuccess(record)
}

func UpsertFailure(req *request.Request) bool {
	return sdl.history.UpsertFailure(req)
}

func DeleteFailure(req *request.Request) {
	sdl.history.DeleteFailure(req)
}

// 获取指定蜘蛛在上一次运行时失败的请求
func PullFailure(spiderName string) []*request.Request {
	return sdl.history.PullFailure(spiderName)
}

func TryFlushHistory() {
	if cache.Task.SuccessInherit {
		sdl.history.FlushSuccess(cache.Task.OutType)
	}
	if cache.Task.FailureInherit {
		sdl.history.FlushFailure(cache.Task.OutType)
	}
}

// 每个spider实例分配到的平均资源量
func (self *scheduler) avgRes() int32 {
	avg := int32(cap(sdl.count) / len(sdl.matrices))
	if avg == 0 {
		avg = 1
	}
	return avg
}
