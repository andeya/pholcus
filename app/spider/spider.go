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
	FORCED_STOP = "-- Forced stop of Spider --"
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
		Namespace       func(sp *Spider) string                                    // namespace for output file/path naming
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
func (sp *Spider) Register() *Spider {
	sp.status = status.STOPPED
	return Species.Add(sp)
}

// GetItemFields returns the result field names for the given rule.
func (sp *Spider) GetItemFields(rule *Rule) []string {
	return rule.ItemFields
}

// GetItemField returns the field name at the given index, or "" if out of range.
func (sp *Spider) GetItemField(rule *Rule, index int) (field string) {
	if index > len(rule.ItemFields)-1 || index < 0 {
		return ""
	}
	return rule.ItemFields[index]
}

// GetItemFieldIndex returns the index of the given field name, or -1 if not found.
func (sp *Spider) GetItemFieldIndex(rule *Rule, field string) (index int) {
	for idx, v := range rule.ItemFields {
		if v == field {
			return idx
		}
	}
	return -1
}

// UpsertItemField appends a result field name to the rule and returns its index.
// If the field already exists, the existing index is returned.
func (sp *Spider) UpsertItemField(rule *Rule, field string) (index int) {
	for i, v := range rule.ItemFields {
		if v == field {
			return i
		}
	}
	rule.ItemFields = append(rule.ItemFields, field)
	return len(rule.ItemFields) - 1
}

// GetName returns the spider name.
func (sp *Spider) GetName() string {
	return sp.Name
}

// GetSubName returns the secondary identifier derived from Keyin (computed once).
func (sp *Spider) GetSubName() string {
	sp.once.Do(func() {
		sp.subName = sp.GetKeyin()
		sp.subName = util.MakeHash(sp.subName)
	})
	return sp.subName
}

// GetRule returns the rule with the given name.
func (sp *Spider) GetRule(ruleName string) *Rule {
	rule, ok := sp.RuleTree.Trunk[ruleName]
	if !ok {
		return nil
	}
	return rule
}

// MustGetRule returns the rule with the given name (panics if missing).
func (sp *Spider) MustGetRule(ruleName string) *Rule {
	rule := sp.GetRule(ruleName)
	if rule == nil {
		panic("spider: rule not found: " + ruleName)
	}
	return rule
}

// GetRules returns the full rule map.
func (sp *Spider) GetRules() map[string]*Rule {
	return sp.RuleTree.Trunk
}

// GetDescription returns the spider description.
func (sp *Spider) GetDescription() string {
	return sp.Description
}

// GetID returns the spider's queue index.
func (sp *Spider) GetID() int {
	return sp.id
}

// SetID assigns the spider's queue index.
func (sp *Spider) SetID(id int) {
	sp.id = id
}

// GetKeyin returns the custom keyword/configuration input.
func (sp *Spider) GetKeyin() string {
	return sp.Keyin
}

// SetKeyin sets the custom keyword/configuration input.
func (sp *Spider) SetKeyin(keyword string) {
	sp.Keyin = keyword
}

// GetLimit returns the crawl limit.
// Negative means request-count limiting; positive means custom rule-based limiting.
func (sp *Spider) GetLimit() int64 {
	return sp.Limit
}

// SetLimit sets the crawl limit.
func (sp *Spider) SetLimit(max int64) {
	sp.Limit = max
}

// GetEnableCookie reports whether requests carry cookies.
func (sp *Spider) GetEnableCookie() bool {
	return sp.EnableCookie
}

// SetPausetime sets a custom pause interval. Only overwrites an existing value when runtime[0] is true.
func (sp *Spider) SetPausetime(pause int64, runtime ...bool) {
	if sp.Pausetime == 0 || len(runtime) > 0 && runtime[0] {
		sp.Pausetime = pause
	}
}

// SetTimer configures a timer identified by id.
// When bell is nil, tol is a countdown sleep duration; otherwise tol specifies the wake-up occurrence.
func (sp *Spider) SetTimer(id string, tol time.Duration, bell *Bell) bool {
	if sp.timer == nil {
		sp.timer = newTimer()
	}
	return sp.timer.set(id, tol, bell)
}

// RunTimer starts the timer and reports whether it can continue to be used.
func (sp *Spider) RunTimer(id string) bool {
	if sp.timer == nil {
		return false
	}
	return sp.timer.sleep(id)
}

// Copy returns a deep copy of the spider, including its rule tree.
func (sp *Spider) Copy() *Spider {
	ghost := &Spider{}
	ghost.Name = sp.Name
	ghost.subName = sp.subName

	ghost.RuleTree = &RuleTree{
		Root:  sp.RuleTree.Root,
		Trunk: make(map[string]*Rule, len(sp.RuleTree.Trunk)),
	}
	for k, v := range sp.RuleTree.Trunk {
		ghost.RuleTree.Trunk[k] = new(Rule)

		ghost.RuleTree.Trunk[k].ItemFields = make([]string, len(v.ItemFields))
		copy(ghost.RuleTree.Trunk[k].ItemFields, v.ItemFields)

		ghost.RuleTree.Trunk[k].ParseFunc = v.ParseFunc
		ghost.RuleTree.Trunk[k].AidFunc = v.AidFunc
	}

	ghost.Description = sp.Description
	ghost.Pausetime = sp.Pausetime
	ghost.EnableCookie = sp.EnableCookie
	ghost.Limit = sp.Limit
	ghost.Keyin = sp.Keyin

	ghost.NotDefaultField = sp.NotDefaultField
	ghost.Namespace = sp.Namespace
	ghost.SubNamespace = sp.SubNamespace

	ghost.timer = sp.timer
	ghost.status = sp.status

	return ghost
}

// ReqmatrixInit initializes the request scheduling matrix for this spider.
func (sp *Spider) ReqmatrixInit() *Spider {
	if sp.Limit < 0 {
		sp.reqMatrix = scheduler.AddMatrix(sp.GetName(), sp.GetSubName(), sp.Limit)
		sp.SetLimit(0)
	} else {
		sp.reqMatrix = scheduler.AddMatrix(sp.GetName(), sp.GetSubName(), math.MinInt64)
	}
	return sp
}

// DoHistory records request history and reports whether a failed request was re-enqueued.
func (sp *Spider) DoHistory(req *request.Request, ok bool) bool {
	return sp.reqMatrix.DoHistory(req, ok)
}

// RequestPush enqueues a request into the scheduling matrix.
func (sp *Spider) RequestPush(req *request.Request) {
	sp.reqMatrix.Push(req)
}

// RequestPull dequeues the next request from the scheduling matrix.
func (sp *Spider) RequestPull() *request.Request {
	return sp.reqMatrix.Pull()
}

func (sp *Spider) RequestUse() {
	sp.reqMatrix.Use()
}

func (sp *Spider) RequestFree() {
	sp.reqMatrix.Free()
}

func (sp *Spider) RequestLen() int {
	return sp.reqMatrix.Len()
}

func (sp *Spider) TryFlushSuccess() {
	sp.reqMatrix.TryFlushSuccess()
}

func (sp *Spider) TryFlushFailure() {
	sp.reqMatrix.TryFlushFailure()
}

// Start executes the spider's root rule.
func (sp *Spider) Start() {
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error(" *     Panic  [root]: %v\n", p)
		}
		sp.lock.Lock()
		sp.status = status.RUN
		sp.lock.Unlock()
	}()
	sp.RuleTree.Root(GetContext(sp, nil))
}

// Stop gracefully stops the spider and cancels all timers.
func (sp *Spider) Stop() {
	sp.lock.Lock()
	defer sp.lock.Unlock()
	if sp.status == status.STOP {
		return
	}
	sp.status = status.STOP
	if sp.timer != nil {
		sp.timer.drop()
		sp.timer = nil
	}
}

// CanStop reports whether the spider can transition to a stopped state.
func (sp *Spider) CanStop() bool {
	sp.lock.RLock()
	defer sp.lock.RUnlock()
	return sp.status != status.STOPPED && sp.reqMatrix.CanStop()
}

// IsStopping reports whether the spider is in the process of stopping.
func (sp *Spider) IsStopping() bool {
	sp.lock.RLock()
	defer sp.lock.RUnlock()
	return sp.status == status.STOP
}

// tryStop returns ErrForcedStop if the spider is being stopped, nil otherwise.
func (sp *Spider) tryStop() error {
	if sp.IsStopping() {
		return ErrForcedStop
	}
	return nil
}

// Defer performs cleanup before the spider exits: cancels timers, waits for in-flight requests, and flushes failures.
func (sp *Spider) Defer() {
	if sp.timer != nil {
		sp.timer.drop()
		sp.timer = nil
	}
	sp.reqMatrix.Wait()
	sp.reqMatrix.TryFlushFailure()
}

// OutDefaultField reports whether default fields (Url/ParentUrl/DownloadTime) should be included in output.
func (sp *Spider) OutDefaultField() bool {
	return !sp.NotDefaultField
}
