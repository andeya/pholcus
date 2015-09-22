package common

import (
	"github.com/henrylee2cn/pholcus/logs"
	"time"
)

type RSS struct {
	// RSS爬虫重新访问的５个级别（分钟）
	Level []float64
	//RSS源的权重,<len(Level),起到调整更新时间级别的规则。如当一个RSS在Level[5]，但是它的Rank是3，那么更新时间调整为Level[5-3] = Level[2] = 180分钟。
	// Rank map[string]float64
	// RSS源对应的间歇采集时间T，取整就得到Level
	T map[string]float64
	// RSS源
	Src map[string]string
	//更新标记
	Flag map[string]chan bool
}

func NewRSS(src map[string]string, level []float64) *RSS {
	rss := &RSS{
		Level: level,
		// Level: []float64{20, 60, 180, 360, 720, 1400},
		T:    make(map[string]float64),
		Src:  src,
		Flag: make(map[string]chan bool),
	}
	for k, _ := range src {
		rss.T[k] = rss.Level[0]
	}
	return rss
}

func (self *RSS) Wait(src string) {
	self.Flag[src] = make(chan bool)
	for k, v := range self.Level {
		if v < self.T[src] {
			continue
		}

		if k != 0 && v != self.T[src] {
			k--
		}
		logs.Log.Critical("************************ ……当前RSS <%s> 的更新周期为 %v 分钟……************************", src, self.Level[k])
		time.Sleep(time.Minute * time.Duration(self.Level[k]))
		break
	}
	close(self.Flag[src])
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
}

func (self *RSS) Update(src string) {
	defer func() {
		recover()
	}()
	select {
	case self.Flag[src] <- true:
	default:
		return
	}
}
