package spider

import (
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/logs"
)

// 规则中的上下文。
type Context struct {
	spider   *Spider
	request  *context.Request
	response *context.Response
	sync.Once
}

func NewContext(sp *Spider, resp *context.Response) *Context {
	if resp == nil {
		return &Context{
			spider:   sp,
			response: resp,
		}
	}
	return &Context{
		spider:   sp,
		response: resp,
	}
}

// 返回响应流
func (self *Context) GetResponse() *context.Response {
	return self.response
}

// 返回原始请求
func (self *Context) GetRequestOriginal() *context.Request {
	return self.response.GetRequest()
}

// 返回请求副本
func (self *Context) GetRequestCopy() *context.Request {
	self.copyRequest()
	return self.request
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

// 将上次原始请求或其副本添加至队列
func (self *Context) AddQueue2() *Context {
	if self.request == nil {
		self.AddQueue(self.response.GetRequest())
	} else {
		self.AddQueue(self.request)
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
		logs.Log.Error("蜘蛛 %s 调用Output()时，指定的规则名不存在！", self.spider.GetName())
		return self
	}

	self.response.SetRuleName(_ruleName)
	switch item2 := item.(type) {
	case map[int]interface{}:
		self.response.AddItem(self.CreatItem(item2, _ruleName))
	case context.Temp:
		for k := range item2 {
			self.spider.UpsertItemField(rule, k)
		}
		self.response.AddItem(item2)
	case map[string]interface{}:
		for k := range item2 {
			self.spider.UpsertItemField(rule, k)
		}
		self.response.AddItem(item2)
	}
	return self
}

// 输出文件。
// name指定文件名，为空时默认保持原文件名不变。
func (self *Context) FileOutput(name ...string) *Context {
	self.response.AddFile(name...)
	return self
}

func (self *Context) GetItemFields(ruleName ...string) []string {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用GetItemFields()时，指定的规则名不存在！", self.spider.GetName())
		return nil
	}
	return self.spider.GetItemFields(rule)
}

// 返回结果字段名的值，不存在时返回空字符串
// ruleName为空时默认当前规则。
func (self *Context) GetItemField(index int, ruleName ...string) (field string) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用GetItemField()时，指定的规则名不存在！", self.spider.GetName())
		return
	}
	return self.spider.GetItemField(rule, index)
}

// 返回结果字段名的索引，不存在时索引为-1
// ruleName为空时默认当前规则。
func (self *Context) GetItemFieldIndex(field string, ruleName ...string) (index int) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用GetItemField()时，指定的规则名不存在！", self.spider.GetName())
		return
	}
	return self.spider.GetItemFieldIndex(rule, field)
}

// 为指定Rule动态追加结果字段名，并返回索引位置
// 已存在时返回原来索引位置
// ruleName为空时默认当前规则。
func (self *Context) UpsertItemField(field string, ruleName ...string) (index int) {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用UpsertItemField()时，指定的规则名不存在！", self.spider.GetName())
		return
	}
	return self.spider.UpsertItemField(rule, field)
}

// 生成文本结果。
// 用ruleName指定匹配的ItemFields字段，为空时默认当前规则。
func (self *Context) CreatItem(item map[int]interface{}, ruleName ...string) map[string]interface{} {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用CreatItem()时，指定的规则名不存在！", self.spider.GetName())
		return nil
	}

	var item2 = make(map[string]interface{}, len(item))
	for k, v := range item {
		field := self.spider.GetItemField(rule, k)
		item2[field] = v
	}
	return item2
}

// 调用指定Rule下辅助函数AidFunc()。
// 用ruleName指定匹配的AidFunc，为空时默认当前规则。
func (self *Context) Aid(aid map[string]interface{}, ruleName ...string) interface{} {
	_, rule, found := self.getRule(ruleName...)
	if !found {
		logs.Log.Error("蜘蛛 %s 调用Aid()时，指定的规则名不存在！", self.spider.GetName())
		return nil
	}

	return rule.AidFunc(self, aid)
}

// 解析响应流。
// 用ruleName指定匹配的ParseFunc字段，为空时默认调用Root()。
func (self *Context) Parse(ruleName ...string) *Context {
	_ruleName, rule, found := self.getRule(ruleName...)
	self.response.SetRuleName(_ruleName)
	if !found {
		self.spider.RuleTree.Root(self)
		return self
	}
	rule.ParseFunc(self)
	return self
}

// 获取蜘蛛名称。
func (self *Context) GetName() string {
	return self.spider.GetName()
}

// 获取蜘蛛描述。
func (self *Context) GetDescription() string {
	return self.spider.GetDescription()
}

// 获取蜘蛛ID。
func (self *Context) GetId() int {
	return self.spider.GetId()
}

// 获取自定义输入。
func (self *Context) GetKeyword() string {
	return self.spider.GetKeyword()
}

// 设置自定义输入。
func (self *Context) SetKeyword(keyword string) *Context {
	self.spider.SetKeyword(keyword)
	return self
}

// 获取采集的最大页数。
func (self *Context) GetMaxPage() int {
	return int(self.spider.GetMaxPage())
}

// 设置采集的最大页数。
func (self *Context) SetMaxPage(max int) *Context {
	self.spider.SetMaxPage(int64(max))
	return self
}

// 返回规则树。
func (self *Context) GetRules() map[string]*Rule {
	return self.spider.GetRules()
}

// 返回指定规则。
func (self *Context) GetRule(ruleName string) (*Rule, bool) {
	return self.spider.GetRule(ruleName)
}

// 自定义暂停区间(随机: Pausetime/2 ~ Pausetime*2)，优先级高于外部传参。
// 当且仅当runtime[0]为true时可覆盖现有值。
func (self *Context) SetPausetime(pause int64, runtime ...bool) *Context {
	self.spider.SetPausetime(pause, runtime...)
	return self
}

// GetBodyStr returns plain string crawled.
func (self *Context) GetText() string {
	return self.response.GetText()
}

// GetHtmlParser returns goquery object binded to target crawl result.
func (self *Context) GetDom() *goquery.Document {
	return self.response.GetDom()
}

func (self *Context) GetUrl() string {
	return self.response.GetUrl() // 与self.request.GetUrl()完全相等
}

// 自定义设置输出结果的"当前链接"字段
func (self *Context) SetItemUrl(itemUrl string) *Context {
	self.response.SetUrl(itemUrl)
	return self
}

func (self *Context) GetMethod() string {
	return self.response.GetMethod()
}

func (self *Context) GetReqHeader() http.Header {
	return self.response.GetRequestHeader()
}

func (self *Context) GetRespHeader() http.Header {
	return self.response.GetResponseHeader()
}

func (self *Context) GetReferer() string {
	return self.response.GetReferer()
}

func (self *Context) GetHost() string {
	return self.response.GetHost()
}

// 自定义设置输出结果的"上级链接"字段
func (self *Context) SetItemReferer(referer string) *Context {
	self.response.SetReferer(referer)
	return self
}

func (self *Context) GetRuleName() string {
	return self.response.GetRuleName()
}

// 返回请求中指定缓存数据
// 强烈建议数据接收者receive为指针类型
// receive为空时，直接输出字符串
func (self *Context) GetTemp(key string, receive ...interface{}) interface{} {
	return self.response.GetTemp(key, receive...)
}

func (self *Context) SetOriginTemp(key string, value interface{}) *Context {
	self.response.SetTemp(key, value)
	return self
}

// 返回请求中缓存数据副本
func (self *Context) CopyTemps() context.Temp {
	temps := make(context.Temp)
	for k, v := range self.response.GetTemps() {
		temps[k] = v
	}
	return temps
}

func (self *Context) SetCopyUrl(u string) *Context {
	self.copyRequest()
	self.request.SetUrl(u)
	return self
}

func (self *Context) SetCopyMethod(method string) *Context {
	self.copyRequest()
	self.request.SetMethod(method)
	return self
}

func (self *Context) EmptyCopyHeader() *Context {
	self.copyRequest()
	self.request.Header = make(http.Header)
	return self
}

func (self *Context) SetCopyHeader(key, value string) *Context {
	self.copyRequest()
	self.request.Header.Set(key, value)
	return self
}

func (self *Context) AddCopyHeader(key, value string) *Context {
	self.copyRequest()
	self.request.Header.Add(key, value)
	return self
}

func (self *Context) SetCopyReferer(referer string) *Context {
	self.copyRequest()
	self.request.SetReferer(referer)
	return self
}

func (self *Context) SetCopyProxy(proxy string) *Context {
	self.copyRequest()
	self.request.Proxy = proxy
	return self
}

func (self *Context) SetCopyRuleName(ruleName string) *Context {
	self.copyRequest()
	self.request.Rule = ruleName
	return self
}

func (self *Context) SetCopyPriority(priority int) *Context {
	self.copyRequest()
	self.request.Priority = priority
	return self
}

func (self *Context) SetCopyDownloader(id int) *Context {
	self.copyRequest()
	self.request.DownloaderID = id
	return self
}

func (self *Context) SetCopyReloadable(can bool) *Context {
	self.copyRequest()
	self.request.Reloadable = can
	return self
}

func (self *Context) SetCopyTemp(key string, value interface{}) *Context {
	self.copyRequest()
	self.request.SetTemp(key, value)
	return self
}

func (self *Context) SetCopyTemps(temp map[string]interface{}) *Context {
	self.copyRequest()
	self.request.SetTemps(temp)
	return self
}

// 获取规则
func (self *Context) getRule(ruleName ...string) (name string, rule *Rule, found bool) {
	if len(ruleName) == 0 {
		if self.response == nil {
			return
		}
		name = self.response.GetRuleName()
	} else {
		name = ruleName[0]
	}
	rule, found = self.spider.GetRule(name)
	return
}

func (self *Context) copyRequest() {
	self.Once.Do(func() { self.request = self.GetRequestOriginal().Copy() })
}

func (self *Context) getRequest() *context.Request {
	if self.request != nil {
		return self.request
	}
	return self.response.GetRequest()
}
