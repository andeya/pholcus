package common

import (
	"github.com/henrylee2cn/pholcus/logs"
	"time"
)

type DailyTimer map[string][3]int //map["src_1"][3]int{"24h","m","s"}

func (self DailyTimer) Wait(src string) {
	tdl := self.deadline(src)
	logs.Log.Critical("************************ ……当前src<%s> 下次定时更新时间为 %v ……************************", src, tdl.Format("2006-01-02 15:04:05"))
	time.Sleep(tdl.Sub(time.Now()))
}

func (self DailyTimer) deadline(src string) time.Time {
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
