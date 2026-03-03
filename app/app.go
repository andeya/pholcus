// app interface for graphical user interface.
// Basic execution order: New() --> [SetLog(io.Writer) -->] Init() --> SpiderPrepare() --> Run()

// Package app 提供了爬虫应用的主入口与任务调度功能。
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
	"github.com/andeya/pholcus/app/downloader"
	"github.com/andeya/pholcus/app/pipeline"
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
		IsPaused() bool                                               // Check if task is paused
		IsStopped() bool                                              // Check if task has stopped
		PauseRecover()                                                // Pause or resume task in Offline mode
		Status() int                                                  // Return current status
		GetSpiderLib() []*spider.Spider                               // Get all spider species
		GetSpiderByName(string) option.Option[*spider.Spider]         // Get spider by name
		GetSpiderQueue() crawler.SpiderQueue                          // Get spider queue interface
		GetOutputLib() []string                                       // Get all output methods
		GetTaskJar() *distribute.TaskJar                              // Return task jar
		distribute.Distributor                                        // Implements distributed interface
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
		CrawlerPool:   crawler.NewCrawlerPool(downloader.SurferDownloader),
	}
}

// SetLog sets global log output to the given writer.
func (l *Logic) SetLog(w io.Writer) App {
	logs.Log().SetOutput(w)
	return l
}

// LogRest pauses log output.
func (l *Logic) LogRest() App {
	logs.Log().PauseOutput()
	return l
}

// LogGoOn resumes log output.
func (l *Logic) LogGoOn() App {
	logs.Log().GoOn()
	return l
}

// GetAppConf returns global config value(s).
func (l *Logic) GetAppConf(k ...string) interface{} {
	defer func() {
		if err := recover(); err != nil {
			logs.Log().Error("panic recovered: %v\n%s", err, debug.Stack())
		}
	}()
	if len(k) == 0 {
		return l.AppConf
	}
	key := titleCase(k[0])
	acv := reflect.ValueOf(l.AppConf).Elem()
	return acv.FieldByName(key).Interface()
}

// SetAppConf sets a global config value.
func (l *Logic) SetAppConf(k string, v interface{}) App {
	defer func() {
		if err := recover(); err != nil {
			logs.Log().Error("panic recovered: %v\n%s", err, debug.Stack())
		}
	}()
	if k == "Limit" && v.(int64) <= 0 {
		v = int64(spider.LIMIT)
	} else if k == "BatchCap" && v.(int) < 1 {
		v = int(1)
	}
	acv := reflect.ValueOf(l.AppConf).Elem()
	key := titleCase(k)
	if acv.FieldByName(key).CanSet() {
		acv.FieldByName(key).Set(reflect.ValueOf(v))
	}

	return l
}

// Init initializes the app; must be called before use (except SetLog).
func (l *Logic) Init(mode int, port int, master string, w ...io.Writer) App {
	l.AppConf = cache.Task

	l.canSocketLog = false
	if len(w) > 0 {
		l.SetLog(w[0])
	}
	l.LogGoOn()

	l.AppConf.Mode, l.AppConf.Port, l.AppConf.Master = mode, port, master
	l.Teleport = teleport.New()
	l.TaskJar = distribute.NewTaskJar()
	l.SpiderQueue = crawler.NewSpiderQueue()
	l.CrawlerPool = crawler.NewCrawlerPool(downloader.SurferDownloader)

	switch l.AppConf.Mode {
	case status.SERVER:
		logs.Log().EnableStealOne(false)
		if l.checkPort() {
			logs.Log().Informational("                                                                                               !! Current run mode: [ SERVER ] !!")
			l.Teleport.SetAPI(distribute.MasterAPI(l)).Server(":" + strconv.Itoa(l.AppConf.Port))
		}

	case status.CLIENT:
		if l.checkAll() {
			logs.Log().Informational("                                                                                               !! Current run mode: [ CLIENT ] !!")
			l.Teleport.SetAPI(distribute.SlaveAPI(l)).Client(l.AppConf.Master, ":"+strconv.Itoa(l.AppConf.Port))
			// Enable inter-node log forwarding
			l.canSocketLog = true
			logs.Log().EnableStealOne(true)
			go l.socketLog()
		}
	case status.OFFLINE:
		logs.Log().EnableStealOne(false)
		logs.Log().Informational("                                                                                               !! Current run mode: [ OFFLINE ] !!")
		return l
	default:
		logs.Log().Warning(" *    —— Please specify a valid run mode! ——")
		return l
	}
	return l
}

// ReInit switches run mode; use when changing mode.
func (l *Logic) ReInit(mode int, port int, master string, w ...io.Writer) App {
	if !l.IsStopped() {
		l.Stop()
	}
	l.LogRest()
	if l.Teleport != nil {
		l.Teleport.Close()
	}
	// Wait for shutdown
	if mode == status.UNSET {
		l = newLogic()
		l.AppConf.Mode = status.UNSET
		return l
	}
	// Restart
	l = newLogic().Init(mode, port, master, w...).(*Logic)
	return l
}

// SpiderPrepare must be called after setting global params and immediately before Run().
// original is the raw spider species from spider package without prior assignment.
// Spiders with explicit Keyin are not reassigned.
// Not called in client mode.
func (l *Logic) SpiderPrepare(original []*spider.Spider) App {
	l.SpiderQueue.Reset()
	for _, sp := range original {
		spcopy := sp.Copy()
		spcopy.SetPausetime(l.AppConf.Pausetime)
		if spcopy.GetLimit() == spider.LIMIT {
			spcopy.SetLimit(l.AppConf.Limit)
		} else {
			spcopy.SetLimit(-1 * l.AppConf.Limit)
		}
		l.SpiderQueue.Add(spcopy)
	}
	l.SpiderQueue.AddKeyins(l.AppConf.Keyins)
	return l
}

// GetOutputLib returns all output methods.
func (l *Logic) GetOutputLib() []string {
	return pipeline.GetOutputLib()
}

// GetSpiderLib returns all spider species.
func (l *Logic) GetSpiderLib() []*spider.Spider {
	return l.SpiderSpecies.Get()
}

// GetSpiderByName returns a spider by name.
func (l *Logic) GetSpiderByName(name string) option.Option[*spider.Spider] {
	return l.SpiderSpecies.GetByNameOpt(name)
}

// GetMode returns current run mode.
func (l *Logic) GetMode() int {
	return l.AppConf.Mode
}

// GetTaskJar returns the task jar.
func (l *Logic) GetTaskJar() *distribute.TaskJar {
	return l.TaskJar
}

// CountNodes returns connected node count in server/client mode.
func (l *Logic) CountNodes() int {
	return l.Teleport.CountNodes()
}

// GetSpiderQueue returns the spider queue interface.
func (l *Logic) GetSpiderQueue() crawler.SpiderQueue {
	return l.SpiderQueue
}

// Run executes the task.
func (l *Logic) Run() {
	l.LogGoOn()
	if l.AppConf.Mode != status.CLIENT && l.SpiderQueue.Len() == 0 {
		logs.Log().Warning(" *     —— Task list cannot be empty ——")
		l.LogRest()
		return
	}
	l.finish = make(chan bool)
	l.finishOnce = sync.Once{}
	l.sum[0], l.sum[1] = 0, 0
	l.takeTime = 0
	l.setStatus(status.RUN)
	defer l.setStatus(status.STOPPED)
	switch l.AppConf.Mode {
	case status.OFFLINE:
		l.offline()
	case status.SERVER:
		l.server()
	case status.CLIENT:
		l.client()
	default:
		return
	}
	<-l.finish
}

// PauseRecover pauses or resumes the task in Offline mode.
func (l *Logic) PauseRecover() {
	switch l.Status() {
	case status.PAUSE:
		l.setStatus(status.RUN)
	case status.RUN:
		l.setStatus(status.PAUSE)
	}

	scheduler.PauseRecover()
}

// Stop terminates the task mid-run in Offline mode.
func (l *Logic) Stop() {
	if l.status == status.STOPPED {
		return
	}
	if l.status != status.STOP {
		// Stop order must not be reversed
		l.setStatus(status.STOP)
		scheduler.Stop()
		l.CrawlerPool.Stop()
	}
	for !l.IsStopped() {
		time.Sleep(time.Second)
	}
}

// IsRunning reports whether the task is running.
func (l *Logic) IsRunning() bool {
	return l.status == status.RUN
}

// IsPaused reports whether the task is paused.
func (l *Logic) IsPaused() bool {
	return l.status == status.PAUSE
}

// IsStopped reports whether the task has stopped.
func (l *Logic) IsStopped() bool {
	return l.status == status.STOPPED
}

// Status returns current run status.
func (l *Logic) Status() int {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()
	return l.status
}

// setStatus sets the run status.
func (l *Logic) setStatus(status int) {
	l.RWMutex.Lock()
	defer l.RWMutex.Unlock()
	l.status = status
}

// --- Private methods ---

// offline runs in offline mode.
func (l *Logic) offline() {
	l.exec()
}

// server runs in server mode; must be called after SpiderPrepare() to add tasks.
// Generated tasks use the same global config.
func (l *Logic) server() {
	defer func() {
		l.finishOnce.Do(func() { close(l.finish) })
	}()

	tasksNum, spidersNum := l.addNewTask()

	if tasksNum == 0 {
		return
	}

	logs.Log().Informational(" * ")
	logs.Log().Informational(` *********************************************************************************************************************************** `)
	logs.Log().Informational(" * ")
	logs.Log().Informational(" *                               —— Successfully added %v tasks, %v spider rules in total ——", tasksNum, spidersNum)
	logs.Log().Informational(" * ")
	logs.Log().Informational(` *********************************************************************************************************************************** `)
}

// addNewTask generates tasks and adds them to the jar in server mode.
func (l *Logic) addNewTask() (tasksNum, spidersNum int) {
	length := l.SpiderQueue.Len()
	t := distribute.Task{}
	l.setTask(&t)

	for i, sp := range l.SpiderQueue.GetAll() {

		t.Spiders = append(t.Spiders, map[string]string{"name": sp.GetName(), "keyin": sp.GetKeyin()})
		spidersNum++

		if i > 0 && i%10 == 0 && length > 10 {
			one := t
			l.TaskJar.Push(&one)
			tasksNum++
			t.Spiders = []map[string]string{}
		}
	}

	if len(t.Spiders) != 0 {
		one := t
		l.TaskJar.Push(&one)
		tasksNum++
	}
	return
}

// client runs in client mode.
func (l *Logic) client() {
	defer func() {
		l.finishOnce.Do(func() { close(l.finish) })
	}()

	for {
		t := l.downTask()
		if l.Status() == status.STOP || l.Status() == status.STOPPED {
			return
		}
		l.taskToRun(t)
		l.sum[0], l.sum[1] = 0, 0
		l.takeTime = 0
		l.exec()
	}
}

// downTask fetches a task from the jar in client mode.
func (l *Logic) downTask() *distribute.Task {
	for {
		if l.Status() == status.STOP || l.Status() == status.STOPPED {
			return nil
		}
		if l.CountNodes() == 0 && l.TaskJar.Len() == 0 {
			time.Sleep(time.Second)
			continue
		}

		if l.TaskJar.Len() == 0 {
			l.Request(nil, "task", "")
			for l.TaskJar.Len() == 0 {
				if l.CountNodes() == 0 {
					break
				}
				time.Sleep(time.Second)
			}
			if l.TaskJar.Len() == 0 {
				continue
			}
		}
		return l.TaskJar.Pull()
	}
}

// taskToRun prepares run conditions from a task in client mode.
func (l *Logic) taskToRun(t *distribute.Task) {
	l.SpiderQueue.Reset()
	l.setAppConf(t)

	for _, n := range t.Spiders {
		spOpt := l.SpiderSpecies.GetByNameOpt(n["name"])
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
		l.SpiderQueue.Add(spcopy)
	}
}

// exec starts task execution.
func (l *Logic) exec() {
	count := l.SpiderQueue.Len()
	cache.ResetPageCount()
	pipeline.RefreshOutput()
	scheduler.Init(l.AppConf.ThreadNum, l.AppConf.ProxyMinute)
	l.CrawlerPool.SetPipelineConfig(l.AppConf.OutType, l.AppConf.BatchCap)
	crawlerCap := l.CrawlerPool.Reset(count)

	logs.Log().Informational(" *     Total tasks (tasks * custom configs): %v\n", count)
	logs.Log().Informational(" *     Crawler pool capacity: %v\n", crawlerCap)
	logs.Log().Informational(" *     Max concurrent goroutines: %v\n", l.AppConf.ThreadNum)
	logs.Log().Informational(" *     Default random pause: %v~%v ms\n", l.AppConf.Pausetime/2, l.AppConf.Pausetime*2)
	logs.Log().App(" *                                                                                                 —— Starting crawl, please wait ——")
	logs.Log().Informational(` *********************************************************************************************************************************** `)

	cache.StartTime = time.Now()

	if l.AppConf.Mode == status.OFFLINE {
		go l.goRun(count)
	} else {
		l.goRun(count)
	}
}

// goRun executes the task.
func (l *Logic) goRun(count int) {
	var i int
	for i = 0; i < count && l.Status() != status.STOP; i++ {
		for l.IsPaused() {
			time.Sleep(time.Second)
		}
		if opt := l.CrawlerPool.UseOpt(); opt.IsSome() {
			c := opt.Unwrap()
			go func(i int, c crawler.Crawler) {
				c.Init(l.SpiderQueue.GetByIndex(i)).Run()
				l.RWMutex.RLock()
				if l.status != status.STOP {
					l.CrawlerPool.Free(c)
				}
				l.RWMutex.RUnlock()
			}(i, c)
		}
	}
	for ii := 0; ii < i; ii++ {
		s := <-cache.ReportChan
		if (s.DataNum == 0) && (s.FileNum == 0) {
			logs.Log().App(" *     [Task subtotal: %s | KEYIN: %s]   No results, duration %v\n", s.SpiderName, s.Keyin, s.Time)
			continue
		}
		logs.Log().Informational(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			logs.Log().App(" *     [Task subtotal: %s | KEYIN: %s]   Collected %v data items, duration %v\n",
				s.SpiderName, s.Keyin, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			logs.Log().App(" *     [Task subtotal: %s | KEYIN: %s]   Downloaded %v files, duration %v\n",
				s.SpiderName, s.Keyin, s.FileNum, s.Time)
		default:
			logs.Log().App(" *     [Task subtotal: %s | KEYIN: %s]   Collected %v data items + %v files, duration %v\n",
				s.SpiderName, s.Keyin, s.DataNum, s.FileNum, s.Time)
		}

		l.sum[0] += s.DataNum
		l.sum[1] += s.FileNum
	}

	l.takeTime = time.Since(cache.StartTime)
	var prefix = func() string {
		if l.Status() == status.STOP {
			return "Task cancelled: "
		}
		return "This run: "
	}()
	logs.Log().Informational(" * ")
	logs.Log().Informational(` *********************************************************************************************************************************** `)
	logs.Log().Informational(" * ")
	switch {
	case l.sum[0] > 0 && l.sum[1] == 0:
		logs.Log().App(" *                            —— %sTotal collected [%v data items], crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, l.sum[0], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), l.takeTime)
	case l.sum[0] == 0 && l.sum[1] > 0:
		logs.Log().App(" *                            —— %sTotal collected [%v files], crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, l.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), l.takeTime)
	case l.sum[0] == 0 && l.sum[1] == 0:
		logs.Log().App(" *                            —— %sNo results, crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), l.takeTime)
	default:
		logs.Log().App(" *                            —— %sTotal collected [%v data items + %v files], crawled [success %v URL + fail %v URL = total %v URL], duration [%v] ——",
			prefix, l.sum[0], l.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), l.takeTime)
	}
	logs.Log().Informational(" * ")
	logs.Log().Informational(` *********************************************************************************************************************************** `)

	if l.AppConf.Mode == status.OFFLINE {
		l.LogRest()
		l.finishOnce.Do(func() { close(l.finish) })
	}
}

// socketLog forwards client logs to the server.
func (l *Logic) socketLog() {
	for l.canSocketLog {
		_, msg, ok := logs.Log().StealOne()
		if !ok {
			return
		}
		if l.Teleport.CountNodes() == 0 {
			continue
		}
		l.Teleport.Request(msg, "log", "")
	}
}

func (l *Logic) checkPort() bool {
	if l.AppConf.Port == 0 {
		logs.Log().Warning(" *     —— Distributed port cannot be empty ——")
		return false
	}
	return true
}

func (l *Logic) checkAll() bool {
	if l.AppConf.Master == "" || !l.checkPort() {
		logs.Log().Warning(" *     —— Server address cannot be empty ——")
		return false
	}
	return true
}

// setAppConf applies task config to global runtime config.
func (l *Logic) setAppConf(task *distribute.Task) {
	l.AppConf.ThreadNum = task.ThreadNum
	l.AppConf.Pausetime = task.Pausetime
	l.AppConf.OutType = task.OutType
	l.AppConf.BatchCap = task.BatchCap
	l.AppConf.SuccessInherit = task.SuccessInherit
	l.AppConf.FailureInherit = task.FailureInherit
	l.AppConf.Limit = task.Limit
	l.AppConf.ProxyMinute = task.ProxyMinute
	l.AppConf.Keyins = task.Keyins
}
func (l *Logic) setTask(task *distribute.Task) {
	task.ThreadNum = l.AppConf.ThreadNum
	task.Pausetime = l.AppConf.Pausetime
	task.OutType = l.AppConf.OutType
	task.BatchCap = l.AppConf.BatchCap
	task.SuccessInherit = l.AppConf.SuccessInherit
	task.FailureInherit = l.AppConf.FailureInherit
	task.Limit = l.AppConf.Limit
	task.ProxyMinute = l.AppConf.ProxyMinute
	task.Keyins = l.AppConf.Keyins
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
