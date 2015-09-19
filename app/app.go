// app interface for graphical user interface.
// 基本业务执行顺序依次为：New()-->[SetLog(io.Writer)-->]Init()-->SpiderPrepare()-->Run()
package app

import (
	"io"
	// "log"
	"strconv"
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/app/crawl"
	"github.com/henrylee2cn/pholcus/app/distribute"
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/henrylee2cn/teleport"
)

type App interface {
	// 设置全局log实时显示终端
	SetLog(io.Writer) App
	// 设置全局log是否异步
	AsyncLog(enable bool) App
	// 继续log打印
	LogGoOn() App
	// 暂停log打印
	LogRest() App

	// 使用App前必须进行先Init初始化，SetLog()除外
	Init(mode int, port int, master string, w ...io.Writer) App

	// 切换运行模式并重设log打印目标
	ReInit(mode int, port int, master string, w ...io.Writer) App

	// 以下Set类方法均为Offline和Server模式用到的
	SetThreadNum(uint) App
	SetPausetime([2]uint) App //暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
	SetOutType(string) App
	SetDockerCap(uint) App
	SetMaxPage(int) App

	// SpiderPrepare()必须在设置全局运行参数之后，就Run()的前一刻执行
	// original为spider包中未有过赋值操作的原始蜘蛛种类
	// 已被显式赋值过的spider将不再重新分配Keyword
	// client模式下不调用该方法
	SpiderPrepare(original []*spider.Spider, keywords string) App

	// Run()对外为阻塞运行方式，其返回时意味着当前任务已经执行完毕
	// Run()必须在所有应当配置项配置完成后调用
	// server模式下生成任务的方法，必须在全局配置和蜘蛛队列设置完成后才可调用
	Run()

	// Offline 模式下中途终止任务
	// 对外为阻塞运行方式，其返回时意味着当前任务已经终止
	Stop()

	// Offline 模式下暂停\恢复任务
	PauseRecover()

	// 返回当前状态
	Status() int

	// 获取当前运行模式，若返回-1则表示应用未启动
	GetRunMode() int

	// 获取全部蜘蛛种类
	GetAllSpiders() []*spider.Spider

	// 通过名字获取某蜘蛛
	GetSpiderByName(string) *spider.Spider

	// 获取执行队列中蜘蛛总数
	SpiderQueueLen() int

	// 获取输出方式列表
	GetOutputLib() []string

	// 服务器客户端模式下返回节点数
	CountNodes() int
}

type Logic struct {
	// 运行模式
	RunMode int
	// 服务器端口号
	Port string
	// 服务器地址（不含Port）
	Master string
	// 全部蜘蛛种类
	spider.Traversal
	// 当前任务的蜘蛛队列
	crawl.SpiderQueue
	// 服务器与客户端间传递任务的存储库
	*distribute.TaskJar
	// 爬行回收池
	crawl.CrawlPool
	// socket长连接双工通信接口，json数据传输
	teleport.Teleport
	// 运行状态
	status       int
	finish       chan bool
	finishOnce   sync.Once
	canSocketLog bool
}

func New() App {
	return &Logic{RunMode: status.UNSET}
}

// 设置全局log实时显示终端
func (self *Logic) SetLog(w io.Writer) App {
	logs.Log.SetOutput(w)
	return self
}

// 设置全局log是否异步
func (self *Logic) AsyncLog(enable bool) App {
	logs.Log.Async(enable)
	return self
}

// 暂停log打印
func (self *Logic) LogRest() App {
	logs.Log.Rest()
	return self
}

// 继续log打印
func (self *Logic) LogGoOn() App {
	logs.Log.GoOn()
	return self
}

// 使用App前必须进行先Init初始化，SetLog()除外
func (self *Logic) Init(mode int, port int, master string, w ...io.Writer) App {
	self.canSocketLog = false
	if len(w) > 0 {
		self.SetLog(w[0])
	}
	self.LogGoOn()

	cache.Task.RunMode, cache.Task.Port, cache.Task.Master = mode, port, master

	self.Traversal = spider.Menu
	self.RunMode = mode
	self.Port = ":" + strconv.Itoa(port)
	self.Master = master
	self.Teleport = teleport.New()
	self.TaskJar = distribute.NewTaskJar()
	self.SpiderQueue = crawl.NewSpiderQueue()
	self.CrawlPool = crawl.NewCrawlPool()

	switch self.RunMode {
	case status.SERVER:
		if self.checkPort() {
			logs.Log.SetStealLevel()
			logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 服务器 ] 模式！！")
			self.Teleport.SetAPI(distribute.ServerApi(self)).Server(self.Port)
		}

	case status.CLIENT:
		if self.checkAll() {
			logs.Log.SetStealLevel()
			logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 客户端 ] 模式！！")
			self.Teleport.SetAPI(distribute.ClientApi(self)).Client(self.Master, self.Port)
		}
	case status.OFFLINE:
		logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 单机 ] 模式！！")
		return self
	default:
		logs.Log.Warning(" *    ——请指定正确的运行模式！——")
		return self
	}
	// 根据RunMode判断是否开启节点间log打印
	self.canSocketLog = true
	go self.socketLog()
	return self
}

// 切换运行模式时使用
func (self *Logic) ReInit(mode int, port int, master string, w ...io.Writer) App {
	self.LogRest()
	self.status = status.STOP
	if self.Teleport != nil {
		self.Teleport.Close()
	}
	self.CrawlPool.Stop()
	if scheduler.Sdl != nil {
		scheduler.Sdl.Stop()
	}
	self = nil
	// 等待结束
	time.Sleep(2.5e9)
	// 重新开启
	return New().Init(mode, port, master, w...)
}

func (self *Logic) SetThreadNum(threadNum uint) App {
	cache.Task.ThreadNum = threadNum
	return self
}

//暂停区间Pausetime[0]~Pausetime[0]+Pausetime[1]
func (self *Logic) SetPausetime(pause [2]uint) App {
	cache.Task.Pausetime = pause
	return self
}

func (self *Logic) SetOutType(outType string) App {
	cache.Task.OutType = outType
	return self
}

func (self *Logic) SetDockerCap(dockerCap uint) App {
	cache.Task.DockerCap = dockerCap
	cache.AutoDockerQueueCap()
	return self
}

func (self *Logic) SetMaxPage(maxPage int) App {
	cache.Task.MaxPage = maxPage
	return self
}

// SpiderPrepare()必须在设置全局运行参数之后，就Run()的前一刻执行
// original为spider包中未有过赋值操作的原始蜘蛛种类
// 已被显式赋值过的spider将不再重新分配Keyword
// client模式下不调用该方法
func (self *Logic) SpiderPrepare(original []*spider.Spider, keywords string) App {
	self.SpiderQueue.Reset()
	// 遍历任务
	for _, sp := range original {
		spgost := sp.Gost()
		spgost.SetPausetime(cache.Task.Pausetime)
		spgost.SetMaxPage(cache.Task.MaxPage)
		self.SpiderQueue.Add(spgost)
	}
	// 遍历关键词
	self.SpiderQueue.AddKeywords(keywords)
	return self
}

// 获取输出方式列表
func (self *Logic) GetOutputLib() []string {
	return collector.OutputLib
}

// 获取全部蜘蛛种类
func (self *Logic) GetAllSpiders() []*spider.Spider {
	return self.Traversal.Get()
}

// 通过名字获取某蜘蛛
func (self *Logic) GetSpiderByName(name string) *spider.Spider {
	return self.Traversal.GetByName(name)
}

// 返回当前运行模式
func (self *Logic) GetRunMode() int {
	return self.RunMode
}

// 服务器客户端模式下返回节点数
func (self *Logic) CountNodes() int {
	return self.Teleport.CountNodes()
}

// 当前蜘蛛队列长度
func (self *Logic) SpiderQueueLen() int {
	return self.SpiderQueue.Len()
}

// 运行任务
func (self *Logic) Run() {
	// 确保开启报告
	self.LogGoOn()
	self.finish = make(chan bool)
	self.finishOnce = sync.Once{}

	// 任务执行
	self.status = status.RUN
	switch cache.Task.RunMode {
	case status.OFFLINE:
		self.offline()
	case status.SERVER:
		self.server()
	case status.CLIENT:
		self.client()
	default:
		return
	}
	<-self.finish
}

// Offline 模式下暂停\恢复任务
func (self *Logic) PauseRecover() {
	switch self.status {
	case status.PAUSE:
		self.status = status.RUN
	case status.RUN:
		self.status = status.PAUSE
	}

	scheduler.Sdl.PauseRecover()
}

// Offline 模式下中途终止任务
func (self *Logic) Stop() {
	self.status = status.STOP
	self.CrawlPool.Stop()
	scheduler.Sdl.Stop()

	// 总耗时
	takeTime := time.Since(cache.StartTime).Minutes()

	// 打印总结报告
	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	logs.Log.Notice(" *                               ！！任务取消：下载页面 %v 个，耗时：%.5f 分钟！！", cache.GetPageCount(0), takeTime)
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	self.LogRest()

	// 标记结束
	self.finishOnce.Do(func() { close(self.finish) })
}

// 返回当前运行状态
func (self *Logic) Status() int {
	return self.status
}

// ******************************************** 私有方法 ************************************************* \\
// 离线模式运行
func (self *Logic) offline() {
	self.exec()
}

// 服务器模式运行，必须在SpiderPrepare()执行之后调用才可以成功添加任务
func (self *Logic) server() {
	// 标记结束
	defer func() {
		self.status = status.STOP
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	// 便利添加任务到库
	tasksNum, spidersNum := self.addNewTask()

	if tasksNum == 0 {
		return
	}

	// 打印报告
	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	logs.Log.Notice(" *                               —— 本次成功添加 %v 条任务，共包含 %v 条采集规则 ——", tasksNum, spidersNum)
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

}

// 服务器模式下，生成task并添加至库
func (self *Logic) addNewTask() (tasksNum, spidersNum int) {
	length := self.SpiderQueue.Len()
	t := distribute.Task{}

	// 从配置读取字段
	t.ThreadNum = cache.Task.ThreadNum
	t.Pausetime = cache.Task.Pausetime
	t.OutType = cache.Task.OutType
	t.DockerCap = cache.Task.DockerCap
	t.DockerQueueCap = cache.Task.DockerQueueCap
	t.MaxPage = cache.Task.MaxPage

	for i, sp := range self.SpiderQueue.GetAll() {

		t.Spiders = append(t.Spiders, map[string]string{"name": sp.GetName(), "keyword": sp.GetKeyword()})
		spidersNum++

		// 每十个蜘蛛存为一个任务
		if i > 0 && i%10 == 0 && length > 10 {
			// 存入
			one := t
			self.TaskJar.Push(&one)
			// logs.Log.Notice(" *     [新增任务]   详情： %#v", *t)

			tasksNum++

			// 清空spider
			t.Spiders = []map[string]string{}
		}
	}

	if len(t.Spiders) != 0 {
		// 存入
		one := t
		self.TaskJar.Push(&one)
		tasksNum++
	}
	return
}

// 客户端模式运行
func (self *Logic) client() {
	// 标记结束
	defer func() {
		self.status = status.STOP
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	for {
		// 从任务库获取一个任务
		t := self.downTask()

		// 准备运行
		self.taskToRun(t)

		// 执行任务
		self.exec()
	}
}

// 客户端模式下获取任务
func (self *Logic) downTask() *distribute.Task {
ReStartLabel:
	for self.CountNodes() == 0 {
		if len(self.TaskJar.Tasks) != 0 {
			break
		}
		time.Sleep(5e7)
	}

	if len(self.TaskJar.Tasks) == 0 {
		self.Request(nil, "task", "")
		for len(self.TaskJar.Tasks) == 0 {
			if self.CountNodes() == 0 {
				goto ReStartLabel
			}
			time.Sleep(5e7)
		}
	}
	return self.TaskJar.Pull()
}

// client模式下从task准备运行条件
func (self *Logic) taskToRun(t *distribute.Task) {
	// 清空历史任务
	self.SpiderQueue.Reset()

	// 更改全局配置
	cache.Task.OutType = t.OutType
	cache.Task.ThreadNum = t.ThreadNum
	cache.Task.DockerCap = t.DockerCap
	cache.Task.DockerQueueCap = t.DockerQueueCap
	cache.Task.Pausetime = t.Pausetime

	// 初始化蜘蛛队列
	for _, n := range t.Spiders {
		if sp := spider.Menu.GetByName(n["name"]); sp != nil {
			spgost := sp.Gost()
			spgost.SetPausetime(t.Pausetime)
			spgost.SetMaxPage(t.MaxPage)
			if v, ok := n["keyword"]; ok {
				spgost.SetKeyword(v)
			}
			self.SpiderQueue.Add(spgost)
		}
	}
}

// 开始执行任务
func (self *Logic) exec() {
	count := self.SpiderQueue.Len()
	cache.ReSetPageCount()

	// 初始化资源队列
	scheduler.Init(cache.Task.ThreadNum)

	// 设置爬虫队列
	crawlCap := self.CrawlPool.Reset(count)

	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	logs.Log.Informational(" *     执行任务总数（任务数[*关键词数]）为 %v 个 ...\n", count)
	logs.Log.Informational(" *     爬虫池容量为 %v ...\n", crawlCap)
	logs.Log.Informational(" *     并发协程最多 %v 个 ...\n", cache.Task.ThreadNum)
	logs.Log.Informational(" *     随机停顿时间为 %v~%v ms ...\n", cache.Task.Pausetime[0], cache.Task.Pausetime[0]+cache.Task.Pausetime[1])
	logs.Log.Informational(" * ")
	logs.Log.Notice(" *                                                                                                 —— 开始抓取，请耐心等候 ——")
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	// 开始计时
	cache.StartTime = time.Now()

	// 根据模式选择合理的并发
	if cache.Task.RunMode == status.OFFLINE {
		go self.goRun(count)
	} else {
		// 不并发是为保证接收服务端任务的同步
		self.goRun(count)
	}
}

// 任务执行
func (self *Logic) goRun(count int) {
	for i := 0; i < count && self.status != status.STOP; i++ {
		if self.status == status.PAUSE {
			time.Sleep(1e9)
			continue
		}
		// 从爬行队列取出空闲蜘蛛，并发执行
		c := self.CrawlPool.Use()
		if c != nil {
			go func(i int, c crawl.Crawler) {
				// 执行并返回结果消息
				c.Init(self.SpiderQueue.GetByIndex(i)).Start()
				// 任务结束后回收该蜘蛛
				self.CrawlPool.Free(c)
			}(i, c)
		}
	}

	// 监控结束任务
	sum := [2]uint{} //数据总数
	for i := 0; i < count; i++ {
		s := <-cache.ReportChan
		if (s.DataNum == 0) && (s.FileNum == 0) {
			continue
		}
		logs.Log.Informational(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			logs.Log.Notice(" *     [输出报告 -> 任务：%v | 关键词：%v]   共输出数据 %v 条，用时 %v 分钟！\n", s.SpiderName, s.Keyword, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			logs.Log.Notice(" *     [输出报告 -> 任务：%v | 关键词：%v]   共下载文件 %v 个，用时 %v 分钟！\n", s.SpiderName, s.Keyword, s.FileNum, s.Time)
		default:
			logs.Log.Notice(" *     [输出报告 -> 任务：%v | 关键词：%v]   共输出数据 %v 条 + 下载文件 %v 个，用时 %v 分钟！\n", s.SpiderName, s.Keyword, s.DataNum, s.FileNum, s.Time)
		}
		logs.Log.Informational(" * ")

		sum[0] += s.DataNum
		sum[1] += s.FileNum
	}

	// 总耗时
	takeTime := time.Since(cache.StartTime).Minutes()

	// 打印总结报告
	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	switch {
	case sum[0] > 0 && sum[1] == 0:
		logs.Log.Notice(" *                            —— 本次合计抓取 %v 条数据，下载页面 %v 个（成功：%v，失败：%v），耗时：%.5f 分钟 ——", sum[0], cache.GetPageCount(0), cache.GetPageCount(1), cache.GetPageCount(-1), takeTime)
	case sum[0] == 0 && sum[1] > 0:
		logs.Log.Notice(" *                            —— 本次合计抓取 %v 个文件，下载页面 %v 个（成功：%v，失败：%v），耗时：%.5f 分钟 ——", sum[1], cache.GetPageCount(0), cache.GetPageCount(1), cache.GetPageCount(-1), takeTime)
	default:
		logs.Log.Notice(" *                            —— 本次合计抓取 %v 条数据 + %v 个文件，下载网页 %v 个（成功：%v，失败：%v），耗时：%.5f 分钟 ——", sum[0], sum[1], cache.GetPageCount(0), cache.GetPageCount(1), cache.GetPageCount(-1), takeTime)
	}
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	// 单机模式并发运行，需要标记任务结束
	if cache.Task.RunMode == status.OFFLINE {
		self.status = status.STOP
		self.finishOnce.Do(func() { close(self.finish) })
	}
}

// 服务器客户端之间log打印
func (self *Logic) socketLog() {
	for {
		if !self.canSocketLog {
			return
		}
		_, msg, normal := logs.Log.StealOne()
		if !normal {
			return
		}
		if msg == "" {
			time.Sleep(5e7)
			continue
		}
		// log.Printf("截获log: %v\n", msg)
		self.Teleport.Request(msg, "log", "")
	}
}

func (self *Logic) checkPort() bool {
	if cache.Task.Port == 0 {
		logs.Log.Warning(" *     —— 亲，分布式端口不能为空哦~")
		return false
	}
	return true
}

func (self *Logic) checkAll() bool {
	if cache.Task.Master == "" || !self.checkPort() {
		logs.Log.Warning(" *     —— 亲，服务器地址不能为空哦~")
		return false
	}
	return true
}
