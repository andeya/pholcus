package config

import (
	"github.com/henrylee2cn/pholcus/common/queue"
	"time"
)

const (
	//软件名
	APP_NAME = "幽灵蛛数据采集 V0.1 (by henrylee2cn)"
	// 蜘蛛池容量
	CRAWLER_CAP = 50

	// 收集器容量
	DATA_CAP = 2 << 14 //65536

	// mongodb数据库服务器
	DB_URL = "127.0.0.1:27017"

	//mongodb数据库名称
	DB_NAME = "temp-collection-tentinet"

	//mongodb数据库集合
	DB_COLLECTION = "news"
)

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
	// 创建默认爬行队列
	CrawlerQueue *queue.Queue

	ThreadNum uint

	OutType string

	// 分段转储容器容量
	DOCKER_CAP uint

	// 分段输出池容量，最小为2
	DOCKER_QUEUE_CAP uint
)

func init() {

	ReportChan = make(chan *Report)

	CrawlerQueue = queue.NewQueue(0)

	InitDockerParam(50000)

}

func InitDockerParam(dockercap uint) {
	DOCKER_CAP = dockercap
	switch {
	case dockercap <= 10:
		DOCKER_QUEUE_CAP = 500
	case dockercap <= 500:
		DOCKER_QUEUE_CAP = 200
	case dockercap <= 1000:
		DOCKER_QUEUE_CAP = 100
	case dockercap <= 10000:
		DOCKER_QUEUE_CAP = 50
	case dockercap <= 100000:
		DOCKER_QUEUE_CAP = 10
	default:
		DOCKER_QUEUE_CAP = 4
	}
}
