// app interface for graphical user interface.
// 基本业务执行顺序依次为：New()-->[SetLog(io.Writer)-->]Init()-->SpiderPrepare()-->Run()
package app

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
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

// 全局唯一的核心接口实例
var LogicApp = New()

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

	// 获取全局参数
	GetAppConf(k ...string) interface{}

	// 设置全局参数，Offline和Server模式用到的
	SetAppConf(k string, v interface{}) App

	// SpiderPrepare()必须在设置全局运行参数之后，就Run()的前一刻执行
	// original为spider包中未有过赋值操作的原始蜘蛛种类
	// 已被显式赋值过的spider将不再重新分配Keyword
	// client模式下不调用该方法
	SpiderPrepare(original []*spider.Spider) App

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

	// 获取全部蜘蛛种类
	GetSpiderLib() []*spider.Spider

	// 通过名字获取某蜘蛛
	GetSpiderByName(string) *spider.Spider

	// 获取蜘蛛队列接口实例
	GetSpiderQueue() crawl.SpiderQueue

	// 获取全部输出方式
	GetOutputLib() []string

	// 服务器客户端模式下返回节点数
	CountNodes() int
}

type Logic struct {
	// 全局配置
	*cache.AppConf
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
	//执行计数
	sum [2]uint64
	// 执行计时
	takeTime time.Duration
	// 运行状态
	status       int
	finish       chan bool
	finishOnce   sync.Once
	canSocketLog bool
	sync.RWMutex
}

/*
 * 任务运行时公共配置
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
*/

func New() App {
	return newLogic()
}

func newLogic() *Logic {
	return &Logic{
		AppConf:     cache.Task,
		Traversal:   spider.Menu,
		status:      status.STOP,
		Teleport:    teleport.New(),
		TaskJar:     distribute.NewTaskJar(),
		SpiderQueue: crawl.NewSpiderQueue(),
		CrawlPool:   crawl.NewCrawlPool(),
	}
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

// 获取全局参数
func (self *Logic) GetAppConf(k ...string) interface{} {
	defer func() {
		if err := recover(); err != nil {
			logs.Log.Error(fmt.Sprintf("%v", err))
		}
	}()
	if len(k) == 0 {
		return self.AppConf
	}
	key := strings.Title(k[0])
	acv := reflect.ValueOf(self.AppConf).Elem()
	return acv.FieldByName(key).Interface()
}

// 设置全局参数
func (self *Logic) SetAppConf(k string, v interface{}) App {
	defer func() {
		if err := recover(); err != nil {
			logs.Log.Error(fmt.Sprintf("%v", err))
		}
	}()
	if k == "MaxPage" && v.(int64) <= 0 {
		v = int64(spider.MAXPAGE)
	}
	acv := reflect.ValueOf(self.AppConf).Elem()
	key := strings.Title(k)
	if acv.FieldByName(key).CanSet() {
		acv.FieldByName(key).Set(reflect.ValueOf(v))
	}
	return self
}

// 使用App前必须进行先Init初始化，SetLog()除外
func (self *Logic) Init(mode int, port int, master string, w ...io.Writer) App {
	self.canSocketLog = false
	if len(w) > 0 {
		self.SetLog(w[0])
	}
	self.LogGoOn()

	self.AppConf.Mode, self.AppConf.Port, self.AppConf.Master = mode, port, master
	self.Teleport = teleport.New()
	self.TaskJar = distribute.NewTaskJar()
	self.SpiderQueue = crawl.NewSpiderQueue()
	self.CrawlPool = crawl.NewCrawlPool()

	switch self.AppConf.Mode {
	case status.SERVER:
		if self.checkPort() {
			logs.Log.SetStealLevel()
			logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 服务器 ] 模式！！")
			self.Teleport.SetAPI(distribute.ServerApi(self)).Server(":" + strconv.Itoa(self.AppConf.Port))
		}

	case status.CLIENT:
		if self.checkAll() {
			logs.Log.SetStealLevel()
			logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 客户端 ] 模式！！")
			self.Teleport.SetAPI(distribute.ClientApi(self)).Client(self.AppConf.Master, ":"+strconv.Itoa(self.AppConf.Port))
		}
	case status.OFFLINE:
		logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 单机 ] 模式！！")
		return self
	default:
		logs.Log.Warning(" *    ——请指定正确的运行模式！——")
		return self
	}
	// 根据Mode判断是否开启节点间log打印
	self.canSocketLog = true
	go self.socketLog()
	return self
}

// 切换运行模式时使用
func (self *Logic) ReInit(mode int, port int, master string, w ...io.Writer) App {
	scheduler.Stop()
	self.LogRest()
	self.setStatus(status.STOP)
	if self.Teleport != nil {
		self.Teleport.Close()
	}
	self.CrawlPool.Stop()
	// 等待结束
	if mode == status.UNSET {
		self = newLogic()
		self.AppConf.Mode = status.UNSET
		return self
	}
	// 重新开启
	self = newLogic().Init(mode, port, master, w...).(*Logic)
	return self
}

// SpiderPrepare()必须在设置全局运行参数之后，Run()的前一刻执行
// original为spider包中未有过赋值操作的原始蜘蛛种类
// 已被显式赋值过的spider将不再重新分配Keyword
// client模式下不调用该方法
func (self *Logic) SpiderPrepare(original []*spider.Spider) App {
	self.SpiderQueue.Reset()
	// 遍历任务
	for _, sp := range original {
		spcopy := sp.Copy()
		spcopy.SetPausetime(self.AppConf.Pausetime)
		if spcopy.GetMaxPage() > 0 {
			spcopy.SetMaxPage(self.AppConf.MaxPage)
		} else {
			spcopy.SetMaxPage(-1 * self.AppConf.MaxPage)
		}
		self.SpiderQueue.Add(spcopy)
	}
	// 遍历关键词
	self.SpiderQueue.AddKeywords(self.AppConf.Keywords)
	return self
}

// 获取全部输出方式
func (self *Logic) GetOutputLib() []string {
	return collector.OutputLib
}

// 获取全部蜘蛛种类
func (self *Logic) GetSpiderLib() []*spider.Spider {
	return self.Traversal.Get()
}

// 通过名字获取某蜘蛛
func (self *Logic) GetSpiderByName(name string) *spider.Spider {
	return self.Traversal.GetByName(name)
}

// 返回当前运行模式
func (self *Logic) GetMode() int {
	return self.AppConf.Mode
}

// 服务器客户端模式下返回节点数
func (self *Logic) CountNodes() int {
	return self.Teleport.CountNodes()
}

// 获取蜘蛛队列接口实例
func (self *Logic) GetSpiderQueue() crawl.SpiderQueue {
	return self.SpiderQueue
}

// 运行任务
func (self *Logic) Run() {
	// 确保开启报告
	self.LogGoOn()
	self.finish = make(chan bool)
	self.finishOnce = sync.Once{}
	// 重置计数
	self.sum[0], self.sum[1] = 0, 0
	// 重置计时
	self.takeTime = 0
	// 任务执行
	self.setStatus(status.RUN)
	switch self.AppConf.Mode {
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
	scheduler.TryFlushHistory()
}

// Offline 模式下暂停\恢复任务
func (self *Logic) PauseRecover() {
	switch self.Status() {
	case status.PAUSE:
		self.setStatus(status.RUN)
	case status.RUN:
		self.setStatus(status.PAUSE)
	}

	scheduler.PauseRecover()
}

// Offline 模式下中途终止任务
func (self *Logic) Stop() {
	self.setStatus(status.STOP)
	self.CrawlPool.Stop()
	scheduler.Stop()
}

// 返回当前运行状态
func (self *Logic) Status() int {
	self.RWMutex.RLock()
	defer self.RWMutex.RUnlock()
	return self.status
}

// 返回当前运行状态
func (self *Logic) setStatus(status int) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	self.status = status
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
		self.setStatus(status.STOP)
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	// 便利添加任务到库
	tasksNum, spidersNum := self.addNewTask()

	if tasksNum == 0 {
		return
	}

	// 打印报告
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	logs.Log.Informational(" *                               —— 本次成功添加 %v 条任务，共包含 %v 条采集规则 ——", tasksNum, spidersNum)
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

}

// 服务器模式下，生成task并添加至库
func (self *Logic) addNewTask() (tasksNum, spidersNum int) {
	length := self.SpiderQueue.Len()
	t := distribute.Task{}
	// 从配置读取字段
	self.setTask(&t)

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
		self.setStatus(status.STOP)
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	for {
		// 从任务库获取一个任务
		t := self.downTask()

		// 准备运行
		self.taskToRun(t)

		// 重置计数
		self.sum[0], self.sum[1] = 0, 0
		// 重置计时
		self.takeTime = 0

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
	self.setAppConf(t)

	// 初始化蜘蛛队列
	for _, n := range t.Spiders {
		sp := spider.Menu.GetByName(n["name"])
		if sp == nil {
			continue
		}
		spcopy := sp.Copy()
		spcopy.SetPausetime(t.Pausetime)
		if spcopy.GetMaxPage() > 0 {
			spcopy.SetMaxPage(t.MaxPage)
		} else {
			spcopy.SetMaxPage(-1 * t.MaxPage)
		}
		if v, ok := n["keyword"]; ok {
			spcopy.SetKeyword(v)
		}
		self.SpiderQueue.Add(spcopy)
	}
}

// 开始执行任务
func (self *Logic) exec() {
	count := self.SpiderQueue.Len()
	cache.ReSetPageCount()

	// 初始化资源队列
	scheduler.Init()

	// 设置爬虫队列
	crawlCap := self.CrawlPool.Reset(count)

	logs.Log.Informational(" *     执行任务总数(任务数[*关键词数])为 %v 个\n", count)
	logs.Log.Informational(" *     爬虫池容量为 %v\n", crawlCap)
	logs.Log.Informational(" *     并发协程最多 %v 个\n", self.AppConf.ThreadNum)
	logs.Log.Informational(" *     随机停顿区间为 %v~%v 毫秒\n", self.AppConf.Pausetime/2, self.AppConf.Pausetime*2)
	logs.Log.Notice(" *                                                                                                 —— 开始抓取，请耐心等候 ——")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	// 开始计时
	cache.StartTime = time.Now()

	// 根据模式选择合理的并发
	if self.AppConf.Mode == status.OFFLINE {
		// 可控制执行状态
		go self.goRun(count)
	} else {
		// 保证接收服务端任务的同步
		self.goRun(count)
	}
}

// 任务执行
func (self *Logic) goRun(count int) {
	// 执行任务
	for i := 0; i < count && self.Status() != status.STOP; i++ {
	wait:
		if self.Status() == status.PAUSE {
			time.Sleep(1e9)
			goto wait
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
	for i := 0; i < count && self.Status() != status.STOP; i++ {
		s := <-cache.ReportChan
		if (s.DataNum == 0) && (s.FileNum == 0) {
			continue
		}
		logs.Log.Informational(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			logs.Log.Notice(" *     [输出报告 -> 任务：%v | 关键词：%v]   共输出数据 %v 条，用时 %v！\n", s.SpiderName, s.Keyword, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			logs.Log.Notice(" *     [输出报告 -> 任务：%v | 关键词：%v]   共下载文件 %v 个，用时 %v！\n", s.SpiderName, s.Keyword, s.FileNum, s.Time)
		default:
			logs.Log.Notice(" *     [输出报告 -> 任务：%v | 关键词：%v]   共输出数据 %v 条 + 下载文件 %v 个，用时 %v！\n", s.SpiderName, s.Keyword, s.DataNum, s.FileNum, s.Time)
		}

		self.sum[0] += s.DataNum
		self.sum[1] += s.FileNum
	}
	// 总耗时
	self.takeTime = time.Since(cache.StartTime)
	var prefix = func() string {
		if self.Status() == status.STOP {
			return "任务中途取消："
		}
		return "本次"
	}()
	// 打印总结报告
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	switch {
	case self.sum[0] > 0 && self.sum[1] == 0:
		logs.Log.Notice(" *                            —— %s合计输出【数据 %v 条】， 实爬URL【成功 %v 页 + 失败 %v 页 = 合计 %v 页】，耗时【%v】 ——", prefix, self.sum[0], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] > 0:
		logs.Log.Notice(" *                            —— %s合计输出【文件 %v 个】， 实爬URL【成功 %v 页 + 失败 %v 页 = 合计 %v 页】，耗时【%v】 ——", prefix, self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] == 0:
		logs.Log.Notice(" *                            —— %s无结果输出，实爬URL【成功 %v 页 + 失败 %v 页 = 合计 %v 页】，耗时【%v】 ——", prefix, cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	default:
		logs.Log.Notice(" *                            —— %s合计输出【数据 %v 条 + 文件 %v 个】，实爬URL【成功 %v 页 + 失败 %v 页 = 合计 %v 页】，耗时【%v】 ——", prefix, self.sum[0], self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	}
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	// 单机模式并发运行，需要标记任务结束
	if self.AppConf.Mode == status.OFFLINE {
		self.setStatus(status.STOP)
		self.LogRest()
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
		self.Teleport.Request(msg, "log", "")
	}
}

func (self *Logic) checkPort() bool {
	if self.AppConf.Port == 0 {
		logs.Log.Warning(" *     —— 亲，分布式端口不能为空哦~")
		return false
	}
	return true
}

func (self *Logic) checkAll() bool {
	if self.AppConf.Master == "" || !self.checkPort() {
		logs.Log.Warning(" *     —— 亲，服务器地址不能为空哦~")
		return false
	}
	return true
}

// 设置任务运行时公共配置
func (self *Logic) setAppConf(task *distribute.Task) {
	self.AppConf.ThreadNum = task.ThreadNum
	self.AppConf.Pausetime = task.Pausetime
	self.AppConf.OutType = task.OutType
	self.AppConf.DockerCap = task.DockerCap
	self.AppConf.DockerQueueCap = task.DockerQueueCap
	self.AppConf.SuccessInherit = task.SuccessInherit
	self.AppConf.FailureInherit = task.FailureInherit
	self.AppConf.MaxPage = task.MaxPage
	self.AppConf.ProxyMinute = task.ProxyMinute
	self.AppConf.Keywords = task.Keywords
}
func (self *Logic) setTask(task *distribute.Task) {
	task.ThreadNum = self.AppConf.ThreadNum
	task.Pausetime = self.AppConf.Pausetime
	task.OutType = self.AppConf.OutType
	task.DockerCap = self.AppConf.DockerCap
	task.DockerQueueCap = self.AppConf.DockerQueueCap
	task.SuccessInherit = self.AppConf.SuccessInherit
	task.FailureInherit = self.AppConf.FailureInherit
	task.MaxPage = self.AppConf.MaxPage
	task.ProxyMinute = self.AppConf.ProxyMinute
	task.Keywords = self.AppConf.Keywords
}
