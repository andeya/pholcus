package node

import (
	"github.com/henrylee2cn/pholcus/node/crawlpool"
	"github.com/henrylee2cn/pholcus/node/spiderqueue"
	"github.com/henrylee2cn/pholcus/node/task"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/henrylee2cn/teleport"
	"log"
	"strconv"
	"time"
)

type Node struct {
	// 运行模式
	RunMode int
	// 服务器端口号
	Port string
	// 服务器地址（不含Port）
	Master string
	// socket长连接双工通信接口，json数据传输
	teleport.Teleport
	// 节点间传递的任务的存储库
	*task.TaskJar
	// 当前任务的蜘蛛队列
	Spiders spiderqueue.SpiderQueue
	// 爬行动作的回收池
	Crawls crawlpool.CrawlPool
	// 节点状态
	Status int
}

func newPholcus() *Node {
	return &Node{
		RunMode:  cache.Task.RunMode,
		Port:     ":" + strconv.Itoa(cache.Task.Port),
		Master:   cache.Task.Master,
		Teleport: teleport.New(),
		TaskJar:  task.NewTaskJar(),
		Spiders:  spiderqueue.New(),
		Crawls:   crawlpool.New(),
		Status:   status.RUN,
	}
}

// 声明实例
var Pholcus *Node = nil

// 运行节点
func PholcusRun() {
	if Pholcus != nil {
		return
	}
	Pholcus = newPholcus()
	switch Pholcus.RunMode {
	case status.SERVER:
		if Pholcus.checkPort() {
			log.Printf("                                                                                                          ！！当前运行模式为：[ 服务器 ] 模式！！")
			Pholcus.Teleport.SetAPI(ServerApi).SetUID("服务端").Server(Pholcus.Port)
		}

	case status.CLIENT:
		if Pholcus.checkAll() {
			log.Printf("                                                                                                          ！！当前运行模式为：[ 客户端 ] 模式！！")
			Pholcus.Teleport.SetAPI(ClientApi).Client(Pholcus.Master, Pholcus.Port)
		}
	// case status.OFFLINE:
	// 	fallthrough
	default:
		log.Printf("                                                                                                          ！！当前运行模式为：[ 单机 ] 模式！！")
		return
	}
	// 开启实时log发送
	go Pholcus.log()
}

// 返回节点数
func (self *Node) CountNodes() int {
	return self.Teleport.CountNodes()
}

// 生成task并添加至库，服务器模式专用
func (self *Node) AddNewTask(spiders []string, keywords string) {
	t := &task.Task{}

	t.Spiders = spiders
	t.Keywords = keywords

	// 从配置读取字段
	t.ThreadNum = cache.Task.ThreadNum
	t.BaseSleeptime = cache.Task.BaseSleeptime
	t.RandomSleepPeriod = cache.Task.RandomSleepPeriod
	t.OutType = cache.Task.OutType
	t.DockerCap = cache.Task.DockerCap
	t.DockerQueueCap = cache.Task.DockerQueueCap
	t.MaxPage = cache.Task.MaxPage

	// 存入
	self.TaskJar.Push(t)
	// log.Printf(" *     [新增任务]   详情： %#v", *t)
}

// 客户端请求获取任务
func (self *Node) GetTaskAlways() {
	self.Request(nil, "task")
}

// 客户端模式模式下获取任务
func (self *Node) DownTask() *task.Task {
	for len(self.TaskJar.Ready) == 0 {
		if self.CountNodes() != 0 {
			self.GetTaskAlways()
			break
		}
		time.Sleep(5e7)
	}
	for len(self.TaskJar.Ready) == 0 {
		time.Sleep(5e7)
	}
	return self.TaskJar.Pull()
}

func (self *Node) log() {
	for {
		self.Teleport.Request(<-cache.SendChan, "log")
	}
}

func (self *Node) checkPort() bool {
	if cache.Task.Port == 0 {
		log.Println(" *     —— 亲，分布式端口不能为空哦~")
		return false
	}
	return true
}

func (self *Node) checkAll() bool {
	if cache.Task.Master == "" || !self.checkPort() {
		log.Println(" *     —— 亲，服务器地址不能为空哦~")
		return false
	}
	return true
}
