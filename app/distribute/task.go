package distribute

import ()

type Task struct {
	Id             int
	Spiders        []map[string]string //蜘蛛规则name字段与keyword字段，规定格式map[string]string{"name":"baidu","keyword":"henry"}
	OutType        string
	MaxPage        int
	ThreadNum      uint
	Pausetime      [2]uint //暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
	DockerCap      uint    //分段输出大小
	DockerQueueCap uint    //分段输出大小
}
