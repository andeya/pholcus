package distribute

type Task struct {
	Id             int
	Spiders        []map[string]string // 蜘蛛规则name字段与keyword字段，规定格式map[string]string{"name":"baidu","keyword":"henry"}
	ThreadNum      int                 // 全局最大并发量
	Pausetime      int64               // 暂停时长参考/ms(随机: Pausetime/2 ~ Pausetime*2)
	OutType        string              // 输出方式
	DockerCap      int                 // 分段转储容器容量
	DockerQueueCap int                 // 分段输出池容量，不小于2
	SuccessInherit bool                // 继承历史成功记录
	FailureInherit bool                // 继承历史失败记录
	MaxPage        int64               // 最大采集页数
	ProxyMinute    int64               // 代理IP更换的间隔分钟数
	// 选填项
	Keywords string //后期split()为slice
}
