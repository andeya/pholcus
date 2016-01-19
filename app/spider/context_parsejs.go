package spider

import (
	"time"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/logs"
)

// 用于动态规则添加请求
func (self *Context) JsAddQueue(jreq map[string]interface{}) *Context {
	req := &context.Request{}
	u, ok := jreq["Url"].(string)
	if !ok {
		return self
	}
	req.Url = u
	req.Rule, _ = jreq["Rule"].(string)
	req.Method, _ = jreq["Method"].(string)
	req.Header, _ = jreq["Header"].(map[string][]string)
	req.PostData, _ = jreq["PostData"].(string)
	req.Reloadable, _ = jreq["Reloadable"].(bool)
	if t, ok := jreq["DialTimeout"].(int64); ok {
		req.DialTimeout = time.Duration(t)
	}
	if t, ok := jreq["ConnTimeout"].(int64); ok {
		req.ConnTimeout = time.Duration(t)
	}
	if t, ok := jreq["RetryPause"].(int64); ok {
		req.RetryPause = time.Duration(t)
	}
	if t, ok := jreq["TryTimes"].(int64); ok {
		req.TryTimes = int(t)
	}
	if t, ok := jreq["RedirectTimes"].(int64); ok {
		req.RedirectTimes = int(t)
	}
	if t, ok := jreq["Priority"].(int64); ok {
		req.Priority = int(t)
	}
	if t, ok := jreq["DownloaderID"].(int64); ok {
		req.DownloaderID = int(t)
	}
	if t, ok := jreq["Temp"].(map[string]interface{}); ok {
		req.Temp = t
	}

	err := req.
		SetSpiderName(self.spider.GetName()).
		SetSpiderId(self.spider.GetId()).
		SetEnableCookie(self.spider.GetEnableCookie()).
		Prepare()

	if err != nil {
		logs.Log.Error("%v", err)
		return self
	}

	if req.GetReferer() == "" && self.response != nil {
		req.SetReferer(self.response.GetUrl())
	}

	self.spider.ReqmatrixPush(req)
	return self
}
