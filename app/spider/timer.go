package spider

import (
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/logs" //信息输出
)

type Timer struct {
	setting map[string]*Clock
	sync.RWMutex
	closed bool
}

func newTimer() *Timer {
	return &Timer{
		setting: make(map[string]*Clock),
	}
}

// 休眠等待，并返回定时器是否可以继续使用
func (self *Timer) sleep(id string) bool {
	self.RLock()
	c, ok := self.setting[id]
	self.RUnlock()
	if ok {
		c.sleep()
		self.RLock()
		_, ok = self.setting[id]
		self.RUnlock()
	}
	return ok
}

func (self *Timer) set(id string, tol time.Duration, t0 *T0) bool {
	self.Lock()
	defer self.Unlock()
	if self.closed {
		return false
	}
	c, ok := newClock(id, tol, t0)
	if !ok {
		return ok
	}
	self.setting[id] = c
	return ok
}

func (self *Timer) drop() {
	self.Lock()
	defer self.Unlock()
	self.closed = true
	for _, c := range self.setting {
		c.wake()
	}
	self.setting = make(map[string]*Clock)
}

const (
	// 闹钟
	A = iota
	// 倒计时
	T
)

type Clock struct {
	id string
	// 闹铃or倒计时
	typ int
	// 计时公差
	tol time.Duration
	// 起始时间
	t0    *T0
	timer *time.Timer
}

type T0 struct {
	Hour int
	Min  int
	Sec  int
}

func newClock(id string, tol time.Duration, t0 *T0) (*Clock, bool) {
	if t0 == nil {
		return &Clock{
			id:    id,
			typ:   T,
			tol:   tol,
			timer: time.NewTimer(0),
		}, true
	}
	if !(t0.Hour >= 0 && t0.Hour < 24 && t0.Min >= 0 && t0.Min < 60 && t0.Sec >= 0 && t0.Sec < 60) {
		return nil, false
	}
	return &Clock{
		id:    id,
		typ:   A,
		tol:   tol,
		t0:    t0,
		timer: time.NewTimer(0),
	}, true
}

func (self *Clock) sleep() {
	d := self.duration()
	logs.Log.Critical("************************ ……<%s> 定时器等待 %v ，计划 %v 恢复 ……************************", self.id, d, time.Now().Add(d).Format("2006-01-02 15:04:05"))
	self.timer.Reset(d)
	<-self.timer.C
}

func (self *Clock) wake() {
	self.timer.Reset(0)
}

func (self *Clock) duration() time.Duration {
	switch self.typ {
	case A:
		t := time.Now()
		year, month, day := t.Date()
		t0 := time.Date(year, month, day, self.t0.Hour, self.t0.Min, self.t0.Sec, 0, time.Local)
		t0.Add(self.tol)
		if t0.Before(t) {
			t0.Add(time.Hour * 24)
		}
		return t0.Sub(t)
	case T:
		return self.tol
	}
	return 0
}
