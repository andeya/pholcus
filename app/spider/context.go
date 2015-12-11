package spider

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
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

	self.Spider.ReqmatrixPush(req)
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
// item类型为map[int]interface{}时，根据ruleName现有的ItemFields字段进行输出，
// item类型为map[string]interface{}时，ruleName不存在的ItemFields字段将被自动添加，
// ruleName为空时默认当前规则。
func (self *Context) Output(item interface{}, ruleName ...string) *Context {
	_ruleName, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用Output()时，指定的规则名不存在！", self.Spider.GetName())
		return self
	}

	self.Response.SetRuleName(_ruleName)
	switch item2 := item.(type) {
	case map[int]interface{}:
		self.Response.AddItem(self.CreatItem(item2, _ruleName))
	case map[string]interface{}:
		for k := range item2 {
			self.Spider.UpsertItemField(rule, k)
		}
		self.Response.AddItem(item2)
	}
	return self
}

// 主动错误处理(如更正历史记录)
func (self *Context) Fatal(v interface{}) {
	panic(v)
}

// 输出文件。
// name指定文件名，为空时默认保持原文件名不变。
func (self *Context) FileOutput(name ...string) *Context {
	self.Response.AddFile(name...)
	return self
}

func (self *Context) GetItemFields(ruleName ...string) []string {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用GetItemFields()时，指定的规则名不存在！", self.Spider.GetName())
		return nil
	}
	return self.Spider.GetItemFields(rule)
}

// 返回结果字段名的值，不存在时返回空字符串
// ruleName为空时默认当前规则。
func (self *Context) GetItemField(index int, ruleName ...string) (field string) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用GetItemField()时，指定的规则名不存在！", self.Spider.GetName())
		return
	}
	return self.Spider.GetItemField(rule, index)
}

// 返回结果字段名的索引，不存在时索引为-1
// ruleName为空时默认当前规则。
func (self *Context) GetItemFieldIndex(field string, ruleName ...string) (index int) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用GetItemField()时，指定的规则名不存在！", self.Spider.GetName())
		return
	}
	return self.Spider.GetItemFieldIndex(rule, field)
}

// 为指定Rule动态追加结果字段名，并返回索引位置
// 已存在时返回原来索引位置
// ruleName为空时默认当前规则。
func (self *Context) UpsertItemField(field string, ruleName ...string) (index int) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用UpsertItemField()时，指定的规则名不存在！", self.Spider.GetName())
		return
	}
	return self.Spider.UpsertItemField(rule, field)
}

// 生成文本结果。
// 用ruleName指定匹配的ItemFields字段，为空时默认当前规则。
func (self *Context) CreatItem(item map[int]interface{}, ruleName ...string) map[string]interface{} {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用CreatItem()时，指定的规则名不存在！", self.Spider.GetName())
		return nil
	}

	var item2 = make(map[string]interface{}, len(item))
	for k, v := range item {
		field := self.Spider.GetItemField(rule, k)
		item2[field] = v
	}
	return item2
}

// 调用指定Rule下辅助函数AidFunc()。
// 用ruleName指定匹配的AidFunc，为空时默认当前规则。
func (self *Context) Aid(aid map[string]interface{}, ruleName ...string) interface{} {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用Aid()时，指定的规则名不存在！", self.Spider.GetName())
		return nil
	}

	return rule.AidFunc(self, aid)
}

// 解析响应流。
// 用ruleName指定匹配的ParseFunc字段，为空时默认调用Root()。
func (self *Context) Parse(ruleName ...string) *Context {
	_ruleName, rule, found := self.getRule(ruleName...)
	self.Response.SetRuleName(_ruleName)
	if !found {
		self.Spider.RuleTree.Root(self)
		return self
	}

	rule.ParseFunc(self)
	return self
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
func (self *Context) GetRule(ruleName string) (*Rule, bool) {
	return self.Spider.GetRule(ruleName)
}

// 自定义暂停时间 pause[0]~(pause[0]+pause[1])，优先级高于外部传参。
// 当且仅当runtime[0]为true时可覆盖现有值。
func (self *Context) SetPausetime(pause [2]uint, runtime ...bool) *Context {
	self.Spider.SetPausetime(pause, runtime...)
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

// 自定义设置输出结果的"当前链接"字段
func (self *Context) SetItemUrl(itemUrl string) *Context {
	self.Response.SetUrl(itemUrl)
	return self
}

func (self *Context) GetMethod() string {
	return self.Response.GetMethod()
}

func (self *Context) GetRequestHeader() http.Header {
	return self.Response.GetRequestHeader()
}

func (self *Context) GetResponseHeader() http.Header {
	return self.Response.GetResponseHeader()
}

func (self *Context) GetReferer() string {
	return self.Response.GetReferer()
}

// 自定义设置输出结果的"上级链接"字段
func (self *Context) SetItemReferer(referer string) *Context {
	self.Response.SetReferer(referer)
	return self
}

func (self *Context) GetRuleName() string {
	return self.Response.GetRuleName()
}

// 返回指定缓存数据，
// 注：为性能考虑，该方法并不保证引用或指针类型的value值被copy，请自行实现
func (self *Context) GetTemp(key string) interface{} {
	return self.Request.Temp[key]
}

// 返回全部缓存数据，
// 当Request会被复用时，为保证缓存数据的独立性，isCopy应该为true
// 注：为性能考虑，isCopy为true并不保证引用或指针类型的value值被copy，请自行实现
func (self *Context) GetTemps(isCopy bool) map[string]interface{} {
	if !isCopy {
		return self.Request.Temp
	}
	copyMap := make(map[string]interface{})
	for k, v := range self.Request.Temp {
		copyMap[k] = v
	}
	return copyMap
}

func (self *Context) SetReqUrl(u string) *Context {
	self.Request.SetUrl(u)
	return self
}

func (self *Context) SetReqmethod(method string) *Context {
	self.Request.SetMethod(method)
	return self
}

func (self *Context) ClearReqHeader() *Context {
	self.Request.Header = make(http.Header)
	return self
}

func (self *Context) SetReqHeader(key, value string) *Context {
	self.Request.Header.Set(key, value)
	return self
}

func (self *Context) AddRequestHeader(key, value string) *Context {
	self.Request.Header.Add(key, value)
	return self
}

func (self *Context) SetReqReferer(referer string) *Context {
	self.Request.SetReferer(referer)
	return self
}

func (self *Context) SetReqProxy(proxy string) *Context {
	self.Request.Proxy = proxy
	return self
}

func (self *Context) SetReqRuleName(ruleName string) *Context {
	self.Request.Rule = ruleName
	return self
}

func (self *Context) SetReqPriority(priority int) *Context {
	self.Request.Priority = priority
	return self
}

func (self *Context) SetReqDownloaderID(id int) *Context {
	self.Request.DownloaderID = id
	return self
}

func (self *Context) SetReqReloadable(can bool) *Context {
	self.Request.Reloadable = can
	return self
}

func (self *Context) SetReqTemp(key string, value interface{}) *Context {
	self.Request.Temp[key] = value
	return self
}

func (self *Context) SetReqTemps(temp map[string]interface{}) *Context {
	self.Request.Temp = temp
	return self
}

// 获取规则
func (self *Context) getRule(ruleName ...string) (name string, rule *Rule, found bool) {
	if len(ruleName) == 0 {
		if self.Response == nil {
			return
		}
		name = self.Response.GetRuleName()
	} else {
		name = ruleName[0]
	}
	rule, found = self.Spider.GetRule(name)
	return
}
