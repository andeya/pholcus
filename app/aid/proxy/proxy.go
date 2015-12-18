package proxy

import (
	// "fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/henrylee2cn/ping"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

const (
	// ping最大时长
	TIMEOUT = 4 //4s
	// 尝试ping的最大次数
	PING_TIMES = 3
	// IP测速的最大并发量
	MAX_THREAD_NUM = 1000
)

type Proxy struct {
	ipRegexp     *regexp.Regexp
	ipMap        map[string]string
	usable       map[string]bool
	speed        []string
	timedelay    []time.Duration
	curProxy     string
	curTimedelay time.Duration
	ticker       *time.Ticker
	tickMinute   int64
	pingPool     chan bool
	sync.Once
}

func New() *Proxy {
	return (&Proxy{
		ipRegexp: regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`),
		ipMap:    map[string]string{},
		usable:   map[string]bool{},
		pingPool: make(chan bool, MAX_THREAD_NUM),
	}).Update()
}

// 代理IP数量
func (self *Proxy) Count() int {
	return len(self.usable)
}

// 更新代理IP列表
func (self *Proxy) Update() *Proxy {
	go func() {
		once.Do(mkdir)

		f, err := os.Open(config.PROXY_FULL_FILE_NAME)
		if err != nil {
			// logs.Log.Error("Error: %v\n", err)
			return
		}
		b, _ := ioutil.ReadAll(f)
		s := strings.Replace(string(b), " ", "", -1)
		s = strings.Replace(s, "\r", "", -1)
		s = strings.Replace(s, "\n\n", "\n", -1)
		s = strings.Trim(s, "\n")

		for _, proxy := range strings.Split(s, "\n") {
			self.ipMap[proxy] = self.ipRegexp.FindString(proxy)
			self.usable[proxy] = true
			// fmt.Printf("+ 代理IP %v：%v\n", i, proxy)
		}
		logs.Log.Informational(" *     读取代理IP: %v 条\n", len(self.usable))
	}()
	return self
}

// 更新继时器
func (self *Proxy) UpdateTicker(tickMinute int64) {
	self.tickMinute = tickMinute
	self.ticker = time.NewTicker(time.Duration(self.tickMinute) * time.Minute)
	self.Once = sync.Once{}
}

// 获取本次循环中未使用的代理IP及其响应时长
func (self *Proxy) GetOne() (string, time.Duration) {
	if len(self.usable) == 0 {
		return "", -1
	}
	select {
	case <-self.ticker.C:
		self.getOne()
	default:
		self.Once.Do(self.getOne)
	}
	// fmt.Printf("获取使用IP：[%v](%v)\n", self.curProxy, self.curTimedelay)
	return self.curProxy, self.curTimedelay
}

func (self *Proxy) getOne() {
	self.updateSort()
	if len(self.speed) == 0 {
		self.curProxy, self.curTimedelay = "", 0
		logs.Log.Informational(" *     设置代理IP失败，没有可用的代理IP\n")
		return
	}
	// fmt.Printf("使用前IP测试%#v\n", self.timedelay)
	self.curProxy = self.speed[0]
	self.curTimedelay = self.timedelay[0]
	self.speed = self.speed[1:]
	self.timedelay = self.timedelay[1:]
	self.usable[self.curProxy] = false
	logs.Log.Informational(" *     设置代理IP为 [%v](%v)\n", self.curProxy, self.curTimedelay)
	// fmt.Printf("当前IP情况%#v\n", self.usable)
	// fmt.Printf("当前未用IP%#v\n", self.speed)
}

// 为代理IP测试并排序
func (self *Proxy) updateSort() *Proxy {
	logs.Log.Informational(" *     正在测速代理IP并排序……")
	if len(self.speed) == 0 {
		for proxy, _ := range self.usable {
			self.usable[proxy] = true
		}
	}

	// 最多尝试ping PING_TIMES次
	for i := PING_TIMES; i > 0; i-- {
		self.speed = []string{}
		self.timedelay = []time.Duration{}
		for proxy, unused := range self.usable {
			if unused {
				self.ping(proxy)
			}
		}
		for len(self.pingPool) > 0 {
			runtime.Gosched()
		}
		if len(self.speed) > 0 {
			sort.Sort(self)
			break
		}
		for proxy, _ := range self.usable {
			self.usable[proxy] = true
		}
	}
	logs.Log.Informational(" *     代理IP测速与排序完成，有效IP：%v 个\n", len(self.speed))

	return self
}

func (self *Proxy) ping(proxy string) {
	self.pingPool <- true
	go func() {
		alive, err, timedelay := ping.Ping(self.ipMap[proxy], TIMEOUT)
		// fmt.Printf("ping：%v, %v, %v, %v\n", self.ipMap[proxy], alive, err, timedelay)
		if !alive || err != nil {
			// 跳过无法ping通的ip
			self.usable[proxy] = false
		} else {
			self.speed = append(self.speed, proxy)
			self.timedelay = append(self.timedelay, timedelay)
		}
		<-self.pingPool
	}()
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

var once = new(sync.Once)

func mkdir() {
	util.Mkdir(config.PROXY_FULL_FILE_NAME)
}
