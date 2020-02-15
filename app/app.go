// app interface for graphical user interface.
// 基本业务执行顺序依次为：New()-->[SetLog(io.Writer)-->]Init()-->SpiderPrepare()-->Run()
package app

import (
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/app/crawler"
	"github.com/henrylee2cn/pholcus/app/distribute"
	"github.com/henrylee2cn/pholcus/app/pipeline"
	"github.com/henrylee2cn/pholcus/app/pipeline/collector"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/app/spider"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/henrylee2cn/teleport"
)

type (
	App interface {
		SetLog(io.Writer) App                                         // 设置全局log实时显示终端
		LogGoOn() App                                                 // 继续log打印
		LogRest() App                                                 // 暂停log打印
		Init(mode int, port int, master string, w ...io.Writer) App   // 使用App前必须进行先Init初始化，SetLog()除外
		ReInit(mode int, port int, master string, w ...io.Writer) App // 切换运行模式并重设log打印目标
		GetAppConf(k ...string) interface{}                           // 获取全局参数
		SetAppConf(k string, v interface{}) App                       // 设置全局参数（client模式下不调用该方法）
		SpiderPrepare(original []*spider.Spider) App                  // 须在设置全局运行参数后Run()前调用（client模式下不调用该方法）
		Run()                                                         // 阻塞式运行直至任务完成（须在所有应当配置项配置完成后调用）
		Stop()                                                        // Offline 模式下中途终止任务（对外为阻塞式运行直至当前任务终止）
		IsRunning() bool                                              // 检查任务是否正在运行
		IsPause() bool                                                // 检查任务是否处于暂停状态
		IsStopped() bool                                              // 检查任务是否已经终止
		PauseRecover()                                                // Offline 模式下暂停\恢复任务
		Status() int                                                  // 返回当前状态
		GetSpiderLib() []*spider.Spider                               // 获取全部蜘蛛种类
		GetSpiderByName(string) *spider.Spider                        // 通过名字获取某蜘蛛
		GetSpiderQueue() crawler.SpiderQueue                          // 获取蜘蛛队列接口实例
		GetOutputLib() []string                                       // 获取全部输出方式
		GetTaskJar() *distribute.TaskJar                              // 返回任务库
		distribute.Distributer                                        // 实现分布式接口
	}
	Logic struct {
		*cache.AppConf                      // 全局配置
		*spider.SpiderSpecies               // 全部蜘蛛种类
		crawler.SpiderQueue                 // 当前任务的蜘蛛队列
		*distribute.TaskJar                 // 服务器与客户端间传递任务的存储库
		crawler.CrawlerPool                 // 爬行回收池
		teleport.Teleport                   // socket长连接双工通信接口，json数据传输
		sum                   [2]uint64     // 执行计数
		takeTime              time.Duration // 执行计时
		status                int           // 运行状态
		finish                chan bool
		finishOnce            sync.Once
		canSocketLog          bool
		sync.RWMutex
	}
)

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
	Limit          int64  // 采集上限，0为不限，若在规则中设置初始值为LIMIT则为自定义限制，否则默认限制请求数
	ProxyMinute    int64  // 代理IP更换的间隔分钟数
	// 选填项
	Keyins string // 自定义输入，后期切分为多个任务的Keyin自定义配置
}
*/

// 全局唯一的核心接口实例
var LogicApp = New()

func New() App {
	return newLogic()
}

func newLogic() *Logic {
	return &Logic{
		AppConf:       cache.Task,
		SpiderSpecies: spider.Species,
		status:        status.STOPPED,
		Teleport:      teleport.New(),
		TaskJar:       distribute.NewTaskJar(),
		SpiderQueue:   crawler.NewSpiderQueue(),
		CrawlerPool:   crawler.NewCrawlerPool(),
	}
}

// 设置全局log实时显示终端
func (self *Logic) SetLog(w io.Writer) App {
	logs.Log.SetOutput(w)
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
			logs.Log.Error("%v", err)
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
			logs.Log.Error("%v", err)
		}
	}()
	if k == "Limit" && v.(int64) <= 0 {
		v = int64(spider.LIMIT)
	} else if k == "DockerCap" && v.(int) < 1 {
		v = int(1)
	}
	acv := reflect.ValueOf(self.AppConf).Elem()
	key := strings.Title(k)
	if acv.FieldByName(key).CanSet() {
		acv.FieldByName(key).Set(reflect.ValueOf(v))
	}

	return self
}

// 使用App前必须先进行Init初始化（SetLog()除外）
func (self *Logic) Init(mode int, port int, master string, w ...io.Writer) App {
	self.canSocketLog = false
	if len(w) > 0 {
		self.SetLog(w[0])
	}
	self.LogGoOn()

	self.AppConf.Mode, self.AppConf.Port, self.AppConf.Master = mode, port, master
	self.Teleport = teleport.New()
	self.TaskJar = distribute.NewTaskJar()
	self.SpiderQueue = crawler.NewSpiderQueue()
	self.CrawlerPool = crawler.NewCrawlerPool()

	switch self.AppConf.Mode {
	case status.SERVER:
		logs.Log.EnableStealOne(false)
		if self.checkPort() {
			logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 服务器 ] 模式！！")
			self.Teleport.SetAPI(distribute.MasterApi(self)).Server(":" + strconv.Itoa(self.AppConf.Port))
		}

	case status.CLIENT:
		if self.checkAll() {
			logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 客户端 ] 模式！！")
			self.Teleport.SetAPI(distribute.SlaveApi(self)).Client(self.AppConf.Master, ":"+strconv.Itoa(self.AppConf.Port))
			// 开启节点间log打印
			self.canSocketLog = true
			logs.Log.EnableStealOne(true)
			go self.socketLog()
		}
	case status.OFFLINE:
		logs.Log.EnableStealOne(false)
		logs.Log.Informational("                                                                                               ！！当前运行模式为：[ 单机 ] 模式！！")
		return self
	default:
		logs.Log.Warning(" *    ——请指定正确的运行模式！——")
		return self
	}
	return self
}

// 切换运行模式时使用
func (self *Logic) ReInit(mode int, port int, master string, w ...io.Writer) App {
	if !self.IsStopped() {
		self.Stop()
	}
	self.LogRest()
	if self.Teleport != nil {
		self.Teleport.Close()
	}
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
// 已被显式赋值过的spider将不再重新分配Keyin
// client模式下不调用该方法
func (self *Logic) SpiderPrepare(original []*spider.Spider) App {
	self.SpiderQueue.Reset()
	// 遍历任务
	for _, sp := range original {
		spcopy := sp.Copy()
		spcopy.SetPausetime(self.AppConf.Pausetime)
		if spcopy.GetLimit() == spider.LIMIT {
			spcopy.SetLimit(self.AppConf.Limit)
		} else {
			spcopy.SetLimit(-1 * self.AppConf.Limit)
		}
		self.SpiderQueue.Add(spcopy)
	}
	// 遍历自定义配置
	self.SpiderQueue.AddKeyins(self.AppConf.Keyins)
	return self
}

// 获取全部输出方式
func (self *Logic) GetOutputLib() []string {
	return collector.DataOutputLib
}

// 获取全部蜘蛛种类
func (self *Logic) GetSpiderLib() []*spider.Spider {
	return self.SpiderSpecies.Get()
}

// 通过名字获取某蜘蛛
func (self *Logic) GetSpiderByName(name string) *spider.Spider {
	return self.SpiderSpecies.GetByName(name)
}

// 返回当前运行模式
func (self *Logic) GetMode() int {
	return self.AppConf.Mode
}

// 返回任务库
func (self *Logic) GetTaskJar() *distribute.TaskJar {
	return self.TaskJar
}

// 服务器客户端模式下返回节点数
func (self *Logic) CountNodes() int {
	return self.Teleport.CountNodes()
}

// 获取蜘蛛队列接口实例
func (self *Logic) GetSpiderQueue() crawler.SpiderQueue {
	return self.SpiderQueue
}

// 运行任务
func (self *Logic) Run() {
	// 确保开启报告
	self.LogGoOn()
	if self.AppConf.Mode != status.CLIENT && self.SpiderQueue.Len() == 0 {
		logs.Log.Warning(" *     —— 亲，任务列表不能为空哦~")
		self.LogRest()
		return
	}
	self.finish = make(chan bool)
	self.finishOnce = sync.Once{}
	// 重置计数
	self.sum[0], self.sum[1] = 0, 0
	// 重置计时
	self.takeTime = 0
	// 设置状态
	self.setStatus(status.RUN)
	defer self.setStatus(status.STOPPED)
	// 任务执行
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
	if self.status == status.STOPPED {
		return
	}
	if self.status != status.STOP {
		// 不可颠倒停止的顺序
		self.setStatus(status.STOP)
		// println("scheduler.Stop()")
		scheduler.Stop()
		// println("self.CrawlerPool.Stop()")
		self.CrawlerPool.Stop()
	}
	// println("wait self.IsStopped()")
	for !self.IsStopped() {
		time.Sleep(time.Second)
	}
}

// 检查任务是否正在运行
func (self *Logic) IsRunning() bool {
	return self.status == status.RUN
}

// 检查任务是否处于暂停状态
func (self *Logic) IsPause() bool {
	return self.status == status.PAUSE
}

// 检查任务是否已经终止
func (self *Logic) IsStopped() bool {
	return self.status == status.STOPPED
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
// 生成的任务与自身当前全局配置相同
func (self *Logic) server() {
	// 标记结束
	defer func() {
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

		t.Spiders = append(t.Spiders, map[string]string{"name": sp.GetName(), "keyin": sp.GetKeyin()})
		spidersNum++

		// 每十个蜘蛛存为一个任务
		if i > 0 && i%10 == 0 && length > 10 {
			// 存入
			one := t
			self.TaskJar.Push(&one)
			// logs.Log.App(" *     [新增任务]   详情： %#v", *t)

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
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	for {
		// 从任务库获取一个任务
		t := self.downTask()

		if self.Status() == status.STOP || self.Status() == status.STOPPED {
			return
		}

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
	if self.Status() == status.STOP || self.Status() == status.STOPPED {
		return nil
	}
	if self.CountNodes() == 0 && self.TaskJar.Len() == 0 {
		time.Sleep(time.Second)
		goto ReStartLabel
	}

	if self.TaskJar.Len() == 0 {
		self.Request(nil, "task", "")
		for self.TaskJar.Len() == 0 {
			if self.CountNodes() == 0 {
				goto ReStartLabel
			}
			time.Sleep(time.Second)
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
		sp := self.GetSpiderByName(n["name"])
		if sp == nil {
			continue
		}
		spcopy := sp.Copy()
		spcopy.SetPausetime(t.Pausetime)
		if spcopy.GetLimit() > 0 {
			spcopy.SetLimit(t.Limit)
		} else {
			spcopy.SetLimit(-1 * t.Limit)
		}
		if v, ok := n["keyin"]; ok {
			spcopy.SetKeyin(v)
		}
		self.SpiderQueue.Add(spcopy)
	}
}

// 开始执行任务
func (self *Logic) exec() {
	count := self.SpiderQueue.Len()
	cache.ResetPageCount()
	// 刷新输出方式的状态
	pipeline.RefreshOutput()
	// 初始化资源队列
	scheduler.Init()

	// 设置爬虫队列
	crawlerCap := self.CrawlerPool.Reset(count)

	logs.Log.Informational(" *     执行任务总数(任务数[*自定义配置数])为 %v 个\n", count)
	logs.Log.Informational(" *     采集引擎池容量为 %v\n", crawlerCap)
	logs.Log.Informational(" *     并发协程最多 %v 个\n", self.AppConf.ThreadNum)
	logs.Log.Informational(" *     默认随机停顿 %v~%v 毫秒\n", self.AppConf.Pausetime/2, self.AppConf.Pausetime*2)
	logs.Log.App(" *                                                                                                 —— 开始抓取，请耐心等候 ——")
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
	var i int
	for i = 0; i < count && self.Status() != status.STOP; i++ {
	pause:
		if self.IsPause() {
			time.Sleep(time.Second)
			goto pause
		}
		// 从爬行队列取出空闲蜘蛛，并发执行
		c := self.CrawlerPool.Use()
		if c != nil {
			go func(i int, c crawler.Crawler) {
				// 执行并返回结果消息
				c.Init(self.SpiderQueue.GetByIndex(i)).Run()
				// 任务结束后回收该蜘蛛
				self.RWMutex.RLock()
				if self.status != status.STOP {
					self.CrawlerPool.Free(c)
				}
				self.RWMutex.RUnlock()
			}(i, c)
		}
	}
	// 监控结束任务
	for ii := 0; ii < i; ii++ {
		s := <-cache.ReportChan
		if (s.DataNum == 0) && (s.FileNum == 0) {
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   无采集结果，用时 %v！\n", s.SpiderName, s.Keyin, s.Time)
			continue
		}
		logs.Log.Informational(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   共采集数据 %v 条，用时 %v！\n",
				s.SpiderName, s.Keyin, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   共下载文件 %v 个，用时 %v！\n",
				s.SpiderName, s.Keyin, s.FileNum, s.Time)
		default:
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   共采集数据 %v 条 + 下载文件 %v 个，用时 %v！\n",
				s.SpiderName, s.Keyin, s.DataNum, s.FileNum, s.Time)
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
		logs.Log.App(" *                            —— %s合计采集【数据 %v 条】， 实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, self.sum[0], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] > 0:
		logs.Log.App(" *                            —— %s合计采集【文件 %v 个】， 实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] == 0:
		logs.Log.App(" *                            —— %s无采集结果，实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	default:
		logs.Log.App(" *                            —— %s合计采集【数据 %v 条 + 文件 %v 个】，实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, self.sum[0], self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	}
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	// 单机模式并发运行，需要标记任务结束
	if self.AppConf.Mode == status.OFFLINE {
		self.LogRest()
		self.finishOnce.Do(func() { close(self.finish) })
	}
}

// 客户端向服务端反馈日志
func (self *Logic) socketLog() {
	for self.canSocketLog {
		_, msg, ok := logs.Log.StealOne()
		if !ok {
			return
		}
		if self.Teleport.CountNodes() == 0 {
			// 与服务器失去连接后，抛掉返馈日志
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
	self.AppConf.SuccessInherit = task.SuccessInherit
	self.AppConf.FailureInherit = task.FailureInherit
	self.AppConf.Limit = task.Limit
	self.AppConf.ProxyMinute = task.ProxyMinute
	self.AppConf.Keyins = task.Keyins
}
func (self *Logic) setTask(task *distribute.Task) {
	task.ThreadNum = self.AppConf.ThreadNum
	task.Pausetime = self.AppConf.Pausetime
	task.OutType = self.AppConf.OutType
	task.DockerCap = self.AppConf.DockerCap
	task.SuccessInherit = self.AppConf.SuccessInherit
	task.FailureInherit = self.AppConf.FailureInherit
	task.Limit = self.AppConf.Limit
	task.ProxyMinute = self.AppConf.ProxyMinute
	task.Keyins = self.AppConf.Keyins
}
