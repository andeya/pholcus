package cache

import (
	"github.com/henrylee2cn/pholcus/runtime/status"
	"sync"
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
	DataNum    uint
	FileNum    uint
	Time       string
}

var (
	// 点击开始按钮的时间点
	StartTime time.Time
	// 文本数据小结报告
	ReportChan chan *Report
	// 请求页面总数[]uint{总数，失败数}
	pageSum [2]uint
)

// 重置页面计数
func ReSetPageCount() {
	pageSum = [2]uint{}
}

// 0 返回总下载页数，负数 返回失败数，正数 返回成功数
func GetPageCount(i int) uint {
	if i > 0 {
		// 返回成功数
		return pageSum[0] - pageSum[1]
	}
	if i < 0 {
		// 返回失败数
		return pageSum[1]
	}
	// 返回总数
	return pageSum[0]
}

// 统计页面总下载数
var pageMutex sync.Mutex

func PageCount() {
	pageMutex.Lock()
	defer pageMutex.Unlock()
	pageSum[0]++
}

// 统计下载失败数
var pageFailMutex sync.Mutex

func PageFailCount() {
	pageFailMutex.Lock()
	defer pageFailMutex.Unlock()
	pageSum[1]++
}

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
