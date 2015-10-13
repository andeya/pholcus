package common

import (
	"github.com/henrylee2cn/pholcus/logs"
	"sort"
	"time"
)

type RSS struct {
	// RSS爬虫重新访问的时间(min)级别，由小到大排序
	Level []float64
	// RSS源对应的间歇采集时间T，取整就得到Level
	T map[string]float64
	// RSS源
	Src map[string]string
	//更新标记
	Flag map[string]chan bool
}

func NewRSS(src map[string]string, level []float64) *RSS {
	if len(level) == 0 {
		level = []float64{60 * 24}
	}
	sort.Float64s(level)
	rss := &RSS{
		Level: level,
		T:     make(map[string]float64),
		Src:   src,
		Flag:  make(map[string]chan bool),
	}
	for k, _ := range src {
		rss.T[k] = rss.Level[0]
	}
	return rss
}

func (self *RSS) Wait(src string) {
	self.Flag[src] = make(chan bool)
	defer func() {
		if err := recover(); err != nil {
			logs.Log.Error("rss: %v", err)
		}
		select {
		case <-self.Flag[src]:
			self.T[src] = self.T[src] / 1.2
			if self.T[src] < self.Level[0] {
				self.T[src] = self.Level[0]
			}
		default:
			self.T[src] = self.T[src] * 1.2
			if self.T[src] > self.Level[len(self.Level)-1] {
				self.T[src] = self.Level[len(self.Level)-1]
			}
		}
	}()
	for k, v := range self.Level {
		if v < self.T[src] {
			continue
		}

		if k != 0 && v != self.T[src] {
			k--
		}
		logs.Log.Critical("************************ ……当前RSS <%s> 的更新周期为 %v 分钟……************************", src, self.Level[k])
		time.Sleep(time.Duration(self.Level[k]) * time.Minute)
		break
	}
	close(self.Flag[src])
}

func (self *RSS) Update(src string) {
	go func() {
		defer func() {
			recover()
		}()
		select {
		case self.Flag[src] <- true:
		default:
			return
		}
	}()
}
