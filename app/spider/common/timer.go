package common

import (
	"github.com/henrylee2cn/pholcus/logs" //信息输出
	"sort"
	"time"
)

// 每日定点定时器
type DailyFixedTimer map[string][3]int //map["src_1"][3]int{"24h","m","s"}

func (self DailyFixedTimer) Wait(src string) {
	tdl := self.deadline(src)
	logs.Log.Critical("************************ ……<%s> 每日定时器等待至 %v ……************************", src, tdl.Format("2006-01-02 15:04:05"))
	time.Sleep(tdl.Sub(time.Now()))
}

func (self DailyFixedTimer) deadline(src string) time.Time {
	t := time.Now()
	if t.Hour() > self[src][0] {
		t = t.Add(24 * time.Hour)
	} else if t.Hour() == self[src][0] && t.Minute() > self[src][1] {
		t = t.Add(24 * time.Hour)
	} else if t.Hour() == self[src][0] && t.Minute() == self[src][1] && t.Second() >= self[src][2] {
		t = t.Add(24 * time.Hour)
	}
	year, month, day := t.Date()
	return time.Date(year, month, day, self[src][0], self[src][1], self[src][2], 0, time.Local)
}

// 动态倒计时器
type CountdownTimer struct {
	// 倒计时的时间(min)级别，由小到大排序
	Level []float64
	// 倒计时对象的非正式计时表
	Ts map[string]float64
	//更新标记
	Flag map[string]chan bool
}

func NewCountdownTimer(level []float64, src []string) *CountdownTimer {
	if len(level) == 0 {
		level = []float64{60 * 24}
	}
	sort.Float64s(level)
	ct := &CountdownTimer{
		Level: level,
		Ts:    make(map[string]float64),
		Flag:  make(map[string]chan bool),
	}
	for _, v := range src {
		ct.Ts[v] = ct.Level[0]
	}
	return ct
}

func (self *CountdownTimer) Wait(src string) {
	self.Flag[src] = make(chan bool)
	defer func() {
		if err := recover(); err != nil {
			logs.Log.Error("动态倒计时器: %v", err)
		}
		select {
		case <-self.Flag[src]:
			self.Ts[src] = self.Ts[src] / 1.2
			if self.Ts[src] < self.Level[0] {
				self.Ts[src] = self.Level[0]
			}
		default:
			self.Ts[src] = self.Ts[src] * 1.2
			if self.Ts[src] > self.Level[len(self.Level)-1] {
				self.Ts[src] = self.Level[len(self.Level)-1]
			}
		}
	}()
	for k, v := range self.Level {
		if v < self.Ts[src] {
			continue
		}

		if k != 0 && v != self.Ts[src] {
			k--
		}
		logs.Log.Critical("************************ ……<%s> 倒计时等待 %v 分钟……************************", src, self.Level[k])
		time.Sleep(time.Duration(self.Level[k]) * time.Minute)
		break
	}
	close(self.Flag[src])
}

func (self *CountdownTimer) Update(src string) {
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
