// 通用资源池，动态增加资源。
package pool

import (
	"sync"
	"time"
)

// 资源接口
type Src interface {
	// 返回指针类型的资源实例
	New() Src
	// 自毁方法，在被资源池删除时调用
	Close()
	// 释放至资源池之前，清理重置自身
	Clean()
	// 判断资源是否已过期
	Expired() bool
}

// 资源池
type Pool struct {
	Src                    // 资源接口
	srcMap   map[Src]bool  // Src须为指针类型
	capacity int           // 资源池容量
	tryTimes int           // 创建新资源时，重试次数，0为默认10次，小于0为无限次
	gctime   time.Duration // 回收监测间隔
	sync.Mutex
}

// 新建一个资源池
func NewPool(src Src, size int, tryTimes int, gctime ...time.Duration) *Pool {
	if len(gctime) == 0 {
		gctime = append(gctime, 60e9)
	}
	if tryTimes <= 0 {
		tryTimes = 10
	}
	pool := &Pool{
		Src:      src,
		srcMap:   make(map[Src]bool),
		capacity: size,
		gctime:   gctime[0],
		tryTimes: tryTimes,
	}
	go pool.gc()

	return pool
}

// 并发安全地获取一个资源
func (self *Pool) GetOne() Src {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()

	for i := 0; i < self.tryTimes; i++ {
		for k, v := range self.srcMap {
			if v {
				continue
			}
			if k.Expired() {
				self.Remove(k)
				continue
			}
			self.use(k)
			return k
		}
		if len(self.srcMap) <= self.capacity {
			self.increment()
		} else {
			time.Sleep(5e8)
		}
	}
	return nil
}

func (self *Pool) Free(m ...Src) {
	defer func() {
		recover()
	}()
	for i, count := 0, len(m); i < count; i++ {
		m[i].Clean()
		self.srcMap[m[i]] = false
	}
}

// 关闭并删除指定资源
func (self *Pool) Remove(m ...Src) {
	defer func() {
		recover()
	}()
	for _, c := range m {
		c.Close()
		delete(self.srcMap, c)
	}
}

// 重置资源池
func (self *Pool) Reset() {
	defer func() {
		recover()
	}()
	for k, _ := range self.srcMap {
		k.Close()
		delete(self.srcMap, k)
	}
}

// 根据情况自动动态增加资源
func (self *Pool) increment() {
	src := self.Src.New()
	if src == nil {
		return
	}
	self.srcMap[src] = false
}

func (self *Pool) use(m Src) {
	self.srcMap[m] = true
}

// 空闲资源回收
func (self *Pool) gc() {
	defer func() {
		recover()
	}()
	for {
		self.Mutex.Lock()
		for k, v := range self.srcMap {
			if !v {
				self.Remove(k)
			}
		}
		self.Mutex.Unlock()
		time.Sleep(self.gctime)
	}
}
