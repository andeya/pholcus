package cache

import (
	"sync/atomic"
	"time"
)

//**************************************任务运行时公共配置****************************************\\

// 任务运行时公共配置
type AppConf struct {
	Mode           int    // 节点角色
	Port           int    // 主节点端口
	Master         string // 服务器(主节点)地址，不含端口
	ThreadNum      int    // 全局最大并发量
	Pausetime      int64  // 暂停时长参考/ms(随机: Pausetime/2 ~ Pausetime*2)
	OutType        string // 输出方式
	DockerCap      int    // 分段转储容器容量
	DockerQueueCap int    // 分段输出池容量，不小于2
	SuccessInherit bool   // 继承历史成功记录
	FailureInherit bool   // 继承历史失败记录
	MaxPage        int64  // 最大采集页数
	ProxyMinute    int64  // 代理IP更换的间隔分钟数
	// 选填项
	Keywords string // 后期切分为slice
}

// 该初始值即默认值
var Task = new(AppConf)

// 根据Task.DockerCap智能调整分段输出池容量Task.DockerQueueCap
func AutoDockerQueueCap() {
	switch {
	case Task.DockerCap <= 10:
		Task.DockerQueueCap = 500
	case Task.DockerCap <= 500:
		Task.DockerQueueCap = 200
	case Task.DockerCap <= 1000:
		Task.DockerQueueCap = 100
	case Task.DockerCap <= 10000:
		Task.DockerQueueCap = 50
	case Task.DockerCap <= 100000:
		Task.DockerQueueCap = 10
	default:
		Task.DockerQueueCap = 4
	}
}

//****************************************任务报告*******************************************\\

type Report struct {
	SpiderName string
	Keyword    string
	DataNum    uint64
	FileNum    uint64
	Time       time.Duration
}

var (
	// 点击开始按钮的时间点
	StartTime time.Time
	// 文本数据小结报告
	ReportChan chan *Report
	// 请求页面总数[]uint{总数，失败数}
	pageSum [2]uint64
)

// 重置页面计数
func ReSetPageCount() {
	pageSum = [2]uint64{}
}

// 0 返回总下载页数，负数 返回失败数，正数 返回成功数
func GetPageCount(i int) uint64 {
	switch {
	case i > 0:
		// 返回成功数
		return pageSum[0]
	case i < 0:
		// 返回失败数
		return pageSum[1]
	case i == 0:
	}
	// 返回总数
	return pageSum[0] + pageSum[1]
}

func PageSuccCount() {
	atomic.AddUint64(&pageSum[0], 1)
}

func PageFailCount() {
	atomic.AddUint64(&pageSum[1], 1)
}

//****************************************节点通信数据*******************************************\\

// type NetData struct {
// 	Type int
// 	Body interface{}
// 	From string
// 	To   string
// }

//****************************************初始化*******************************************\\

func init() {
	// 任务报告
	ReportChan = make(chan *Report)

	// 根据Task.DockerCap智能调整分段输出池容量Task.DockerQueueCap
	AutoDockerQueueCap()
}
