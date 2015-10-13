// 通用对象池，动态增加对象。
package pool

import (
	"log"
	"sync"
	"time"
)

// 对象池
type Pool struct {
	Cap  int
	Src  map[Fish]bool // Fish须为指针类型
	Fish               // 对象接口
	sync.Mutex
}

// 新建一个对象池，默认容量为1024
func NewPool(fish Fish, size ...int) *Pool {
	if len(size) == 0 {
		size = append(size, 1024)
	}
	return &Pool{
		Cap:  size[0],
		Src:  make(map[Fish]bool),
		Fish: fish,
	}
}

// 对象接口
type Fish interface {
	// 返回指针类型的对象实例
	New() Fish
	// 自毁方法，在被对象池删除时调用
	Close()
	// 释放至对象池之前，清理重置自身
	Clean()
	// 判断对象是否可用
	Usable() bool
}

// 默认对象，自定义对象可包含该结构体
type Default struct{}

func (Default) New() Fish {
	log.Println("对象无效，尚未自定义 New()Fish 方法！")
	return nil
}
func (Default) Close()       {}
func (Default) Clean()       {}
func (Default) Usable() bool { return true }

// 并发安全地获取一个对象
func (self *Pool) GetOne() Fish {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()

	for {
		for k, v := range self.Src {
			if v {
				continue
			}
			if !k.Usable() {
				self.Remove(k)
				continue
			}
			self.Src[k] = true
			return k
		}
		if len(self.Src) <= self.Cap {
			self.increment()
		} else {
			time.Sleep(5e8)
		}
	}
	return nil
}

func (self *Pool) Free(m ...Fish) {
	for i, count := 0, len(m); i < count; i++ {
		m[i].Clean()
		self.Src[m[i]] = false
	}
}

// 关闭并删除指定对象
func (self *Pool) Remove(m ...Fish) {
	for _, c := range m {
		c.Close()
		delete(self.Src, c)
	}
}

// 重置对象池
func (self *Pool) Reset() {
	for k, _ := range self.Src {
		k.Close()
		delete(self.Src, k)
	}
}

// 根据情况自动动态增加对象
func (self *Pool) increment() {
	self.Src[self.Fish.New()] = false
}
