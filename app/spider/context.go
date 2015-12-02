package spider

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/app/scheduler"
	"github.com/henrylee2cn/pholcus/logs"
)

// 规则中的上下文。
type Context struct {
	Spider   *Spider
	Request  *context.Request
	Response *context.Response
}

func NewContext(sp *Spider, resp *context.Response) *Context {
	if resp == nil {
		return &Context{
			Spider:   sp,
			Response: resp,
		}
	}
	return &Context{
		Spider:   sp,
		Request:  resp.GetRequest(),
		Response: resp,
	}
}

// 生成并添加请求至队列。
// Request.Url与Request.Rule必须设置。
// Request.Spider无需手动设置(由系统自动设置)。
// Request.EnableCookie在Spider字段中统一设置，规则请求中指定的无效。
// 以下字段有默认值，可不设置:
// Request.Method默认为GET方法;
// Request.DialTimeout默认为常量context.DefaultDialTimeout，小于0时不限制等待响应时长;
// Request.ConnTimeout默认为常量context.DefaultConnTimeout，小于0时不限制下载超时;
// Request.TryTimes默认为常量context.DefaultTryTimes，小于0时不限制失败重载次数;
// Request.RedirectTimes默认不限制重定向次数，小于0时可禁止重定向跳转;
// Request.RetryPause默认为常量context.DefaultRetryPause;
// Request.DownloaderID指定下载器ID，0为默认的Surf高并发下载器，功能完备，1为PhantomJS下载器，特点破防力强，速度慢，低并发。
// 默认自动补填Referer。
func (self *Context) AddQueue(req *context.Request) *Context {
	err := req.
		SetSpiderName(self.Spider.GetName()).
		SetSpiderId(self.Spider.GetId()).
		SetEnableCookie(self.Spider.GetEnableCookie()).
		Prepare()

	if err != nil {
		logs.Log.Error("%v", err)
		return self
	}

	if req.GetReferer() == "" && self.Response != nil {
		req.SetReferer(self.Response.GetUrl())
	}

	scheduler.Sdl.Push(req)
	return self
}

// 批量url生成请求，并添加至队列。
func (self *Context) BulkAddQueue(urls []string, req *context.Request) *Context {
	for _, url := range urls {
		req.SetUrl(url)
		self.AddQueue(req)
	}
	return self
}

// 输出文本结果。
// item允许的类型为map[int]interface{}或map[string]interface{}。
// 用ruleName指定匹配的OutFeild字段，为空时默认当前规则。
func (self *Context) Output(item interface{}, ruleName ...string) *Context {
	if len(ruleName) == 0 {
		if self.Request == nil {
			logs.Log.Error("蜘蛛 %s 空响应调用Output()，必须指定规则名！", self.Spider.GetName())
			return self
		}
		ruleName = append(ruleName, self.Request.GetRuleName())
	}

	self.Request.SetRuleName(ruleName[0])
	switch item2 := item.(type) {
	case map[int]interface{}:
		self.Response.AddItem(self.CreatItem(item2, ruleName[0]))
	case map[string]interface{}:
		self.Response.AddItem(item2)
	}
	return self
}

// 输出文件。
// name指定文件名，为空时默认保持原文件名不变。
func (self *Context) FileOutput(name ...string) *Context {
	self.Response.AddFile(name...)
	return self
}

// 生成文本结果。
// 用ruleName指定匹配的OutFeild字段，为空时默认当前规则。
func (self *Context) CreatItem(item map[int]interface{}, ruleName ...string) map[string]interface{} {
	if len(ruleName) == 0 {
		if self.Request == nil {
			logs.Log.Error("蜘蛛 %s 空响应调用CreatItem()，必须指定规则名！", self.Spider.GetName())
			return nil
		}
		ruleName = append(ruleName, self.Request.GetRuleName())
	}
	var item2 = make(map[string]interface{}, len(item))
	for k, v := range item {
		item2[self.Spider.IndexOutFeild(ruleName[0], k)] = v
	}
	return item2
}

// 调用指定Rule下辅助函数AidFunc()。
// 用ruleName指定匹配的AidFunc，为空时默认当前规则。
func (self *Context) Aid(aid map[string]interface{}, ruleName ...string) interface{} {
	if len(ruleName) == 0 {
		if self.Request == nil {
			logs.Log.Error("蜘蛛 %s 空响应调用Aid()，必须指定规则名！", self.Spider.GetName())
			return nil
		}
		ruleName = append(ruleName, self.Request.GetRuleName())
	}
	return self.Spider.GetRule(ruleName[0]).AidFunc(self, aid)
}

// 解析响应流。
// 用ruleName指定匹配的ParseFunc字段，为空时默认调用Root()。
func (self *Context) Parse(ruleName ...string) *Context {
	if len(ruleName) == 0 || ruleName[0] == "" {
		if self.Response != nil {
			self.Request.SetRuleName("")
		}
		self.Spider.RuleTree.Root(self)
		return self
	}
	self.Request.SetRuleName(ruleName[0])
	self.Spider.GetRule(ruleName[0]).ParseFunc(self)
	return self
}

// 返回采集语义字段。
// 用ruleName指定匹配的OutFeild字段，为空时默认当前规则。
func (self *Context) IndexOutFeild(index int, ruleName ...string) (feild string) {
	if len(ruleName) == 0 {
		if self.Request == nil {
			logs.Log.Error("蜘蛛 %s 空响应调用IndexOutFeild()，必须指定规则名！", self.Spider.GetName())
			return ""
		}
		ruleName = append(ruleName, self.Request.GetRuleName())
	}
	return self.Spider.IndexOutFeild(ruleName[0], index)
}

// 返回采集语义字段的索引位置，不存在时返回-1。
// 用ruleName指定匹配的OutFeild字段，为空时默认当前规则。
func (self *Context) FindOutFeild(feild string, ruleName ...string) (index int) {
	if len(ruleName) == 0 {
		if self.Request == nil {
			logs.Log.Error("蜘蛛 %s 空响应调用FindOutFeild()，必须指定规则名！", self.Spider.GetName())
			return
		}
		ruleName = append(ruleName, self.Request.GetRuleName())
	}
	return self.Spider.FindOutFeild(ruleName[0], feild)
}

// 为指定Rule动态追加采集语义字段，并返回索引位置。
// 已存在时返回原来索引位置。
// 用ruleName指定匹配的OutFeild字段，为空时默认当前规则。
func (self *Context) AddOutFeild(feild string, ruleName ...string) (index int) {
	if len(ruleName) == 0 {
		if self.Request == nil {
			logs.Log.Error("蜘蛛 %s 空响应调用AddOutFeild()，必须指定规则名！", self.Spider.GetName())
			return
		}
		ruleName = append(ruleName, self.Request.GetRuleName())
	}
	return self.Spider.AddOutFeild(ruleName[0], feild)
}

// 设置代理服务器列表。
func (self *Context) SetProxys(proxys []string) *Context {
	self.Spider.SetProxys(proxys)
	return self
}

// 添加代理服务器。
func (self *Context) AddProxys(proxy ...string) *Context {
	self.Spider.AddProxys(proxy...)
	return self
}

// 获取代理服务器列表。
func (self *Context) GetProxys() []string {
	return self.Spider.GetProxys()
}

// 获取下一个代理服务器。
func (self *Context) GetOneProxy() string {
	return self.Spider.GetOneProxy()
}

// 获取蜘蛛名称。
func (self *Context) GetName() string {
	return self.Spider.GetName()
}

// 获取蜘蛛描述。
func (self *Context) GetDescription() string {
	return self.Spider.GetDescription()
}

// 获取蜘蛛ID。
func (self *Context) GetId() int {
	return self.Spider.GetId()
}

// 获取自定义输入。
func (self *Context) GetKeyword() string {
	return self.Spider.GetKeyword()
}

// 设置自定义输入。
func (self *Context) SetKeyword(keyword string) *Context {
	self.Spider.SetKeyword(keyword)
	return self
}

// 获取采集的最大页数。
func (self *Context) GetMaxPage() int {
	return self.Spider.GetMaxPage()
}

// 设置采集的最大页数。
func (self *Context) SetMaxPage(max int) *Context {
	self.Spider.SetMaxPage(max)
	return self
}

// 返回规则树。
func (self *Context) GetRules() map[string]*Rule {
	return self.Spider.GetRules()
}

// 返回指定规则。
func (self *Context) GetRule(ruleName string) *Rule {
	return self.Spider.GetRule(ruleName)
}

// 自定义暂停时间 pause[0]~(pause[0]+pause[1])，优先级高于外部传参。
// 当且仅当runtime[0]为true时可覆盖现有值。
func (self *Context) SetPausetime(pause [2]uint, runtime ...bool) *Context {
	self.Spider.SetPausetime(pause, runtime...)
	return self
}

func (self *Context) SetText(s string) *Context {
	self.Response.SetText(s)
	return self
}

// GetBodyStr returns plain string crawled.
func (self *Context) GetText() string {
	return self.Response.GetText()
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Context) GetDom() *goquery.Document {
	return self.Response.GetDom()
}

// GetRequest returns request oject of self page.
func (self *Context) GetRequest() *context.Request {
	return self.Request
}

// GetResponse returns response oject of self page.
func (self *Context) GetResponse() *context.Response {
	return self.Response
}

func (self *Context) GetUrl() string {
	return self.Response.GetUrl() // 与self.Request.GetUrl()完全相等
}

func (self *Context) GetMethod() string {
	return self.Request.GetMethod()
}

func (self *Context) GetReferer() string {
	return self.Request.GetReferer()
}

func (self *Context) GetRuleName() string {
	return self.Request.GetRuleName()
}

func (self *Context) GetTemp(key string) interface{} {
	return self.Request.GetTemp(key)
}

func (self *Context) GetTemps() map[string]interface{} {
	return self.Request.GetTemps()
}

func (self *Context) SetRequestUrl(u string) *Context {
	self.Request.SetUrl(u)
	return self
}

func (self *Context) SetRequestMethod(method string) *Context {
	self.Request.SetMethod(method)
	return self
}

func (self *Context) SetRequestReferer(referer string) *Context {
	self.Request.SetReferer(referer)
	return self
}

func (self *Context) SetRequestRuleName(ruleName string) *Context {
	self.Request.SetRuleName(ruleName)
	return self
}

func (self *Context) SetRequestTemp(key string, value interface{}) *Context {
	self.Request.SetTemp(key, value)
	return self
}

func (self *Context) SetRequestTemps(temp map[string]interface{}) *Context {
	self.Request.SetAllTemps(temp)
	return self
}
