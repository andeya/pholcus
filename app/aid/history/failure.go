package history

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/henrylee2cn/pholcus/app/downloader/context"
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/logs"
)

type Failure struct {
	// [spiderName][reqJson]
	list        map[string]map[string]bool
	inheritable bool
	sync.RWMutex
}

// 获取指定蜘蛛在上一次运行时失败的请求
func (self *Failure) PullFailure(spiderName string) (reqs []*context.Request) {
	if len(self.list[spiderName]) == 0 {
		return
	}

	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	for v, ok := range self.list[spiderName] {
		if !ok {
			continue
		}
		if req, err := context.UnSerialize(v); err == nil {
			reqs = append(reqs, req)
		}
	}
	self.list[spiderName] = make(map[string]bool)
	return
}

// 更新或加入失败记录
// 对比是否已存在，不存在就记录
func (self *Failure) UpsertFailure(req *context.Request) bool {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	spName := req.GetSpiderName()
	s := req.Serialize()

	if one, ok := self.list[spName]; !ok {
		self.list[spName] = make(map[string]bool)
	} else if one[s] {
		return false
	}
	self.list[spName][s] = true
	return true
}

// 删除失败记录
func (self *Failure) DeleteFailure(req *context.Request) {
	self.RWMutex.Lock()
	s := req.Serialize()
	delete(self.list[req.GetSpiderName()], s)
	self.RWMutex.Unlock()
}

func (self *Failure) flush(provider string) (fLen int) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	for _, val := range self.list {
		fLen += len(val)
	}
	if fLen == 0 {
		return
	}

	switch provider {
	case "mgo":
		s, c, err := mgo.Open(MGO_DB, FAILURE_FILE)
		if err != nil {
			logs.Log.Error("从mgo读取成功记录: %v", err)
			return
		}
		c.DropCollection()
		var docs = []interface{}{}
		for _, val := range self.list {
			for key := range val {
				docs = append(docs, map[string]interface{}{"_id": key})
			}
		}
		c.Insert(docs...)
		mgo.Close(s)

	case "mysql":
		db, ok := mysql.MysqlPool.GetOne().(*mysql.MysqlSrc)
		if !ok || db == nil {
			logs.Log.Error("链接Mysql数据库超时，无法保存去重记录！")
			return 0
		}

		stmt, err := db.DB.Prepare(`DROP TABLE ` + FAILURE_FILE)
		if err != nil {
			return
		}
		_, err = stmt.Exec()

		table := mysql.New(db.DB).
			SetTableName(FAILURE_FILE).
			AddColumn(`failure MEDIUMTEXT`).
			Create()
		for _, val := range self.list {
			for key := range val {
				table.AddRow(key).Update()
			}
		}
		mysql.MysqlPool.Free(db)

	default:
		once.Do(mkdir)
		os.Remove(FAILURE_FILE_FULL)

		f, _ := os.OpenFile(FAILURE_FILE_FULL, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)

		b, _ := json.Marshal(self.list)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()
	}
	return
}
