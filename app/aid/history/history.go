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
	Historier interface {
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
	SUCCESS_SUFFIX = config.HISTORY_TAG + "__y"
	FAILURE_SUFFIX = config.HISTORY_TAG + "__n"
	SUCCESS_FILE   = config.HISTORY_DIR + "/" + SUCCESS_SUFFIX
	FAILURE_FILE   = config.HISTORY_DIR + "/" + FAILURE_SUFFIX
)

// New creates a Historier for the given spider name and optional subname.
func New(name string, subName string) Historier {
	successTabName := SUCCESS_SUFFIX + "__" + name
	successFileName := SUCCESS_FILE + "__" + name
	failureTabName := FAILURE_SUFFIX + "__" + name
	failureFileName := FAILURE_FILE + "__" + name
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
func (self *History) ReadSuccess(provider string, inherit bool) result.VoidResult {
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()

	if !inherit {
		// Not inheriting history
		self.Success.old = make(map[string]bool)
		self.Success.new = make(map[string]bool)
		self.Success.inheritable = false
		return result.OkVoid()

	} else if self.Success.inheritable {
		// Both current and previous runs inherit history
		return result.OkVoid()

	} else {
		// Previous run did not inherit, but current run does
		self.Success.old = make(map[string]bool)
		self.Success.new = make(map[string]bool)
		self.Success.inheritable = true
	}

	switch provider {
	case "mgo":
		var docs = map[string]interface{}{}
		r := mgo.Mgo(&docs, "find", map[string]interface{}{
			"Database":   config.DB_NAME,
			"Collection": self.Success.tabName,
		})
		if r.IsErr() {
			logs.Log.Error(" *     Fail  [read success record][mgo]: %v\n", r.UnwrapErr())
			return result.OkVoid()
		}
		for _, v := range docs["Docs"].([]interface{}) {
			self.Success.old[v.(bson.M)["_id"].(string)] = true
		}

	case "mysql":
		_, err := mysql.DB()
		if err != nil {
			logs.Log.Error(" *     Fail  [read success record][mysql]: %v\n", err)
			return result.OkVoid()
		}
		table, ok := getReadMysqlTable(self.Success.tabName)
		if !ok {
			table = mysql.New().Unwrap().SetTableName(self.Success.tabName)
			setReadMysqlTable(self.Success.tabName, table)
		}
		r := table.SelectAll()
		if r.IsErr() {
			return result.OkVoid()
		}
		rows := r.Unwrap()

		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			self.Success.old[id] = true
		}

	default:
		f, err := os.Open(self.Success.fileName)
		if err != nil {
			return result.OkVoid()
		}
		defer closer.LogClose(f, logs.Log.Error)
		b, _ := io.ReadAll(f)
		if len(b) == 0 {
			return result.OkVoid()
		}
		b[0] = '{'
		json.Unmarshal(append(b, '}'), &self.Success.old)
	}
	logs.Log.Informational(" *     [read success record]: %v\n", len(self.Success.old))
	return result.OkVoid()
}

// ReadFailure reads failure records from the given provider.
func (self *History) ReadFailure(provider string, inherit bool) result.VoidResult {
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()

	if !inherit {
		// Not inheriting history
		self.Failure.list = make(map[string]*request.Request)
		self.Failure.inheritable = false
		return result.OkVoid()

	} else if self.Failure.inheritable {
		// Both current and previous runs inherit history
		return result.OkVoid()

	} else {
		// Previous run did not inherit, but current run does
		self.Failure.list = make(map[string]*request.Request)
		self.Failure.inheritable = true
	}
	var fLen int
	switch provider {
	case "mgo":
		if mgo.Error() != nil {
			logs.Log.Error(" *     Fail  [read failure record][mgo]: %v\n", mgo.Error())
			return result.OkVoid()
		}

		var docs = []interface{}{}
		mgo.Call(func(src pool.Src) error {
			c := src.(*mgo.MgoSrc).DB(config.DB_NAME).C(self.Failure.tabName)
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
			self.Failure.list[key] = reqResult.Unwrap()
		}

	case "mysql":
		_, err := mysql.DB()
		if err != nil {
			logs.Log.Error(" *     Fail  [read failure record][mysql]: %v\n", err)
			return result.OkVoid()
		}
		table, ok := getReadMysqlTable(self.Failure.tabName)
		if !ok {
			table = mysql.New().Unwrap().SetTableName(self.Failure.tabName)
			setReadMysqlTable(self.Failure.tabName, table)
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
			self.Failure.list[key] = reqResult.Unwrap()
			fLen++
		}

	default:
		f, err := os.Open(self.Failure.fileName)
		if err != nil {
			return result.OkVoid()
		}
		defer closer.LogClose(f, logs.Log.Error)
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
			self.Failure.list[key] = reqResult.Unwrap()
		}
	}

	logs.Log.Informational(" *     [read failure record]: %v\n", fLen)
	return result.OkVoid()
}

// Empty clears the cache without output.
func (self *History) Empty() {
	self.RWMutex.Lock()
	self.Success.new = make(map[string]bool)
	self.Success.old = make(map[string]bool)
	self.Failure.list = make(map[string]*request.Request)
	self.RWMutex.Unlock()
}

// FlushSuccess flushes success records to I/O without clearing cache.
func (self *History) FlushSuccess(provider string) (r result.VoidResult) {
	defer r.Catch()
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()
	sucLen := self.Success.flush(provider).Unwrap()
	if sucLen <= 0 {
		return result.OkVoid()
	}
	logs.Log.Informational(" *     [add success record]: %v\n", sucLen)
	return result.OkVoid()
}

// FlushFailure flushes failure records to I/O without clearing cache.
func (self *History) FlushFailure(provider string) (r result.VoidResult) {
	defer r.Catch()
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()
	failLen := self.Failure.flush(provider).Unwrap()
	if failLen <= 0 {
		return result.OkVoid()
	}
	logs.Log.Informational(" *     [add failure record]: %v\n", failLen)
	return result.OkVoid()
}

var (
	readMysqlTable     = map[string]*mysql.MyTable{}
	readMysqlTableLock sync.RWMutex
)

func getReadMysqlTable(name string) (*mysql.MyTable, bool) {
	readMysqlTableLock.RLock()
	tab, ok := readMysqlTable[name]
	readMysqlTableLock.RUnlock()
	if ok {
		return tab.Clone(), true
	}
	return nil, false
}

func setReadMysqlTable(name string, tab *mysql.MyTable) {
	readMysqlTableLock.Lock()
	readMysqlTable[name] = tab
	readMysqlTableLock.Unlock()
}

var (
	writeMysqlTable     = map[string]*mysql.MyTable{}
	writeMysqlTableLock sync.RWMutex
)

func getWriteMysqlTable(name string) (*mysql.MyTable, bool) {
	writeMysqlTableLock.RLock()
	tab, ok := writeMysqlTable[name]
	writeMysqlTableLock.RUnlock()
	if ok {
		return tab.Clone(), true
	}
	return nil, false
}

func setWriteMysqlTable(name string, tab *mysql.MyTable) {
	writeMysqlTableLock.Lock()
	writeMysqlTable[name] = tab
	writeMysqlTableLock.Unlock()
}
