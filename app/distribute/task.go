package distribute

// 用于分布式分发的任务
type Task struct {
	Id             int
	Spiders        []map[string]string // 蜘蛛规则name字段与keyin字段，规定格式map[string]string{"name":"baidu","keyin":"henry"}
	ThreadNum      int                 // 全局最大并发量
	Pausetime      int64               // 暂停时长参考/ms(随机: Pausetime/2 ~ Pausetime*2)
	OutType        string              // 输出方式
	DockerCap      int                 // 分段转储容器容量
	DockerQueueCap int                 // 分段输出池容量，不小于2
	SuccessInherit bool                // 继承历史成功记录
	FailureInherit bool                // 继承历史失败记录
	Limit          int64               // 采集上限，0为不限，若在规则中设置初始值为LIMIT则为自定义限制，否则默认限制请求数
	ProxyMinute    int64               // 代理IP更换的间隔分钟数
	// 选填项
	Keyins string // 自定义输入，后期切分为多个任务的Keyin自定义配置
}
