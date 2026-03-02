// app interface for graphical user interface.
// Basic execution order: New() --> [SetLog(io.Writer) -->] Init() --> SpiderPrepare() --> Run()
package app

import (
	"io"
	"reflect"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/andeya/gust/option"
	"github.com/andeya/pholcus/app/crawler"
	"github.com/andeya/pholcus/app/distribute"
	"github.com/andeya/pholcus/app/distribute/teleport"
	"github.com/andeya/pholcus/app/pipeline"
	"github.com/andeya/pholcus/app/pipeline/collector"
	"github.com/andeya/pholcus/app/scheduler"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
)

type (
	App interface {
		SetLog(io.Writer) App                                         // Set global log output to terminal
		LogGoOn() App                                                 // Resume log output
		LogRest() App                                                 // Pause log output
		Init(mode int, port int, master string, w ...io.Writer) App   // Must call Init before using App (except SetLog)
		ReInit(mode int, port int, master string, w ...io.Writer) App // Switch run mode and reset log output target
		GetAppConf(k ...string) interface{}                           // Get global config
		SetAppConf(k string, v interface{}) App                       // Set global config (not called in client mode)
		SpiderPrepare(original []*spider.Spider) App                  // Must call after setting global params and before Run() (not called in client mode)
		Run()                                                         // Block until task completes (call after all config is done)
		Stop()                                                        // Terminate task mid-run in Offline mode (blocks until current task stops)
		IsRunning() bool                                              // Check if task is running
		IsPause() bool                                                // Check if task is paused
		IsStopped() bool                                              // Check if task has stopped
		PauseRecover()                                                // Pause or resume task in Offline mode
		Status() int                                                  // Return current status
		GetSpiderLib() []*spider.Spider                               // Get all spider species
		GetSpiderByName(string) option.Option[*spider.Spider]         // Get spider by name
		GetSpiderQueue() crawler.SpiderQueue                          // Get spider queue interface
		GetOutputLib() []string                                       // Get all output methods
		GetTaskJar() *distribute.TaskJar                              // Return task jar
		distribute.Distributer                                        // Implements distributed interface
	}
	Logic struct {
		*cache.AppConf                      // Global config
		*spider.SpiderSpecies               // All spider species
		crawler.SpiderQueue                 // Spider queue for current task
		*distribute.TaskJar                 // Task storage passed between server and client
		crawler.CrawlerPool                 // Crawler pool
		teleport.Teleport                   // Socket duplex communication, JSON transport
		sum                   [2]uint64     // Execution count
		takeTime              time.Duration // Execution duration
		status                int           // Run status
		finish                chan bool
		finishOnce            sync.Once
		canSocketLog          bool
		sync.RWMutex
	}
)

/*
 * Common config for task runtime
type AppConf struct {
	Mode           int    // Node role
	Port           int    // Master node port
	Master         string // Master server address (no port)
	ThreadNum      int    // Global max concurrency
	Pausetime      int64  // Pause duration in ms (random: Pausetime/2 ~ Pausetime*2)
	OutType        string // Output method
	DockerCap      int    // Segment dump container capacity
	DockerQueueCap int    // Segment output pool capacity, >= 2
	SuccessInherit bool   // Inherit historical success records
	FailureInherit bool   // Inherit historical failure records
	Limit          int64  // Collection limit, 0=unlimited; if rule sets LIMIT then custom limit, else default request limit
	ProxyMinute    int64  // Proxy IP rotation interval in minutes
	// Optional
	Keyins string // Custom input, later split into Keyin config for multiple tasks
}
*/

// LogicApp is the global singleton core interface instance.
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

// SetLog sets global log output to the given writer.
func (self *Logic) SetLog(w io.Writer) App {
	logs.Log.SetOutput(w)
	return self
}

// LogRest pauses log output.
func (self *Logic) LogRest() App {
	logs.Log.Rest()
	return self
}

// LogGoOn resumes log output.
func (self *Logic) LogGoOn() App {
	logs.Log.GoOn()
	return self
}

// GetAppConf returns global config value(s).
func (self *Logic) GetAppConf(k ...string) interface{} {
	defer func() {
		if err := recover(); err != nil {
			logs.Log.Error("panic recovered: %v\n%s", err, debug.Stack())
		}
	}()
	if len(k) == 0 {
		return self.AppConf
	}
	key := titleCase(k[0])
	acv := reflect.ValueOf(self.AppConf).Elem()
	return acv.FieldByName(key).Interface()
}

// SetAppConf sets a global config value.
func (self *Logic) SetAppConf(k string, v interface{}) App {
	defer func() {
		if err := recover(); err != nil {
			logs.Log.Error("panic recovered: %v\n%s", err, debug.Stack())
		}
	}()
	if k == "Limit" && v.(int64) <= 0 {
		v = int64(spider.LIMIT)
	} else if k == "DockerCap" && v.(int) < 1 {
		v = int(1)
	}
	acv := reflect.ValueOf(self.AppConf).Elem()
	key := titleCase(k)
	if acv.FieldByName(key).CanSet() {
		acv.FieldByName(key).Set(reflect.ValueOf(v))
	}

	return self
}

// Init initializes the app; must be called before use (except SetLog).
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
			logs.Log.Informational("                                                                                               !! Current run mode: [ SERVER ] !!")
			self.Teleport.SetAPI(distribute.MasterApi(self)).Server(":" + strconv.Itoa(self.AppConf.Port))
		}

	case status.CLIENT:
		if self.checkAll() {
			logs.Log.Informational("                                                                                               !! Current run mode: [ CLIENT ] !!")
			self.Teleport.SetAPI(distribute.SlaveApi(self)).Client(self.AppConf.Master, ":"+strconv.Itoa(self.AppConf.Port))
			// Enable inter-node log forwarding
			self.canSocketLog = true
			logs.Log.EnableStealOne(true)
			go self.socketLog()
		}
	case status.OFFLINE:
		logs.Log.EnableStealOne(false)
		logs.Log.Informational("                                                                                               !! Current run mode: [ OFFLINE ] !!")
		return self
	default:
		logs.Log.Warning(" *    —— Please specify a valid run mode! ——")
		return self
	}
	return self
}

// ReInit switches run mode; use when changing mode.
func (self *Logic) ReInit(mode int, port int, master string, w ...io.Writer) App {
	if !self.IsStopped() {
		self.Stop()
	}
	self.LogRest()
	if self.Teleport != nil {
		self.Teleport.Close()
	}
	// Wait for shutdown
	if mode == status.UNSET {
		self = newLogic()
		self.AppConf.Mode = status.UNSET
		return self
	}
	// Restart
	self = newLogic().Init(mode, port, master, w...).(*Logic)
	return self
}

// SpiderPrepare must be called after setting global params and immediately before Run().
// original is the raw spider species from spider package without prior assignment.
// Spiders with explicit Keyin are not reassigned.
// Not called in client mode.
func (self *Logic) SpiderPrepare(original []*spider.Spider) App {
	self.SpiderQueue.Reset()
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
	self.SpiderQueue.AddKeyins(self.AppConf.Keyins)
	return self
}

// GetOutputLib returns all output methods.
func (self *Logic) GetOutputLib() []string {
	return collector.DataOutputLib
}

// GetSpiderLib returns all spider species.
func (self *Logic) GetSpiderLib() []*spider.Spider {
	return self.SpiderSpecies.Get()
}

// GetSpiderByName returns a spider by name.
func (self *Logic) GetSpiderByName(name string) option.Option[*spider.Spider] {
	return self.SpiderSpecies.GetByNameOpt(name)
}

// GetMode returns current run mode.
func (self *Logic) GetMode() int {
	return self.AppConf.Mode
}

// GetTaskJar returns the task jar.
func (self *Logic) GetTaskJar() *distribute.TaskJar {
	return self.TaskJar
}

// CountNodes returns connected node count in server/client mode.
func (self *Logic) CountNodes() int {
	return self.Teleport.CountNodes()
}

// GetSpiderQueue returns the spider queue interface.
func (self *Logic) GetSpiderQueue() crawler.SpiderQueue {
	return self.SpiderQueue
}

// Run executes the task.
func (self *Logic) Run() {
	self.LogGoOn()
	if self.AppConf.Mode != status.CLIENT && self.SpiderQueue.Len() == 0 {
		logs.Log.Warning(" *     —— Task list cannot be empty ——")
		self.LogRest()
		return
	}
	self.finish = make(chan bool)
	self.finishOnce = sync.Once{}
	self.sum[0], self.sum[1] = 0, 0
	self.takeTime = 0
	self.setStatus(status.RUN)
	defer self.setStatus(status.STOPPED)
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

// PauseRecover pauses or resumes the task in Offline mode.
func (self *Logic) PauseRecover() {
	switch self.Status() {
	case status.PAUSE:
		self.setStatus(status.RUN)
	case status.RUN:
		self.setStatus(status.PAUSE)
	}

	scheduler.PauseRecover()
}

// Stop terminates the task mid-run in Offline mode.
func (self *Logic) Stop() {
	if self.status == status.STOPPED {
		return
	}
	if self.status != status.STOP {
		// Stop order must not be reversed
		self.setStatus(status.STOP)
		scheduler.Stop()
		self.CrawlerPool.Stop()
	}
	for !self.IsStopped() {
		time.Sleep(time.Second)
	}
}

// IsRunning reports whether the task is running.
func (self *Logic) IsRunning() bool {
	return self.status == status.RUN
}

// IsPause reports whether the task is paused.
func (self *Logic) IsPause() bool {
	return self.status == status.PAUSE
}

// IsStopped reports whether the task has stopped.
func (self *Logic) IsStopped() bool {
	return self.status == status.STOPPED
}

// Status returns current run status.
func (self *Logic) Status() int {
	self.RWMutex.RLock()
	defer self.RWMutex.RUnlock()
	return self.status
}

// setStatus sets the run status.
func (self *Logic) setStatus(status int) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	self.status = status
}

// --- Private methods ---

// offline runs in offline mode.
func (self *Logic) offline() {
	self.exec()
}

// server runs in server mode; must be called after SpiderPrepare() to add tasks.
// Generated tasks use the same global config.
func (self *Logic) server() {
	defer func() {
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	tasksNum, spidersNum := self.addNewTask()

	if tasksNum == 0 {
		return
	}

	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	logs.Log.Informational(" *                               —— Successfully added %v tasks, %v spider rules in total ——", tasksNum, spidersNum)
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)
}

// addNewTask generates tasks and adds them to the jar in server mode.
func (self *Logic) addNewTask() (tasksNum, spidersNum int) {
	length := self.SpiderQueue.Len()
	t := distribute.Task{}
	self.setTask(&t)

	for i, sp := range self.SpiderQueue.GetAll() {

		t.Spiders = append(t.Spiders, map[string]string{"name": sp.GetName(), "keyin": sp.GetKeyin()})
		spidersNum++

		if i > 0 && i%10 == 0 && length > 10 {
			one := t
			self.TaskJar.Push(&one)
			tasksNum++
			t.Spiders = []map[string]string{}
		}
	}

	if len(t.Spiders) != 0 {
		one := t
		self.TaskJar.Push(&one)
		tasksNum++
	}
	return
}

// client runs in client mode.
func (self *Logic) client() {
	defer func() {
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	for {
		t := self.downTask()
		if self.Status() == status.STOP || self.Status() == status.STOPPED {
			return
		}
		self.taskToRun(t)
		self.sum[0], self.sum[1] = 0, 0
		self.takeTime = 0
		self.exec()
	}
}

// downTask fetches a task from the jar in client mode.
func (self *Logic) downTask() *distribute.Task {
	for {
		if self.Status() == status.STOP || self.Status() == status.STOPPED {
			return nil
		}
		if self.CountNodes() == 0 && self.TaskJar.Len() == 0 {
			time.Sleep(time.Second)
			continue
		}

		if self.TaskJar.Len() == 0 {
			self.Request(nil, "task", "")
			for self.TaskJar.Len() == 0 {
				if self.CountNodes() == 0 {
					break
				}
				time.Sleep(time.Second)
			}
			if self.TaskJar.Len() == 0 {
				continue
			}
		}
		return self.TaskJar.Pull()
	}
}

// taskToRun prepares run conditions from a task in client mode.
func (self *Logic) taskToRun(t *distribute.Task) {
	self.SpiderQueue.Reset()
	self.setAppConf(t)

	for _, n := range t.Spiders {
		spOpt := self.SpiderSpecies.GetByNameOpt(n["name"])
		if spOpt.IsNone() {
			continue
		}
		spcopy := spOpt.Unwrap().Copy()
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

// exec starts task execution.
func (self *Logic) exec() {
	count := self.SpiderQueue.Len()
	cache.ResetPageCount()
	pipeline.RefreshOutput()
	scheduler.Init()
	crawlerCap := self.CrawlerPool.Reset(count)

	logs.Log.Informational(" *     Total tasks (tasks * custom configs): %v\n", count)
	logs.Log.Informational(" *     Crawler pool capacity: %v\n", crawlerCap)
	logs.Log.Informational(" *     Max concurrent goroutines: %v\n", self.AppConf.ThreadNum)
	logs.Log.Informational(" *     Default random pause: %v~%v ms\n", self.AppConf.Pausetime/2, self.AppConf.Pausetime*2)
	logs.Log.App(" *                                                                                                 —— Starting crawl, please wait ——")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	cache.StartTime = time.Now()

	if self.AppConf.Mode == status.OFFLINE {
		go self.goRun(count)
	} else {
		self.goRun(count)
	}
}

// goRun executes the task.
func (self *Logic) goRun(count int) {
	var i int
	for i = 0; i < count && self.Status() != status.STOP; i++ {
		for self.IsPause() {
			time.Sleep(time.Second)
		}
		if opt := self.CrawlerPool.UseOpt(); opt.IsSome() {
			c := opt.Unwrap()
			go func(i int, c crawler.Crawler) {
				c.Init(self.SpiderQueue.GetByIndex(i)).Run()
				self.RWMutex.RLock()
				if self.status != status.STOP {
					self.CrawlerPool.Free(c)
				}
				self.RWMutex.RUnlock()
			}(i, c)
		}
	}
	for ii := 0; ii < i; ii++ {
		s := <-cache.ReportChan
		if (s.DataNum == 0) && (s.FileNum == 0) {
			logs.Log.App(" *     [Task subtotal: %s | KEYIN: %s]   No results, duration %v\n", s.SpiderName, s.Keyin, s.Time)
			continue
		}
		logs.Log.Informational(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			logs.Log.App(" *     [Task subtotal: %s | KEYIN: %s]   Collected %v data items, duration %v\n",
				s.SpiderName, s.Keyin, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			logs.Log.App(" *     [Task subtotal: %s | KEYIN: %s]   Downloaded %v files, duration %v\n",
				s.SpiderName, s.Keyin, s.FileNum, s.Time)
		default:
			logs.Log.App(" *     [Task subtotal: %s | KEYIN: %s]   Collected %v data items + %v files, duration %v\n",
				s.SpiderName, s.Keyin, s.DataNum, s.FileNum, s.Time)
		}

		self.sum[0] += s.DataNum
		self.sum[1] += s.FileNum
	}

	self.takeTime = time.Since(cache.StartTime)
	var prefix = func() string {
		if self.Status() == status.STOP {
			return "Task cancelled: "
		}
		return "This run: "
	}()
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	switch {
	case self.sum[0] > 0 && self.sum[1] == 0:
		logs.Log.App(" *                            —— %sTotal collected [%v data items], crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, self.sum[0], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] > 0:
		logs.Log.App(" *                            —— %sTotal collected [%v files], crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] == 0:
		logs.Log.App(" *                            —— %sNo results, crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	default:
		logs.Log.App(" *                            —— %sTotal collected [%v data items + %v files], crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, self.sum[0], self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	}
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)

	if self.AppConf.Mode == status.OFFLINE {
		self.LogRest()
		self.finishOnce.Do(func() { close(self.finish) })
	}
}

// socketLog forwards client logs to the server.
func (self *Logic) socketLog() {
	for self.canSocketLog {
		_, msg, ok := logs.Log.StealOne()
		if !ok {
			return
		}
		if self.Teleport.CountNodes() == 0 {
			continue
		}
		self.Teleport.Request(msg, "log", "")
	}
}

func (self *Logic) checkPort() bool {
	if self.AppConf.Port == 0 {
		logs.Log.Warning(" *     —— Distributed port cannot be empty ——")
		return false
	}
	return true
}

func (self *Logic) checkAll() bool {
	if self.AppConf.Master == "" || !self.checkPort() {
		logs.Log.Warning(" *     —— Server address cannot be empty ——")
		return false
	}
	return true
}

// setAppConf applies task config to global runtime config.
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

func titleCase(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
