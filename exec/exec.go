package exec

import (
	"flag"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/henrylee2cn/pholcus/app"
	"github.com/henrylee2cn/pholcus/cmd"
	"github.com/henrylee2cn/pholcus/common/gc"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"github.com/henrylee2cn/pholcus/web"
)

var (
	uiflag             *string
	modeflag           *int
	portflag           *int
	masterflag         *string
	keyinsflag         *string
	limitflag          *int64
	outputflag         *string
	threadflag         *int
	pauseflag          *int64
	proxyflag          *int64
	dockerflag         *int
	successInheritflag *bool
	failureInheritflag *bool
)

func init() {
	// 开启最大核心数运行
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 开启手动GC
	gc.ManualGC()
}

func DefaultRun(uiDefault string) {
	fmt.Printf("%v\n\n", config.FULL_NAME)
	flag.String("a *********************************************** common *********************************************** -a", "", "")
	// 操作界面
	uiflag = flag.String("_ui", uiDefault, "   <选择操作界面> [web] [gui] [cmd]")
	flagCommon()
	web.Flag()
	cmd.Flag()
	flag.String("z", "", "README:   参数设置参考 [xxx] 提示，参数中包含多个值时以 \",\" 间隔。\r\n")
	flag.Parse()
	writeFlag()
	run(*uiflag)
}

func flagCommon() {
	//运行模式
	modeflag = flag.Int(
		"a_mode",
		cache.Task.Mode,
		"   <运行模式: ["+strconv.Itoa(status.OFFLINE)+"] 单机    ["+strconv.Itoa(status.SERVER)+"] 服务端    ["+strconv.Itoa(status.CLIENT)+"] 客户端>")

	//端口号，非单机模式填写
	portflag = flag.Int(
		"a_port",
		cache.Task.Port,
		"   <端口号: 只填写数字即可，不含冒号，单机模式不填>")

	//主节点ip，客户端模式填写
	masterflag = flag.String(
		"a_master",
		cache.Task.Master,
		"   <服务端IP: 不含端口，客户端模式下使用>")

	// 自定义配置
	keyinsflag = flag.String(
		"a_keyins",
		cache.Task.Keyins,
		"   <自定义配置: 多任务请分别多包一层“<>”>")

	// 采集上限
	limitflag = flag.Int64(
		"a_limit",
		cache.Task.Limit,
		"   <采集上限（默认限制URL数）> [>=0]")

	// 输出方式
	outputflag = flag.String(
		"a_outtype",
		cache.Task.OutType,
		func() string {
			var outputlib string
			for _, v := range app.LogicApp.GetOutputLib() {
				outputlib += "[" + v + "] "
			}
			return "   <输出方式: > " + strings.TrimRight(outputlib, " ")
		}())

	// 并发协程数
	threadflag = flag.Int(
		"a_thread",
		cache.Task.ThreadNum,
		"   <并发协程> [1~99999]")

	// 平均暂停时间
	pauseflag = flag.Int64(
		"a_pause",
		cache.Task.Pausetime,
		"   <平均暂停时间/ms> [>=100]")

	// 代理IP更换频率
	proxyflag = flag.Int64(
		"a_proxyminute",
		cache.Task.ProxyMinute,
		"   <代理IP更换频率: /m，为0时不使用代理> [>=0]")

	// 分批输出
	dockerflag = flag.Int(
		"a_dockercap",
		cache.Task.DockerCap,
		"   <分批输出> [1~5000000]")

	// 继承历史成功记录
	successInheritflag = flag.Bool(
		"a_success",
		cache.Task.SuccessInherit,
		"   <继承并保存成功记录> [true] [false]")

	// 继承历史失败记录
	failureInheritflag = flag.Bool(
		"a_failure",
		cache.Task.FailureInherit,
		"   <继承并保存失败记录> [true] [false]")
}

func writeFlag() {
	cache.Task.Mode = *modeflag
	cache.Task.Port = *portflag
	cache.Task.Master = *masterflag
	cache.Task.Keyins = *keyinsflag
	cache.Task.Limit = *limitflag
	cache.Task.OutType = *outputflag
	cache.Task.ThreadNum = *threadflag
	cache.Task.Pausetime = *pauseflag
	cache.Task.ProxyMinute = *proxyflag
	cache.Task.DockerCap = *dockerflag
	cache.Task.SuccessInherit = *successInheritflag
	cache.Task.FailureInherit = *failureInheritflag
}
