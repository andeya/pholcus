// Package exec provides entry points to launch CMD or Web interface based on run mode.
package exec

import (
	"flag"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/cmd"
	"github.com/andeya/pholcus/common/gc"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
	"github.com/andeya/pholcus/web"
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
	batchCapFlag       *int
	successInheritflag *bool
	failureInheritflag *bool
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	gc.ManualGC()
}

// DefaultRun starts the application with the given default UI.
func DefaultRun(uiDefault string) {
	fmt.Printf("%v\n\n", config.FullName)
	flag.String("a *********************************************** common *********************************************** -a", "", "")
	uiflag = flag.String("_ui", uiDefault, "   <Select UI> [web] [gui] [cmd]")
	flagCommon()
	web.Flag()
	cmd.Flag()
	flag.String("z", "", "README:   See [xxx] for parameter settings; separate multiple values with \",\".\r\n")
	flag.Parse()
	writeFlag()
	run(*uiflag)
}

func flagCommon() {
	rc := &config.Conf().Run

	modeflag = flag.Int(
		"a_mode",
		rc.Mode,
		"   <Run mode: ["+strconv.Itoa(status.OFFLINE)+"] Standalone    ["+strconv.Itoa(status.SERVER)+"] Server    ["+strconv.Itoa(status.CLIENT)+"] Client>")

	portflag = flag.Int(
		"a_port",
		rc.Port,
		"   <Port: numbers only, no colon; leave empty for standalone mode>")

	masterflag = flag.String(
		"a_master",
		rc.Master,
		"   <Server IP: no port, for client mode>")

	keyinsflag = flag.String(
		"a_keyins",
		"",
		"   <Custom config: wrap each task in \"<>\" for multiple tasks>")

	limitflag = flag.Int64(
		"a_limit",
		rc.Limit,
		"   <Crawl limit (default URL count)> [>=0]")

	outputflag = flag.String(
		"a_outtype",
		rc.OutType,
		func() string {
			var outputlib string
			for _, v := range app.LogicApp.GetOutputLib() {
				outputlib += "[" + v + "] "
			}
			return "   <Output type: > " + strings.TrimRight(outputlib, " ")
		}())

	threadflag = flag.Int(
		"a_thread",
		rc.ThreadNum,
		"   <Concurrency> [1~99999]")

	pauseflag = flag.Int64(
		"a_pause",
		rc.Pausetime,
		"   <Avg pause time/ms> [>=100]")

	proxyflag = flag.Int64(
		"a_proxyminute",
		rc.ProxyMinute,
		"   <Proxy rotation: /min, 0=no proxy> [>=0]")

	batchCapFlag = flag.Int(
		"a_batchcap",
		rc.BatchCap,
		"   <Batch output capacity> [1~5000000]")

	successInheritflag = flag.Bool(
		"a_success",
		rc.SuccessInherit,
		"   <Inherit success records> [true] [false]")

	failureInheritflag = flag.Bool(
		"a_failure",
		rc.FailureInherit,
		"   <Inherit failure records> [true] [false]")
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
	cache.Task.BatchCap = *batchCapFlag
	cache.Task.SuccessInherit = *successInheritflag
	cache.Task.FailureInherit = *failureInheritflag
}
