package task

import ()

type Task struct {
	Id                int
	Spiders           []string //蜘蛛规则name字段
	OutType           string
	Keywords          string
	MaxPage           int
	ThreadNum         uint
	BaseSleeptime     uint
	RandomSleepPeriod uint //随机暂停最大增益时长
	DockerCap         uint //分段输出大小
	DockerQueueCap    uint //分段输出大小
	Status            int
	Owner             string // 客户端地址
}
