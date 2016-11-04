package models

import (
	"fmt"
	"github.com/henrylee2cn/teleport"
	"log"
	"sync"
	"time"
)

// 仅请求一次并返回数据
func RequestOnce(body interface{}, operation string, flag string, nodeuid ...string) (interface{}, bool) {
	m := hubConns.getOne()
	m.Teleport.Request(body, operation, flag, nodeuid...)
	r := <-m.result
	hubConns.free(m)
	return r[0], r[1].(bool)
}

// 获取一个管理中心连接
func GetManage() *Manage {
	return hubConns.getOne()
}

// 请求并返回数据，须配合 GetManage()与FreeManage(）一起使用
func (self *Manage) Request(body interface{}, operation string, flag string, nodeuid ...string) (interface{}, bool) {
	self.Teleport.Request(body, operation, flag, nodeuid...)
	r := <-self.result
	return r[0], r[1].(bool)
}

// 释放管理中心连接
func FreeManage(m ...*Manage) {
	hubConns.free(m...)
}

// 关闭并删除指定连接
func RemoveManage(m ...*Manage) {
	hubConns.remove(m...)
}

// 重置连接池
func ResetManage() {
	hubConns.reset()
}

type Manage struct {
	teleport.Teleport
	result chan [2]interface{}
}

const (
	// 功能模块名称
	M_WEB = "data580" // data580网站
	// port of socket server
	MANANGE_SOCKET_PORT = ":8000"
	MANANGE_SOCKET_IP   = "127.0.0.1"
)

// 新建连接
func newManage() *Manage {
	m := &Manage{
		Teleport: teleport.New(),
		result:   make(chan [2]interface{}, 1),
	}
	uid := M_WEB + ":" + fmt.Sprint(time.Now().Unix())
	m.SetAPI(newManageApi(m)).SetUID(uid).Client(MANANGE_SOCKET_IP, MANANGE_SOCKET_PORT)
	return m
}

// 管理中心连接池
type hubPool struct {
	Cap int
	Src map[*Manage]bool
	sync.Mutex
}

var hubConns = &hubPool{
	Cap: 1024,
	Src: make(map[*Manage]bool),
}

// 并发安全地获取一个连接
func (self *hubPool) getOne() *Manage {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()

	for {
		for k, v := range self.Src {
			if v {
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

func (self *hubPool) free(m ...*Manage) {
	for i, count := 0, len(m); i < count; i++ {
		m[i].result = make(chan [2]interface{}, 1)
		self.Src[m[i]] = false
	}
}

// 关闭并删除指定连接
func (self *hubPool) remove(m ...*Manage) {
	for _, c := range m {
		c.Close()
		delete(self.Src, c)
	}
}

// 重置连接池
func (self *hubPool) reset() {
	for k, _ := range self.Src {
		k.Close()
		delete(self.Src, k)
	}
}

// 根据情况自动动态增加连接
func (self *hubPool) increment() {
	if len(self.Src) < self.Cap {
		self.Src[newManage()] = false
	}
}

func newManageApi(m *Manage) teleport.API {
	return teleport.API{
		// 获取数据
		"get": &result{m},
		// 获取table的清单
		"list": &result{m},
	}
}

// 获取返回的结果
type result struct {
	*Manage
}

func (self *result) Process(receive *teleport.NetData) *teleport.NetData {
	if receive.Status != teleport.SUCCESS {
		log.Printf("error: %v，%v", receive.Body, receive.Status)
		self.result <- [2]interface{}{receive.Body, false}
		return nil
	}
	self.result <- [2]interface{}{receive.Body, true}
	return nil
}
