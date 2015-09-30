package distribute

type Task struct {
	Id                   int
	Spiders              []map[string]string //蜘蛛规则name字段与keyword字段，规定格式map[string]string{"name":"baidu","keyword":"henry"}
	ThreadNum            uint
	Pausetime            [2]uint //暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
	OutType              string
	DockerCap            uint   //分段转储容器容量
	DockerQueueCap       uint   //分段输出池容量，不小于2
	InheritDeduplication bool   //继承之前的去重记录
	DeduplicationTarget  string //去重记录保存位置,"file"或"mgo"
	// 选填项
	MaxPage  int
	Keywords string //后期split()为slice
}
