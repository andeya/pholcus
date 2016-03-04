package history

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"gopkg.in/mgo.v2/bson"

	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/pool"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

const (
	SUCCESS_SUFFIX = config.HISTORY_TAG + "_y"
	FAILURE_SUFFIX = config.HISTORY_TAG + "_n"
	SUCCESS_FILE   = config.HISTORY_DIR + "/" + SUCCESS_SUFFIX
	FAILURE_FILE   = config.HISTORY_DIR + "/" + FAILURE_SUFFIX
)

type (
	Historier interface {
		ReadSuccess(provider string, inherit bool) // 读取成功记录
		UpsertSuccess(string) bool                 // 更新或加入成功记录
		HasSuccess(string) bool                    // 检查是否存在某条成功记录
		DeleteSuccess(string)                      // 删除成功记录
		FlushSuccess(provider string)              // I/O输出成功记录，但不清缓存

		ReadFailure(provider string, inherit bool) // 读取失败记录
		PullFailure() map[*request.Request]bool    // 拉取失败记录并清空
		UpsertFailure(*request.Request) bool       // 更新或加入失败记录
		DeleteFailure(*request.Request)            // 删除失败记录
		FlushFailure(provider string)              // I/O输出失败记录，但不清缓存

		Empty() // 清空缓存，但不输出
	}
	History struct {
		*Success
		*Failure
		provider string
		sync.RWMutex
	}
)

func New(name string) Historier {
	return &History{
		Success: &Success{
			name: name,
			new:  make(map[string]bool),
			old:  make(map[string]bool),
		},
		Failure: &Failure{
			name: name,
			list: make(map[*request.Request]bool),
		},
	}
}

// 读取成功记录
func (self *History) ReadSuccess(provider string, inherit bool) {
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()

	if !inherit {
		// 不继承历史记录时
		self.Success.old = make(map[string]bool)
		self.Success.new = make(map[string]bool)
		self.Success.inheritable = false
		return

	} else if self.Success.inheritable {
		// 本次与上次均继承历史记录时
		return

	} else {
		// 上次没有继承历史记录，但本次继承时
		self.Success.old = make(map[string]bool)
		self.Success.new = make(map[string]bool)
		self.Success.inheritable = true
	}

	switch provider {
	case "mgo":
		var docs = map[string]interface{}{}
		err := mgo.Mgo(&docs, "find", map[string]interface{}{
			"Database":   config.DB_NAME,
			"Collection": SUCCESS_SUFFIX + "_" + self.Success.name,
		})
		if err != nil {
			logs.Log.Error(" *     Fail  [读取成功记录][mgo]: %v\n", err)
			return
		}
		for _, v := range docs["Docs"].([]interface{}) {
			self.Success.old[v.(bson.M)["_id"].(string)] = true
		}

	case "mysql":
		db, err := mysql.DB()
		if err != nil {
			logs.Log.Error(" *     Fail  [读取成功记录][mysql]: %v\n", err)
			return
		}
		rows, err := mysql.New(db).
			SetTableName("`" + SUCCESS_SUFFIX + "_" + self.Success.name + "`").
			SelectAll()
		if err != nil {
			return
		}

		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			self.Success.old[id] = true
		}

	default:
		f, err := os.Open(SUCCESS_FILE + "_" + self.Success.name)
		if err != nil {
			return
		}
		defer f.Close()
		b, _ := ioutil.ReadAll(f)
		if len(b) == 0 {
			return
		}
		b[0] = '{'
		json.Unmarshal(append(b, '}'), &self.Success.old)
	}
	logs.Log.Informational(" *     [读取成功记录]: %v 条\n", len(self.Success.old))
}

// 读取失败记录
func (self *History) ReadFailure(provider string, inherit bool) {
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()

	if !inherit {
		// 不继承历史记录时
		self.Failure.list = make(map[*request.Request]bool)
		self.Failure.inheritable = false
		return

	} else if self.Failure.inheritable {
		// 本次与上次均继承历史记录时
		return

	} else {
		// 上次没有继承历史记录，但本次继承时
		self.Failure.list = make(map[*request.Request]bool)
		self.Failure.inheritable = true
	}
	var fLen int
	switch provider {
	case "mgo":
		if mgo.Error() != nil {
			logs.Log.Error(" *     Fail  [读取失败记录][mgo]: %v\n", mgo.Error())
			return
		}

		var docs = []interface{}{}
		mgo.Call(func(src pool.Src) error {
			c := src.(*mgo.MgoSrc).DB(config.DB_NAME).C(FAILURE_SUFFIX + "_" + self.Failure.name)
			return c.Find(nil).All(&docs)
		})

		fLen = len(docs)

		for _, v := range docs {
			failure := v.(bson.M)["_id"].(string)
			req, err := request.UnSerialize(failure)
			if err != nil {
				continue
			}
			self.Failure.list[req] = true
		}

	case "mysql":
		db, err := mysql.DB()
		if err != nil {
			logs.Log.Error(" *     Fail  [读取失败记录][mysql]: %v\n", err)
			return
		}
		rows, err := mysql.New(db).
			SetTableName("`" + FAILURE_SUFFIX + "_" + self.Failure.name + "`").
			SelectAll()
		if err != nil {
			// logs.Log.Error("读取Mysql数据库中成功记录失败：%v", err)
			return
		}

		for rows.Next() {
			var id int
			var failure string
			err = rows.Scan(&id, &failure)
			req, err := request.UnSerialize(failure)
			if err != nil {
				continue
			}
			self.Failure.list[req] = true
			fLen++
		}

	default:
		f, err := os.Open(FAILURE_FILE + "_" + self.Failure.name)
		if err != nil {
			return
		}
		b, _ := ioutil.ReadAll(f)
		f.Close()

		if len(b) == 0 {
			return
		}

		docs := []string{}
		json.Unmarshal(b, &docs)

		fLen = len(docs)

		for _, s := range docs {
			req, err := request.UnSerialize(s)
			if err != nil {
				continue
			}
			self.Failure.list[req] = true
		}
	}

	logs.Log.Informational(" *     [读取失败记录]: %v 条\n", fLen)
}

// 清空缓存，但不输出
func (self *History) Empty() {
	self.RWMutex.Lock()
	self.Success.new = make(map[string]bool)
	self.Success.old = make(map[string]bool)
	self.Failure.list = make(map[*request.Request]bool)
	self.RWMutex.Unlock()
}

// I/O输出成功记录，但不清缓存
func (self *History) FlushSuccess(provider string) {
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()
	sucLen, err := self.Success.flush(provider)
	if sucLen <= 0 {
		return
	}
	// logs.Log.Informational(" * ")
	if err != nil {
		logs.Log.Error("%v", err)
	} else {
		logs.Log.Informational(" *     [添加成功记录]: %v 条\n", sucLen)
	}
}

// I/O输出失败记录，但不清缓存
func (self *History) FlushFailure(provider string) {
	self.RWMutex.Lock()
	self.provider = provider
	self.RWMutex.Unlock()
	failLen, err := self.Failure.flush(provider)
	if failLen <= 0 {
		return
	}
	// logs.Log.Informational(" * ")
	if err != nil {
		logs.Log.Error("%v", err)
	} else {
		logs.Log.Informational(" *     [添加失败记录]: %v 条\n", failLen)
	}
}
