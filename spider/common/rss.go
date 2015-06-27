package common

import (
	"github.com/henrylee2cn/pholcus/reporter"
	"math"
	"time"
)

type RSS struct {
	// RSS爬虫重新访问的５个级别（分钟）
	Level []int
	//RSS源的权重,<len(Level),起到调整更新时间级别的规则。如当一个RSS在Level[5]，但是它的Rank是3，那么更新时间调整为Level[5-3] = Level[2] = 180分钟。
	// Rank map[string]int
	// RSS源对应的间歇采集时间T，取整就得到Level
	T map[string]int
	// RSS源
	Src map[string]string
	//更新标记
	Flag map[string]bool
}

func NewRSS(src map[string]string, level []int) *RSS {
	rss := &RSS{
		Level: level,
		// Level: []int{20, 60, 180, 360, 720, 1400},
		T:    make(map[string]int),
		Src:  src,
		Flag: make(map[string]bool),
	}
	for k, _ := range src {
		rss.T[k] = rss.Level[0]
	}
	return rss
}

func (self *RSS) Wait(src string) {
	for k, v := range self.Level {
		if v > self.T[src] {
			if k == 0 {
				k = 1
			}
			reporter.Log.Printf("************************ ……当前RSS <%s> 的更新周期为 %v 分钟……************************", src, self.Level[k-1])
			time.Sleep(time.Minute * time.Duration(self.Level[k-1]))
			break
		}
	}
	if self.Flag[src] {
		self.T[src] = int(math.Floor(float64(self.T[src]) / 1.2))
		if self.T[src] < self.Level[0] {
			self.T[src] = self.Level[0]
		}
	} else {
		self.T[src] = int(math.Floor(float64(self.T[src]) * 1.2))
		if self.T[src] > self.Level[len(self.Level)-1] {
			self.T[src] = self.Level[len(self.Level)-1]
		}
	}
	self.Flag[src] = false
}

func (self *RSS) Updata(src string) {
	self.Flag[src] = true
}
