package proxy

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/henrylee2cn/ping"

	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

const TIMEOUT = 4 //4s

type Proxy struct {
	usable       map[string]bool
	speed        []string
	timedelay    []time.Duration
	curProxy     string
	curTimedelay time.Duration
	ticker       *time.Ticker
	tickMinute   int64
	sync.Once
}

func New() *Proxy {
	return (&Proxy{
		usable: map[string]bool{},
	}).Update()
}

// 代理IP数量
func (self *Proxy) Count() int {
	return len(self.usable)
}

// 更新代理IP列表
func (self *Proxy) Update() *Proxy {
	f, err := os.Open(config.PROXY_FILE)
	if err != nil {
		// logs.Log.Error("Error: %v\n", err)
		return self
	}
	b, _ := ioutil.ReadAll(f)
	s := strings.Replace(string(b), " ", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n\n", "\n", -1)

	for i, proxy := range strings.Split(s, "\n") {
		self.usable[proxy] = true
		fmt.Printf("+ 代理IP %v：%v\n", i, proxy)
	}

	return self
}

// 更新继时器
func (self *Proxy) UpdateTicker(tickMinute int64) {
	if self.tickMinute == tickMinute {
		return
	}
	self.tickMinute = tickMinute
	self.ticker = time.NewTicker(time.Duration(self.tickMinute) * time.Minute)
	self.Once = sync.Once{}
}

// 获取本次循环中未使用的代理IP及其响应时长
func (self *Proxy) GetOne() (string, time.Duration) {
	if len(self.usable) == 0 {
		return "", -1
	}
	self.updateSort()
	select {
	case <-self.ticker.C:
		self.curProxy = self.speed[1]
		self.curTimedelay = self.timedelay[1]
		self.speed = self.speed[1:]
		self.timedelay = self.timedelay[1:]
		logs.Log.Informational(" *     设置代理IP为 [%v](%v)\n", self.curProxy, self.curTimedelay)
	default:
		self.Once.Do(func() {
			self.curProxy = self.speed[1]
			self.curTimedelay = self.timedelay[1]
			self.speed = self.speed[1:]
			self.timedelay = self.timedelay[1:]
			logs.Log.Informational(" *     设置代理IP为 [%v](%v)\n", self.curProxy, self.curTimedelay)
		})
	}
	return self.curProxy, self.curTimedelay
}

// 为代理IP测试并排序
func (self *Proxy) updateSort() *Proxy {
	if len(self.speed) == 0 {
		for proxy, _ := range self.usable {
			self.usable[proxy] = true
		}
	}
	self.speed = []string{}
	self.timedelay = []time.Duration{}

	for proxy, unused := range self.usable {
		if unused {
			alive, err, timedelay := ping.Ping(proxy, TIMEOUT)
			self.speed = append(self.speed, proxy)
			if !alive || err != nil {
				self.timedelay = append(self.timedelay, TIMEOUT+1)
			} else {
				self.timedelay = append(self.timedelay, timedelay)
			}
		}
	}

	sort.Sort(self)

	return self
}

// 实现排序接口
func (self *Proxy) Len() int {
	return len(self.speed)
}
func (self *Proxy) Less(i, j int) bool {
	return self.timedelay[i] < self.timedelay[j]
}
func (self *Proxy) Swap(i, j int) {
	self.speed[i], self.speed[j] = self.speed[j], self.speed[i]
	self.timedelay[i], self.timedelay[j] = self.timedelay[j], self.timedelay[i]
}
