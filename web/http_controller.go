package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app"
	"github.com/andeya/pholcus/common/session"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/status"
)

var globalSessions *session.Manager

func init() {
	r := result.Ret(session.NewManager("memory", `{"cookieName":"pholcusSession", "enableSetCookie,omitempty": true, "secure": false, "sessionIDHashFunc": "sha1", "sessionIDHashKey": "", "cookieLifeTime": 157680000, "providerConfig": ""}`))
	if r.IsErr() {
		log.Fatal(r.UnwrapErr())
	}
	globalSessions = r.Unwrap()
	// go globalSessions.GC()
}

func web(rw http.ResponseWriter, req *http.Request) {
	r := globalSessions.SessionStart(rw, req)
	if r.IsErr() {
		logs.Log().Error("session start: %v", r.UnwrapErr())
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
	sess := r.Unwrap()
	defer sess.SessionRelease(rw)
	indexR := result.Ret(viewsFS.ReadFile("views/index.html"))
	if indexR.IsErr() {
		logs.Log().Error("read index.html: %v", indexR.UnwrapErr())
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
	index := indexR.Unwrap()
	tR := result.Ret(template.New("index").Parse(string(index)))
	if tR.IsErr() {
		logs.Log().Error("%v", tR.UnwrapErr())
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}
	t := tR.Unwrap()
	data := map[string]interface{}{
		"title":   config.Name,
		"version": config.Version,
		"author":  config.Author,
		"mode": map[string]int{
			"offline": status.OFFLINE,
			"server":  status.SERVER,
			"client":  status.CLIENT,
			"unset":   status.UNSET,
			"curr":    app.LogicApp.GetAppConf("mode").(int),
		},
		"status": map[string]int{
			"stopped": status.STOPPED,
			"stop":    status.STOP,
			"run":     status.RUN,
			"pause":   status.PAUSE,
		},
		"port": app.LogicApp.GetAppConf("port").(int),
		"ip":   app.LogicApp.GetAppConf("master").(string),
	}
	_ = t.Execute(rw, data)
}
