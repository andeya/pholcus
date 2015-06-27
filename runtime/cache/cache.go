package cache

import (
	"github.com/henrylee2cn/pholcus/runtime/status"
	"time"
)

//**************************************任务运行时公共配置****************************************\\

// 任务运行时公共配置
type TaskConf struct {
	RunMode           int    // 节点角色
	Port              int    // 主节点端口
	Master            string //服务器(主节点)地址，不含端口
	ThreadNum         uint
	BaseSleeptime     uint
	RandomSleepPeriod uint //随机暂停最大增益时长
	OutType           string
	DockerCap         uint //分段转储容器容量
	DockerQueueCap    uint //分段输出池容量，不小于2
	// 选填项
	MaxPage int
}

var Task = &TaskConf{
	RunMode:           status.OFFLINE,
	Port:              2015,
	Master:            "127.0.0.1",
	ThreadNum:         20,
	BaseSleeptime:     1000,
	RandomSleepPeriod: 3000,
	OutType:           "csv",
	DockerCap:         10000,

	MaxPage: 100,
}

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
	Num        string
	Time       string
}

var (
	// 点击开始按钮的时间点
	StartTime time.Time
	// 小结报告通道
	ReportChan chan *Report
	// 请求页面计数
	ReqSum uint
)

//****************************************节点通信数据*******************************************\\

type NetData struct {
	Type int
	Body interface{}
	From string
	To   string
}

var (
	// 发送数据的缓存池(目前只在客户端存放将要发送给服务器的报告)
	SendChan = make(chan interface{}, 1024)
)

// 生成并发送信息，注意body不可为变量地址
func PushNetData(body interface{}) {
	SendChan <- body
}

//****************************************初始化*******************************************\\

func init() {
	// 任务报告
	ReportChan = make(chan *Report)

	// 根据Task.DockerCap智能调整分段输出池容量Task.DockerQueueCap
	AutoDockerQueueCap()
}
