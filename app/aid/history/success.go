package history

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

type Success struct {
	// [hash(url+method)]true
	new         map[string]bool
	old         map[string]bool
	inheritable bool
	sync.RWMutex
}

// 更新或加入成功记录
// 对比是否已存在，不存在就记录
func (self *Success) UpsertSuccess(record Record) bool {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	s := util.MakeUnique(record.GetUrl() + record.GetMethod())
	if self.old[s] {
		return false
	}
	if self.new[s] {
		return false
	}
	self.new[s] = true
	return true
}

// 删除成功记录
func (self *Success) DeleteSuccess(record Record) {
	self.RWMutex.Lock()
	s := util.MakeUnique(record.GetUrl() + record.GetMethod())
	delete(self.new, s)
	self.RWMutex.Unlock()
}

func (self *Success) flush(provider string) (sLen int) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	sLen = len(self.new)
	if sLen == 0 {
		return
	}

	switch provider {
	case "mgo":
		var docs = make([]map[string]interface{}, sLen)
		var i int
		for key := range self.new {
			docs[i] = map[string]interface{}{"_id": key}
			self.old[key] = true
			i++
		}
		mgo.Mgo(nil, "insert", map[string]interface{}{
			"Database":   MGO_DB,
			"Collection": SUCCESS_FILE,
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
			SetTableName(SUCCESS_FILE).
			CustomPrimaryKey(`id VARCHAR(255) not null primary key`).
			Create()
		for key := range self.new {
			table.AddRow(key).Update()
			self.old[key] = true
		}

	default:
		once.Do(mkdir)
		f, _ := os.OpenFile(SUCCESS_FILE_FULL, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)

		b, _ := json.Marshal(self.new)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()

		for key := range self.new {
			self.old[key] = true
		}
	}
	self.new = make(map[string]bool)
	return
}
