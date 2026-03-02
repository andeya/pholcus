package spider

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/app/scheduler"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/status"
)

var ErrForcedStop = errors.New("forced stop")

const (
	KEYIN       = util.USE_KEYIN // rules that use Spider.Keyin must set its initial value to USE_KEYIN
	LIMIT       = math.MaxInt64  // rules that customize Limit must set its initial value to LIMIT
	FORCED_STOP = "——主动终止Spider——"
)

type (
	// Spider defines a crawl spider with its rules and runtime state.
	Spider struct {
		// User-defined fields
		Name            string                                                     // display name (must be unique)
		Description     string                                                     // display description
		Pausetime       int64                                                      // random pause range (50%~200%); if set in rule, overrides UI parameter
		Limit           int64                                                      // request limit (0 = unlimited; set to LIMIT for custom limit logic in rules)
		Keyin           string                                                     // custom input config (set to KEYIN in rules to enable)
		EnableCookie    bool                                                       // whether requests carry cookies
		NotDefaultField bool                                                       // disable default output fields Url/ParentUrl/DownloadTime
		Namespace       func(self *Spider) string                                  // namespace for output file/path naming
		SubNamespace    func(self *Spider, dataCell map[string]interface{}) string // sub-namespace, may depend on specific data content
		RuleTree        *RuleTree                                                  // crawl rule tree

		// System-assigned fields
		id        int
		subName   string            // secondary identifier derived from Keyin
		reqMatrix *scheduler.Matrix // request scheduling matrix
		timer     *Timer
		status    int
		lock      sync.RWMutex
		once      sync.Once
	}
	// RuleTree defines the crawl rule tree.
	RuleTree struct {
		Root  func(*Context)   // entry point
		Trunk map[string]*Rule // rule map (keyed by rule name)
	}
	// Rule defines a single crawl rule node.
	Rule struct {
		ItemFields []string                                           // result field names (optional; preserves field order)
		ParseFunc  func(*Context)                                     // content parsing function
		AidFunc    func(*Context, map[string]interface{}) interface{} // auxiliary helper function
	}
)

// Register adds this spider to the global species list.
func (self *Spider) Register() *Spider {
	self.status = status.STOPPED
	return Species.Add(self)
}

// GetItemFields returns the result field names for the given rule.
func (self *Spider) GetItemFields(rule *Rule) []string {
	return rule.ItemFields
}

// GetItemField returns the field name at the given index, or "" if out of range.
func (self *Spider) GetItemField(rule *Rule, index int) (field string) {
	if index > len(rule.ItemFields)-1 || index < 0 {
		return ""
	}
	return rule.ItemFields[index]
}

// GetItemFieldIndex returns the index of the given field name, or -1 if not found.
func (self *Spider) GetItemFieldIndex(rule *Rule, field string) (index int) {
	for idx, v := range rule.ItemFields {
		if v == field {
			return idx
		}
	}
	return -1
}

// UpsertItemField appends a result field name to the rule and returns its index.
// If the field already exists, the existing index is returned.
func (self *Spider) UpsertItemField(rule *Rule, field string) (index int) {
	for i, v := range rule.ItemFields {
		if v == field {
			return i
		}
	}
	rule.ItemFields = append(rule.ItemFields, field)
	return len(rule.ItemFields) - 1
}

// GetName returns the spider name.
func (self *Spider) GetName() string {
	return self.Name
}

// GetSubName returns the secondary identifier derived from Keyin (computed once).
func (self *Spider) GetSubName() string {
	self.once.Do(func() {
		self.subName = self.GetKeyin()
		self.subName = util.MakeHash(self.subName)
	})
	return self.subName
}

// GetRule returns the rule with the given name.
func (self *Spider) GetRule(ruleName string) *Rule {
	rule, ok := self.RuleTree.Trunk[ruleName]
	if !ok {
		return nil
	}
	return rule
}

// MustGetRule returns the rule with the given name (panics if missing).
func (self *Spider) MustGetRule(ruleName string) *Rule {
	rule := self.GetRule(ruleName)
	if rule == nil {
		panic("spider: rule not found: " + ruleName)
	}
	return rule
}

// GetRules returns the full rule map.
func (self *Spider) GetRules() map[string]*Rule {
	return self.RuleTree.Trunk
}

// GetDescription returns the spider description.
func (self *Spider) GetDescription() string {
	return self.Description
}

// GetId returns the spider's queue index.
func (self *Spider) GetId() int {
	return self.id
}

// SetId assigns the spider's queue index.
func (self *Spider) SetId(id int) {
	self.id = id
}

// GetKeyin returns the custom keyword/configuration input.
func (self *Spider) GetKeyin() string {
	return self.Keyin
}

// SetKeyin sets the custom keyword/configuration input.
func (self *Spider) SetKeyin(keyword string) {
	self.Keyin = keyword
}

// GetLimit returns the crawl limit.
// Negative means request-count limiting; positive means custom rule-based limiting.
func (self *Spider) GetLimit() int64 {
	return self.Limit
}

// SetLimit sets the crawl limit.
func (self *Spider) SetLimit(max int64) {
	self.Limit = max
}

// GetEnableCookie reports whether requests carry cookies.
func (self *Spider) GetEnableCookie() bool {
	return self.EnableCookie
}

// SetPausetime sets a custom pause interval. Only overwrites an existing value when runtime[0] is true.
func (self *Spider) SetPausetime(pause int64, runtime ...bool) {
	if self.Pausetime == 0 || len(runtime) > 0 && runtime[0] {
		self.Pausetime = pause
	}
}

// SetTimer configures a timer identified by id.
// When bell is nil, tol is a countdown sleep duration; otherwise tol specifies the wake-up occurrence.
func (self *Spider) SetTimer(id string, tol time.Duration, bell *Bell) bool {
	if self.timer == nil {
		self.timer = newTimer()
	}
	return self.timer.set(id, tol, bell)
}

// RunTimer starts the timer and reports whether it can continue to be used.
func (self *Spider) RunTimer(id string) bool {
	if self.timer == nil {
		return false
	}
	return self.timer.sleep(id)
}

// Copy returns a deep copy of the spider, including its rule tree.
func (self *Spider) Copy() *Spider {
	ghost := &Spider{}
	ghost.Name = self.Name
	ghost.subName = self.subName

	ghost.RuleTree = &RuleTree{
		Root:  self.RuleTree.Root,
		Trunk: make(map[string]*Rule, len(self.RuleTree.Trunk)),
	}
	for k, v := range self.RuleTree.Trunk {
		ghost.RuleTree.Trunk[k] = new(Rule)

		ghost.RuleTree.Trunk[k].ItemFields = make([]string, len(v.ItemFields))
		copy(ghost.RuleTree.Trunk[k].ItemFields, v.ItemFields)

		ghost.RuleTree.Trunk[k].ParseFunc = v.ParseFunc
		ghost.RuleTree.Trunk[k].AidFunc = v.AidFunc
	}

	ghost.Description = self.Description
	ghost.Pausetime = self.Pausetime
	ghost.EnableCookie = self.EnableCookie
	ghost.Limit = self.Limit
	ghost.Keyin = self.Keyin

	ghost.NotDefaultField = self.NotDefaultField
	ghost.Namespace = self.Namespace
	ghost.SubNamespace = self.SubNamespace

	ghost.timer = self.timer
	ghost.status = self.status

	return ghost
}

// ReqmatrixInit initializes the request scheduling matrix for this spider.
func (self *Spider) ReqmatrixInit() *Spider {
	if self.Limit < 0 {
		self.reqMatrix = scheduler.AddMatrix(self.GetName(), self.GetSubName(), self.Limit)
		self.SetLimit(0)
	} else {
		self.reqMatrix = scheduler.AddMatrix(self.GetName(), self.GetSubName(), math.MinInt64)
	}
	return self
}

// DoHistory records request history and reports whether a failed request was re-enqueued.
func (self *Spider) DoHistory(req *request.Request, ok bool) bool {
	return self.reqMatrix.DoHistory(req, ok)
}

// RequestPush enqueues a request into the scheduling matrix.
func (self *Spider) RequestPush(req *request.Request) {
	self.reqMatrix.Push(req)
}

// RequestPull dequeues the next request from the scheduling matrix.
func (self *Spider) RequestPull() *request.Request {
	return self.reqMatrix.Pull()
}

func (self *Spider) RequestUse() {
	self.reqMatrix.Use()
}

func (self *Spider) RequestFree() {
	self.reqMatrix.Free()
}

func (self *Spider) RequestLen() int {
	return self.reqMatrix.Len()
}

func (self *Spider) TryFlushSuccess() {
	self.reqMatrix.TryFlushSuccess()
}

func (self *Spider) TryFlushFailure() {
	self.reqMatrix.TryFlushFailure()
}

// Start executes the spider's root rule.
func (self *Spider) Start() {
	defer func() {
		if p := recover(); p != nil {
			logs.Log.Error(" *     Panic  [root]: %v\n", p)
		}
		self.lock.Lock()
		self.status = status.RUN
		self.lock.Unlock()
	}()
	self.RuleTree.Root(GetContext(self, nil))
}

// Stop gracefully stops the spider and cancels all timers.
func (self *Spider) Stop() {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.status == status.STOP {
		return
	}
	self.status = status.STOP
	if self.timer != nil {
		self.timer.drop()
		self.timer = nil
	}
}

// CanStop reports whether the spider can transition to a stopped state.
func (self *Spider) CanStop() bool {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.status != status.STOPPED && self.reqMatrix.CanStop()
}

// IsStopping reports whether the spider is in the process of stopping.
func (self *Spider) IsStopping() bool {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.status == status.STOP
}

// tryStop returns ErrForcedStop if the spider is being stopped, nil otherwise.
func (self *Spider) tryStop() error {
	if self.IsStopping() {
		return ErrForcedStop
	}
	return nil
}

// Defer performs cleanup before the spider exits: cancels timers, waits for in-flight requests, and flushes failures.
func (self *Spider) Defer() {
	if self.timer != nil {
		self.timer.drop()
		self.timer = nil
	}
	self.reqMatrix.Wait()
	self.reqMatrix.TryFlushFailure()
}

// OutDefaultField reports whether default fields (Url/ParentUrl/DownloadTime) should be included in output.
func (self *Spider) OutDefaultField() bool {
	return !self.NotDefaultField
}
