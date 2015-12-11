package history

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

type Failure struct {
	// [spiderName][reqJson]
	new         map[string]map[string]bool
	old         map[string]map[string]bool
	inheritable bool
	sync.RWMutex
}

// 获取指定蜘蛛在上一次运行时失败的请求
func (self *Failure) PullFailure(spiderName string) (reqs []*context.Request) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	for v := range self.old[spiderName] {
		if req, err := context.UnSerialize(v); err == nil {
			reqs = append(reqs, req)
		}
	}
	delete(self.old, spiderName)
	return
}

// 更新或加入失败记录
// 对比是否已存在，不存在就记录
func (self *Failure) UpsertFailure(req *context.Request) bool {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	spName := req.GetSpiderName()
	s := util.MakeUnique(req.Serialize())

	if list, ok := self.old[spName]; !ok {
		self.old[spName] = make(map[string]bool)
	} else if list[s] {
		return false
	}

	if list, ok := self.new[spName]; !ok {
		self.new[spName] = make(map[string]bool)
	} else if list[s] {
		return false
	}
	self.new[spName][s] = false

	return true
}

// 删除失败记录
func (self *Failure) DeleteFailure(req *context.Request) {
	self.RWMutex.Lock()
	s := util.MakeUnique(req.Serialize())
	delete(self.new[req.GetSpiderName()], s)
	delete(self.old[req.GetSpiderName()], s)
	self.RWMutex.Unlock()
}

func (self *Failure) flush(provider string) (fLen int) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	for keys, val := range self.new {
		for key := range val {
			self.old[keys][key] = true
			fLen++
		}
	}
	if fLen == 0 {
		return
	}

	switch provider {
	case "mgo":
		var docs = []map[string]interface{}{}
		for _, val := range self.new {
			for key := range val {
				docs = append(docs, map[string]interface{}{"_id": key})
			}
		}
		mgo.Mgo(nil, "insert", map[string]interface{}{
			"Database":   MGO_DB,
			"Collection": FAILURE_FILE,
			"Docs":       docs,
		})

	case "mysql":
		db, ok := mysql.MysqlPool.GetOne().(*mysql.MysqlSrc)
		if !ok || db == nil {
			logs.Log.Error("链接Mysql数据库超时，无法保存去重记录！")
			return 0
		}
		defer mysql.MysqlPool.Free(db)
		table := mysql.New(db.DB).
			SetTableName(FAILURE_FILE).
			CustomPrimaryKey(`id VARCHAR(255) not null primary key`).
			Create()
		for _, val := range self.new {
			for key := range val {
				table.AddRow(key).Update()
			}
		}

	default:
		once.Do(mkdir)

		f, _ := os.OpenFile(FAILURE_FILE_FULL, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)

		b, _ := json.Marshal(self.new)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()
	}
	self.new = make(map[string]map[string]bool)
	return
}
