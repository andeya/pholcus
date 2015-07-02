package task

import ()

type Task struct {
	Id                int
	Spiders           []map[string]string //蜘蛛规则name字段与keyword字段，规定格式map[string]string{"name":"baidu","keyword":"henry"}
	OutType           string
	MaxPage           int
	ThreadNum         uint
	BaseSleeptime     uint
	RandomSleepPeriod uint //随机暂停最大增益时长
	DockerCap         uint //分段输出大小
	DockerQueueCap    uint //分段输出大小
}
