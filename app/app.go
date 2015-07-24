// app interface for graphical user interface.
// 必须调用的三（或四）个函数方法，依次为：New()-->[SetLog(io.Writer)-->]APP.SetRunMode(int)-->APP.Run()
package app

import (
	"github.com/henrylee2cn/pholcus/crawl"
	"github.com/henrylee2cn/pholcus/crawl/pipeline/collector"
	"github.com/henrylee2cn/pholcus/crawl/scheduler"
	. "github.com/henrylee2cn/pholcus/node"
	"github.com/henrylee2cn/pholcus/node/task"
	"github.com/henrylee2cn/pholcus/reporter"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/henrylee2cn/pholcus/spider"
	_ "github.com/henrylee2cn/pholcus/spider/spiders"
	"io"
	"log"
	"sync"
	"time"
)

type App interface {
	// 获取输出方式列表
	GetOutputLib() []string

	// 获取全部蜘蛛种类
	GetAllSpiders() []*spider.Spider

	// 通过名字获取某蜘蛛
	GetSpiderByName(string) *spider.Spider

	// 获取执行队列中蜘蛛总数
	SpiderQueueLen() int

	// status.OFFLINE  status.SERVER  status.CLIENT
	// New()之后Run()之前必须调用一次该方法
	SetRunMode(int) App

	// server与client模式下设置
	SetPort(int) App

	// client模式下设置
	SetMaster(string) App

	// 以下Set类方法均为Offline和Server模式用到的
	SetThreadNum(uint) App
	SetBaseSleeptime(uint) App
	SetRandomSleepPeriod(uint) App
	SetOutType(string) App
	SetDockerCap(uint) App
	SetMaxPage(int) App

	// Offline模式下设置
	// SetSpiderQueue()必须在设置全局运行参数之后运行
	// original为spider包中未有过赋值操作的原始蜘蛛种类
	// 已被显式赋值过的spider将不再重新分配Keyword
	SetSpiderQueue(original []spider.Spider, keywords string) App

	// 设置全局log输出目标，不设置或设置为nil则为go语言默认
	SetLog(io.Writer) App

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
}

type Logic struct {
	spider.Traversal
	status     int
	finish     chan bool
	finishOnce sync.Once
}

func New() App {
	// 开启报告，调用前必须先设置Log.Output(w io.Writer)
	reporter.Log.Run()

	return &Logic{
		Traversal: spider.Menu,
	}
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

// 设置全局log输出目标，不设置或设置为nil则为go语言默认
func (self *Logic) SetLog(w io.Writer) App {
	reporter.Log.SetOutput(w)
	return self
}

var modeOnce sync.Once

func (self *Logic) SetRunMode(mode int) App {
	cache.Task.RunMode = mode
	// 运行pholcus核心
	modeOnce.Do(PholcusRun)
	return self
}

func (self *Logic) SetPort(port int) App {
	cache.Task.Port = port
	return self
}

func (self *Logic) SetMaster(master string) App {
	cache.Task.Master = master
	return self
}

func (self *Logic) SetThreadNum(threadNum uint) App {
	cache.Task.ThreadNum = threadNum
	return self
}

func (self *Logic) SetBaseSleeptime(baseSleeptime uint) App {
	cache.Task.BaseSleeptime = baseSleeptime
	return self
}

func (self *Logic) SetRandomSleepPeriod(randomSleepPeriod uint) App {
	cache.Task.RandomSleepPeriod = randomSleepPeriod
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

// SetSpiderQueue()必须在设置全局运行参数之后运行
// original为spider包中未有过赋值操作的原始蜘蛛种类
// 已被显式赋值过的spider将不再重新分配Keyword
func (self *Logic) SetSpiderQueue(original []spider.Spider, keywords string) App {
	Pholcus.Spiders.Reset()
	// 遍历任务
	for _, sp := range original {
		sp.SetPausetime(cache.Task.BaseSleeptime, cache.Task.RandomSleepPeriod)
		sp.SetMaxPage(cache.Task.MaxPage)
		me := sp
		Pholcus.Spiders.Add(&me)
	}
	// 遍历关键词
	Pholcus.Spiders.AddKeywords(keywords)
	return self
}

func (self *Logic) SpiderQueueLen() int {
	return Pholcus.Spiders.Len()
}

func (self *Logic) Run() {
	// 确保开启报告
	reporter.Log.Run()
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
		log.Println(" *    ——请指定正确的运行模式！——")
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
	Pholcus.Crawls.Stop()
	scheduler.Sdl.Stop()

	// 总耗时
	takeTime := time.Since(cache.StartTime).Minutes()

	// 打印总结报告
	log.Println(` *********************************************************************************************************************************** `)
	log.Printf(" * ")
	log.Printf(" *                               ！！任务取消：下载页面 %v 个，耗时：%.5f 分钟！！", cache.GetPageCount(0), takeTime)
	log.Printf(" * ")
	log.Println(` *********************************************************************************************************************************** `)

	reporter.Log.Stop()

	// 标记结束
	self.finishOnce.Do(func() { close(self.finish) })
}

// 返回当前运行状态
func (self *Logic) Status() int {
	return self.status
}

// ******************************************** 私有方法 ************************************************* \\

func (self *Logic) offline() {
	self.exec()
}

// 必须在SetSpiderQueue()执行之后调用才可以成功添加任务
func (self *Logic) server() {
	// 标记结束
	defer func() {
		self.status = status.STOP
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	// 便利添加任务到库
	tasksNum, spidersNum := Pholcus.AddNewTask()

	if tasksNum == 0 {
		return
	}

	// 打印报告
	log.Println(` *********************************************************************************************************************************** `)
	log.Printf(" * ")
	log.Printf(" *                               —— 本次成功添加 %v 条任务，共包含 %v 条采集规则 ——", tasksNum, spidersNum)
	log.Printf(" * ")
	log.Println(` *********************************************************************************************************************************** `)

}

func (self *Logic) client() {
	// 标记结束
	defer func() {
		self.status = status.STOP
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	for {
		// 从任务库获取一个任务
		t := Pholcus.DownTask()
		// reporter.Log.Printf("成功获取任务 %#v", t)

		// 准备运行
		self.taskToRun(t)

		// 执行任务
		self.exec()
	}
}

// client模式下从task准备运行条件
func (self *Logic) taskToRun(t *task.Task) {
	// 清空历史任务
	Pholcus.Spiders.Reset()

	// 更改全局配置
	cache.Task.OutType = t.OutType
	cache.Task.ThreadNum = t.ThreadNum
	cache.Task.DockerCap = t.DockerCap
	cache.Task.DockerQueueCap = t.DockerQueueCap

	// 初始化蜘蛛队列
	for _, n := range t.Spiders {
		if sp := spider.Menu.GetByName(n["name"]); sp != nil {
			sp.SetPausetime(t.BaseSleeptime, t.RandomSleepPeriod)
			sp.SetMaxPage(t.MaxPage)
			if v, ok := n["keyword"]; ok {
				sp.SetKeyword(v)
			}
			one := *sp
			Pholcus.Spiders.Add(&one)
		}
	}
}

// 开始执行任务
func (self *Logic) exec() {
	count := Pholcus.Spiders.Len()
	cache.ReSetPageCount()

	// 初始化资源队列
	scheduler.Init(cache.Task.ThreadNum)

	// 设置爬虫队列
	crawlNum := Pholcus.Crawls.Reset(count)

	log.Println(` *********************************************************************************************************************************** `)
	log.Printf(" * ")
	log.Printf(" *     执行任务总数（任务数[*关键词数]）为 %v 个...\n", count)
	log.Printf(" *     爬虫队列可容纳蜘蛛 %v 只...\n", crawlNum)
	log.Printf(" *     并发协程最多 %v 个……\n", cache.Task.ThreadNum)
	log.Printf(" *     随机停顿时间为 %v~%v ms ……\n", cache.Task.BaseSleeptime, cache.Task.BaseSleeptime+cache.Task.RandomSleepPeriod)
	log.Printf(" * ")
	log.Printf(" *                                                                                                 —— 开始抓取，请耐心等候 ——")
	log.Printf(" * ")
	log.Println(` *********************************************************************************************************************************** `)

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
		c := Pholcus.Crawls.Use()
		if c != nil {
			go func(i int, c crawl.Crawler) {
				// 执行并返回结果消息
				c.Init(Pholcus.Spiders.GetByIndex(i)).Start()
				// 任务结束后回收该蜘蛛
				Pholcus.Crawls.Free(c.GetId())
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
		log.Printf(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			reporter.Log.Printf(" *     [输出报告 -> 任务：%v | 关键词：%v]   共输出数据 %v 条，用时 %v 分钟！\n", s.SpiderName, s.Keyword, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			reporter.Log.Printf(" *     [输出报告 -> 任务：%v | 关键词：%v]   共下载文件 %v 个，用时 %v 分钟！\n", s.SpiderName, s.Keyword, s.FileNum, s.Time)
		default:
			reporter.Log.Printf(" *     [输出报告 -> 任务：%v | 关键词：%v]   共输出数据 %v 条 + 下载文件 %v 个，用时 %v 分钟！\n", s.SpiderName, s.Keyword, s.DataNum, s.FileNum, s.Time)
		}
		log.Printf(" * ")

		sum[0] += s.DataNum
		sum[1] += s.FileNum
	}

	// 总耗时
	takeTime := time.Since(cache.StartTime).Minutes()

	// 打印总结报告
	log.Println(` *********************************************************************************************************************************** `)
	log.Printf(" * ")
	switch {
	case sum[0] > 0 && sum[1] == 0:
		reporter.Log.Printf(" *                            —— 本次合计抓取 %v 条数据，下载页面 %v 个（成功：%v，失败：%v），耗时：%.5f 分钟 ——", sum[0], cache.GetPageCount(0), cache.GetPageCount(1), cache.GetPageCount(-1), takeTime)
	case sum[0] == 0 && sum[1] > 0:
		reporter.Log.Printf(" *                            —— 本次合计抓取 %v 个文件，下载页面 %v 个（成功：%v，失败：%v），耗时：%.5f 分钟 ——", sum[1], cache.GetPageCount(0), cache.GetPageCount(1), cache.GetPageCount(-1), takeTime)
	default:
		reporter.Log.Printf(" *                            —— 本次合计抓取 %v 条数据 + %v 个文件，下载网页 %v 个（成功：%v，失败：%v），耗时：%.5f 分钟 ——", sum[0], sum[1], cache.GetPageCount(0), cache.GetPageCount(1), cache.GetPageCount(-1), takeTime)
	}
	log.Printf(" * ")
	log.Println(` *********************************************************************************************************************************** `)

	// 单机模式并发运行，需要标记任务结束
	if cache.Task.RunMode == status.OFFLINE {
		self.status = status.STOP
		self.finishOnce.Do(func() { close(self.finish) })
	}
}
