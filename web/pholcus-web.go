// Pholcus is a high-concurrency, distributed, heavyweight crawler written in pure Go.
// It supports standalone, server, and client modes with Web, GUI, and CLI interfaces.
// Web UI package.
package web

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/andeya/gust/iterator"
	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/cache"
)

var (
	ip         *string
	port       *int
	addr       string
	spiderMenu []map[string]string
)

// Flag parses command-line flags for the web server.
func Flag() {
	flag.String("b ******************************************** only for web ******************************************** -b", "", "")
	ip = flag.String("b_ip", "0.0.0.0", "   <Web Server IP>")
	port = flag.Int("b_port", 9090, "   <Web Server Port>")
}

// Run starts the web server and opens the default browser.
func Run() {
	appInit()

	addr = *ip + ":" + strconv.Itoa(*port)

	Router()

	log.Printf("[pholcus] Server running on %v\n", addr)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "http://localhost:"+strconv.Itoa(*port))
	case "darwin":
		cmd = exec.Command("open", "http://localhost:"+strconv.Itoa(*port))
	}
	if cmd != nil {
		go func() {
			log.Println("[pholcus] Open the default browser after two seconds...")
			time.Sleep(time.Second * 2)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}()
	}

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logs.Log.Emergency("ListenAndServe: %v", err)
	}
}

func appInit() {
	app.LogicApp.SetLog(Lsc).SetAppConf("Mode", cache.Task.Mode)

	spiderMenu = iterator.Map(iterator.FromSlice(app.LogicApp.GetSpiderLib()), func(sp *spider.Spider) map[string]string {
		return map[string]string{"name": sp.GetName(), "description": sp.GetDescription()}
	}).Collect()
}
