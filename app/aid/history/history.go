package history

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"gopkg.in/mgo.v2/bson"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/common/closer"
	"github.com/andeya/pholcus/common/mgo"
	"github.com/andeya/pholcus/common/mysql"
	"github.com/andeya/pholcus/common/pool"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

type (
	HistoryStore interface {
		ReadSuccess(provider string, inherit bool) result.VoidResult // Read success records
		UpsertSuccess(string) bool                                   // Upsert a success record
		HasSuccess(string) bool                                      // Check if a success record exists
		DeleteSuccess(string)                                        // Delete a success record
		FlushSuccess(provider string) result.VoidResult              // Flush success records to I/O without clearing cache

		ReadFailure(provider string, inherit bool) result.VoidResult // Read failure records
		PullFailure() map[string]*request.Request                    // Pull failure records and clear
		UpsertFailure(*request.Request) bool                         // Upsert a failure record
		DeleteFailure(*request.Request)                              // Delete a failure record
		FlushFailure(provider string) result.VoidResult              // Flush failure records to I/O without clearing cache

		Empty() // Clear cache without output
	}
	// History stores success and failure records for crawl deduplication.
	History struct {
		*Success
		*Failure
		provider string
		sync.RWMutex
	}
)

const (
	SuccessSuffix = config.HistoryTag + "__y"
	FailureSuffix = config.HistoryTag + "__n"
	SuccessFile   = config.HistoryDir + "/" + SuccessSuffix
	FailureFile   = config.HistoryDir + "/" + FailureSuffix
)

// New creates a HistoryStore for the given spider name and optional subname.
func New(name string, subName string) HistoryStore {
	successTabName := SuccessSuffix + "__" + name
	successFileName := SuccessFile + "__" + name
	failureTabName := FailureSuffix + "__" + name
	failureFileName := FailureFile + "__" + name
	if subName != "" {
		successTabName += "__" + subName
		successFileName += "__" + subName
		failureTabName += "__" + subName
		failureFileName += "__" + subName
	}
	return &History{
		Success: &Success{
			tabName:  util.FileNameReplace(successTabName),
			fileName: successFileName,
			new:      make(map[string]bool),
			old:      make(map[string]bool),
		},
		Failure: &Failure{
			tabName:  util.FileNameReplace(failureTabName),
			fileName: failureFileName,
			list:     make(map[string]*request.Request),
		},
	}
}

// ReadSuccess reads success records from the given provider.
func (h *History) ReadSuccess(provider string, inherit bool) result.VoidResult {
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()

	if !inherit {
		// Not inheriting history
		h.Success.old = make(map[string]bool)
		h.Success.new = make(map[string]bool)
		h.Success.inheritable = false
		return result.OkVoid()

	} else if h.Success.inheritable {
		// Both current and previous runs inherit history
		return result.OkVoid()

	} else {
		// Previous run did not inherit, but current run does
		h.Success.old = make(map[string]bool)
		h.Success.new = make(map[string]bool)
		h.Success.inheritable = true
	}

	switch provider {
	case "mgo":
		var docs = map[string]interface{}{}
		r := mgo.Mgo(&docs, "find", map[string]interface{}{
			"Database":   config.Conf().DBName,
			"Collection": h.Success.tabName,
		})
		if r.IsErr() {
			logs.Log().Error(" *     Fail  [read success record][mgo]: %v\n", r.UnwrapErr())
			return result.OkVoid()
		}
		for _, v := range docs["Docs"].([]interface{}) {
			h.Success.old[v.(bson.M)["_id"].(string)] = true
		}

	case "mysql":
		_, err := mysql.DB()
		if err != nil {
			logs.Log().Error(" *     Fail  [read success record][mysql]: %v\n", err)
			return result.OkVoid()
		}
		table, ok := getReadMysqlTable(h.Success.tabName)
		if !ok {
			table = mysql.New().Unwrap().SetTableName(h.Success.tabName)
			setReadMysqlTable(h.Success.tabName, table)
		}
		r := table.SelectAll()
		if r.IsErr() {
			return result.OkVoid()
		}
		rows := r.Unwrap()

		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			h.Success.old[id] = true
		}

	default:
		f, err := os.Open(h.Success.fileName)
		if err != nil {
			return result.OkVoid()
		}
		defer closer.LogClose(f, logs.Log().Error)
		b, _ := io.ReadAll(f)
		if len(b) == 0 {
			return result.OkVoid()
		}
		b[0] = '{'
		json.Unmarshal(append(b, '}'), &h.Success.old)
	}
	logs.Log().Informational(" *     [read success record]: %v\n", len(h.Success.old))
	return result.OkVoid()
}

// ReadFailure reads failure records from the given provider.
func (h *History) ReadFailure(provider string, inherit bool) result.VoidResult {
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()

	if !inherit {
		// Not inheriting history
		h.Failure.list = make(map[string]*request.Request)
		h.Failure.inheritable = false
		return result.OkVoid()

	} else if h.Failure.inheritable {
		// Both current and previous runs inherit history
		return result.OkVoid()

	} else {
		// Previous run did not inherit, but current run does
		h.Failure.list = make(map[string]*request.Request)
		h.Failure.inheritable = true
	}
	var fLen int
	switch provider {
	case "mgo":
		if mgo.Error() != nil {
			logs.Log().Error(" *     Fail  [read failure record][mgo]: %v\n", mgo.Error())
			return result.OkVoid()
		}

		var docs = []interface{}{}
		mgo.Call(func(src pool.Src) error {
			c := src.(*mgo.MgoSrc).DB(config.Conf().DBName).C(h.Failure.tabName)
			return c.Find(nil).All(&docs)
		}).Unwrap()

		fLen = len(docs)

		for _, v := range docs {
			key := v.(bson.M)["_id"].(string)
			failure := v.(bson.M)["failure"].(string)
			reqResult := request.UnSerialize(failure)
			if reqResult.IsErr() {
				continue
			}
			h.Failure.list[key] = reqResult.Unwrap()
		}

	case "mysql":
		_, err := mysql.DB()
		if err != nil {
			logs.Log().Error(" *     Fail  [read failure record][mysql]: %v\n", err)
			return result.OkVoid()
		}
		table, ok := getReadMysqlTable(h.Failure.tabName)
		if !ok {
			table = mysql.New().Unwrap().SetTableName(h.Failure.tabName)
			setReadMysqlTable(h.Failure.tabName, table)
		}
		r := table.SelectAll()
		if r.IsErr() {
			return result.OkVoid()
		}
		rows := r.Unwrap()

		for rows.Next() {
			var key, failure string
			err = rows.Scan(&key, &failure)
			reqResult := request.UnSerialize(failure)
			if reqResult.IsErr() {
				continue
			}
			h.Failure.list[key] = reqResult.Unwrap()
			fLen++
		}

	default:
		f, err := os.Open(h.Failure.fileName)
		if err != nil {
			return result.OkVoid()
		}
		defer closer.LogClose(f, logs.Log().Error)
		b, _ := io.ReadAll(f)

		if len(b) == 0 {
			return result.OkVoid()
		}

		docs := map[string]string{}
		json.Unmarshal(b, &docs)

		fLen = len(docs)

		for key, s := range docs {
			reqResult := request.UnSerialize(s)
			if reqResult.IsErr() {
				continue
			}
			h.Failure.list[key] = reqResult.Unwrap()
		}
	}

	logs.Log().Informational(" *     [read failure record]: %v\n", fLen)
	return result.OkVoid()
}

// Empty clears the cache without output.
func (h *History) Empty() {
	h.RWMutex.Lock()
	h.Success.new = make(map[string]bool)
	h.Success.old = make(map[string]bool)
	h.Failure.list = make(map[string]*request.Request)
	h.RWMutex.Unlock()
}

// FlushSuccess flushes success records to I/O without clearing cache.
func (h *History) FlushSuccess(provider string) (r result.VoidResult) {
	defer r.Catch()
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()
	sucLen := h.Success.flush(provider).Unwrap()
	if sucLen <= 0 {
		return result.OkVoid()
	}
	logs.Log().Informational(" *     [add success record]: %v\n", sucLen)
	return result.OkVoid()
}

// FlushFailure flushes failure records to I/O without clearing cache.
func (h *History) FlushFailure(provider string) (r result.VoidResult) {
	defer r.Catch()
	h.RWMutex.Lock()
	h.provider = provider
	h.RWMutex.Unlock()
	failLen := h.Failure.flush(provider).Unwrap()
	if failLen <= 0 {
		return result.OkVoid()
	}
	logs.Log().Informational(" *     [add failure record]: %v\n", failLen)
	return result.OkVoid()
}

var (
	readMysqlTable     = map[string]*mysql.Table{}
	readMysqlTableLock sync.RWMutex
)

func getReadMysqlTable(name string) (*mysql.Table, bool) {
	readMysqlTableLock.RLock()
	tab, ok := readMysqlTable[name]
	readMysqlTableLock.RUnlock()
	if ok {
		return tab.Clone(), true
	}
	return nil, false
}

func setReadMysqlTable(name string, tab *mysql.Table) {
	readMysqlTableLock.Lock()
	readMysqlTable[name] = tab
	readMysqlTableLock.Unlock()
}

var (
	writeMysqlTable     = map[string]*mysql.Table{}
	writeMysqlTableLock sync.RWMutex
)

func getWriteMysqlTable(name string) (*mysql.Table, bool) {
	writeMysqlTableLock.RLock()
	tab, ok := writeMysqlTable[name]
	writeMysqlTableLock.RUnlock()
	if ok {
		return tab.Clone(), true
	}
	return nil, false
}

func setWriteMysqlTable(name string, tab *mysql.Table) {
	writeMysqlTableLock.Lock()
	writeMysqlTable[name] = tab
	writeMysqlTableLock.Unlock()
}
