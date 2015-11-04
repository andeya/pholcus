package common

import (
	"github.com/henrylee2cn/pholcus/logs" //信息输出
	"sort"
	"sync"
	"time"
)

// 每日定点定时器
type DailyFixedTimer map[string][3]int //map["routine_1"][3]int{"24h","m","s"}

func (self DailyFixedTimer) Wait(routine string) {
	tdl := self.deadline(routine)
	logs.Log.Critical("************************ ……<%s> 每日定时器等待至 %v ……************************", routine, tdl.Format("2006-01-02 15:04:05"))
	time.Sleep(tdl.Sub(time.Now()))
}

func (self DailyFixedTimer) deadline(routine string) time.Time {
	t := time.Now()
	if t.Hour() > self[routine][0] {
		t = t.Add(24 * time.Hour)
	} else if t.Hour() == self[routine][0] && t.Minute() > self[routine][1] {
		t = t.Add(24 * time.Hour)
	} else if t.Hour() == self[routine][0] && t.Minute() == self[routine][1] && t.Second() >= self[routine][2] {
		t = t.Add(24 * time.Hour)
	}
	year, month, day := t.Date()
	return time.Date(year, month, day, self[routine][0], self[routine][1], self[routine][2], 0, time.Local)
}

// 动态倒计时器
type CountdownTimer struct {
	// 倒计时的时间(min)级别，由小到大排序
	Level []float64
	// 倒计时对象的非正式计时表
	Routines map[string]*routineTime
	//更新标记
	Flag map[string]chan bool
	sync.Mutex
}

type routineTime struct {
	Min  float64
	Curr float64
}

// 参数routines为 map[string]float64{倒计时对象UID: 最小等待的参考时间}
func NewCountdownTimer(level []float64, routines map[string]float64) *CountdownTimer {
	if len(level) == 0 {
		level = []float64{60 * 24}
	}
	sort.Float64s(level)
	ct := &CountdownTimer{
		Level:    level,
		Routines: make(map[string]*routineTime),
		Flag:     make(map[string]chan bool),
	}
	for routine, minTime := range routines {
		ct.Routines[routine] = &routineTime{
			Curr: ct.Level[0],
			Min:  minTime,
		}
	}
	return ct
}

// 需在执行Update()的协程执行之后调用
func (self *CountdownTimer) Wait(routine string) {
	// 保存指针
	if _, ok := self.Routines[routine]; !ok {
		return
	}
	rt := self.Routines[routine]

	self.Flag[routine] = make(chan bool)
	defer func() {
		if err := recover(); err != nil {
			logs.Log.Error("动态倒计时器: %v", err)
		}
		select {
		case <-self.Flag[routine]:
			n := rt.Curr / 1.2
			if n > rt.Min {
				rt.Curr = n
			} else {
				// 等待时间不能小于设定时间
				rt.Curr = rt.Min
			}

			if rt.Curr < self.Level[0] {
				// 等待时间不能小于最小水平
				rt.Curr = self.Level[0]
			}
		default:
			rt.Curr = rt.Curr * 1.2
			if rt.Curr > self.Level[len(self.Level)-1] {
				// 等待时间不能大于最大水平
				rt.Curr = self.Level[len(self.Level)-1]
			}
		}
	}()
	for k, v := range self.Level {
		if v < rt.Curr {
			continue
		}

		if k != 0 && v != rt.Curr {
			k--
		}
		logs.Log.Critical("************************ ……<%s> 倒计时等待 %v 分钟……************************", routine, self.Level[k])
		time.Sleep(time.Duration(self.Level[k]) * time.Minute)
		break
	}
	close(self.Flag[routine])
}

// 需在Wait()方法执行之前，在新的协程调用
func (self *CountdownTimer) Update(routine string) {
	defer func() {
		recover()
	}()

	if _, ok := self.Routines[routine]; !ok {
		return
	}

	select {
	case self.Flag[routine] <- true:
	default:
		return
	}
}

func (self *CountdownTimer) SetRoutine(routine string, minTime float64) *CountdownTimer {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()
	self.Routines[routine] = &routineTime{
		Curr: self.Level[0],
		Min:  minTime,
	}
	return self
}

func (self *CountdownTimer) RemoveRoutine(routine string) *CountdownTimer {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()
	delete(self.Routines, routine)
	delete(self.Flag, routine)
	return self
}

func (self *CountdownTimer) SetLevel(level []float64) *CountdownTimer {
	self.Mutex.Lock()
	defer self.Mutex.Unlock()
	self.Level = level
	return self
}
