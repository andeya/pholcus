package gui

import (
	"github.com/henrylee2cn/pholcus/crawl"
	"github.com/henrylee2cn/pholcus/crawl/scheduler"
	. "github.com/henrylee2cn/pholcus/gui/model"
	. "github.com/henrylee2cn/pholcus/node"
	"github.com/henrylee2cn/pholcus/reporter"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	_ "github.com/henrylee2cn/pholcus/spider/spiders"
	"log"
	"strconv"
	"time"
)

func Run() {
	runmodeWindow()
}

// 开始执行任务
func Exec(count int) {

	cache.ReqSum = 0

	// 初始化资源队列
	scheduler.Init(Input.ThreadNum)

	// 设置爬虫队列
	crawlNum := Pholcus.Crawls.Reset(count)

	log.Println(` ********************************************************************************************************************************************** `)
	log.Printf(" * ")
	log.Printf(" *     执行任务总数（任务数[*关键词数]）为 %v 个...\n", count)
	log.Printf(" *     爬虫队列可容纳蜘蛛 %v 只...\n", crawlNum)
	log.Printf(" *     并发协程最多 %v 个……\n", Input.ThreadNum)
	log.Printf(" *     随机停顿时间为 %v~%v ms ……\n", Input.BaseSleeptime, Input.BaseSleeptime+Input.RandomSleepPeriod)
	log.Printf(" * ")
	log.Printf(" *                                                                                                             —— 开始抓取，请耐心等候 ——")
	log.Printf(" * ")
	log.Println(` ********************************************************************************************************************************************** `)

	// 开始计时
	cache.StartTime = time.Now()

	// 任务执行
	status.Crawl = status.RUN

	// 根据模式选择合理的并发
	if cache.Task.RunMode == status.OFFLINE {
		go GoRun(count)
	} else {
		// 保证了打印信息的同步输出
		GoRun(count)
	}
}

// 任务执行
func GoRun(count int) {
	for i := 0; i < count && status.Crawl == status.RUN; i++ {
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
	sum := 0 //数据总数
	for i := 0; i < count; i++ {
		s := <-cache.ReportChan

		log.Println(` ********************************************************************************************************************************************** `)
		log.Printf(" * ")
		reporter.Log.Printf(" *     [结束报告 -> 任务：%v | 关键词：%v]   共输出数据 %v 条，用时 %v 分钟！\n", s.SpiderName, s.Keyword, s.Num, s.Time)
		log.Printf(" * ")
		log.Println(` ********************************************************************************************************************************************** `)

		if slen, err := strconv.Atoi(s.Num); err == nil {
			sum += slen
		}
	}

	// 总耗时
	takeTime := time.Since(cache.StartTime).Minutes()

	// 打印总结报告
	log.Println(` ********************************************************************************************************************************************** `)
	log.Printf(" * ")
	reporter.Log.Printf(" *                               —— 本次抓取合计 %v 条数据，下载页面 %v 个，耗时：%.5f 分钟 ——", sum, cache.ReqSum, takeTime)
	log.Printf(" * ")
	log.Println(` ********************************************************************************************************************************************** `)

	if cache.Task.RunMode == status.OFFLINE {
		// 按钮状态控制
		toggleRunBtn.SetEnabled(true)
		toggleRunBtn.SetText("开始运行")
	}
}

//中途终止任务
func Stop() {
	status.Crawl = status.STOP
	Pholcus.Crawls.Stop()
	scheduler.Sdl.Stop()
	reporter.Log.Stop()

	// 总耗时
	takeTime := time.Since(cache.StartTime).Minutes()

	// 打印总结报告
	log.Println(` ********************************************************************************************************************************************** `)
	log.Printf(" * ")
	log.Printf(" *                               ！！任务取消：下载页面 %v 个，耗时：%.5f 分钟！！", cache.ReqSum, takeTime)
	log.Printf(" * ")
	log.Println(` ********************************************************************************************************************************************** `)

	// 按钮状态控制
	toggleRunBtn.SetEnabled(true)
	toggleRunBtn.SetText("开始运行")
}
