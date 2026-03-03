// Package cmd implements the command-line interface for Pholcus.
package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/cache"
	"github.com/andeya/pholcus/runtime/status"
)

var (
	spiderflag *string
)

// Flag registers command-line flags for the CMD interface.
func Flag() {
	flag.String("c ******************************************** only for cmd ******************************************** -c", "", "")

	spiderflag = flag.String(
		"c_spider",
		"",
		func() string {
			var spiderlist string
			for k, v := range app.LogicApp.GetSpiderLib() {
				spiderlist += "   [" + strconv.Itoa(k) + "] " + v.GetName() + "  " + v.GetDescription() + "\r\n"
			}
			return "   <Spider list: separate multiple spiders with \",\">\r\n" + spiderlist
		}())

	flag.String(
		"c_z",
		"",
		"CMD-EXAMPLE: $ pholcus -_ui=cmd -a_mode="+strconv.Itoa(status.OFFLINE)+" -c_spider=3,8 -a_outtype=csv -a_thread=20 -a_batchcap=5000 -a_pause=300 -a_proxyminute=0 -a_keyins=\"<pholcus><golang>\" -a_limit=10 -a_success=true -a_failure=true\n",
	)
}

// Run starts the application in the configured mode.
func Run() {
	app.LogicApp.Init(cache.Task.Mode, cache.Task.Port, cache.Task.Master)
	if cache.Task.Mode == status.UNSET {
		return
	}
	switch app.LogicApp.GetAppConf("Mode").(int) {
	case status.SERVER:
		for {
			parseInput()
			run()
		}
	case status.CLIENT:
		run()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	default:
		run()
	}
}

func run() {
	sps := []*spider.Spider{}
	*spiderflag = strings.TrimSpace(*spiderflag)
	if *spiderflag == "*" {
		sps = app.LogicApp.GetSpiderLib()

	} else {
		for _, idx := range strings.Split(*spiderflag, ",") {
			idx = strings.TrimSpace(idx)
			if idx == "" {
				continue
			}
			i, _ := strconv.Atoi(idx)
			sps = append(sps, app.LogicApp.GetSpiderLib()[i])
		}
	}

	app.LogicApp.SpiderPrepare(sps).Run()
}

// parseInput reads task parameters from stdin in server mode.
func parseInput() {
	logs.Log().Informational("\nRequired task parameter: %v\nOptional task parameters: %v\n", "-c_spider", []string{
		"-a_keyins",
		"-a_limit",
		"-a_outtype",
		"-a_thread",
		"-a_pause",
		"-a_proxyminute",
		"-a_batchcap",
		"-a_success",
		"-a_failure"})
	logs.Log().Informational("\nAdd task:\n")
retry:
	*spiderflag = ""
	input := [12]string{}
	fmt.Scanln(&input[0], &input[1], &input[2], &input[3], &input[4], &input[5], &input[6], &input[7], &input[8], &input[9])
	if strings.Index(input[0], "=") < 4 {
		logs.Log().Informational("\nInvalid task parameters, please re-enter:")
		goto retry
	}
	for _, v := range input {
		i := strings.Index(v, "=")
		if i < 4 {
			continue
		}
		key, value := v[:i], v[i+1:]
		switch key {
		case "-a_keyins":
			cache.Task.Keyins = value
		case "-a_limit":
			limit, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			cache.Task.Limit = limit
		case "-a_outtype":
			cache.Task.OutType = value
		case "-a_thread":
			thread, err := strconv.Atoi(value)
			if err != nil {
				break
			}
			cache.Task.ThreadNum = thread
		case "-a_pause":
			pause, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			cache.Task.Pausetime = pause
		case "-a_proxyminute":
			proxyminute, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				break
			}
			cache.Task.ProxyMinute = proxyminute
		case "-a_batchcap":
			batchcap, err := strconv.Atoi(value)
			if err != nil {
				break
			}
			if batchcap < 1 {
				batchcap = 1
			}
			cache.Task.BatchCap = batchcap
		case "-a_success":
			if value == "true" {
				cache.Task.SuccessInherit = true
			} else if value == "false" {
				cache.Task.SuccessInherit = false
			}
		case "-a_failure":
			if value == "true" {
				cache.Task.FailureInherit = true
			} else if value == "false" {
				cache.Task.FailureInherit = false
			}
		case "-c_spider":
			*spiderflag = value
		default:
			logs.Log().Informational("\nUnknown parameter detected. Required: %v\nOptional: %v\n", "-c_spider", []string{
				"-a_keyins",
				"-a_limit",
				"-a_outtype",
				"-a_thread",
				"-a_pause",
				"-a_proxyminute",
				"-a_batchcap",
				"-a_success",
				"-a_failure"})
			goto retry
		}
	}
}
