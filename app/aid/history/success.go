package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/config"
)

type Success struct {
	name string
	// [hash(url+method)]true
	new         map[string]bool
	old         map[string]bool
	inheritable bool
	sync.RWMutex
}

// 更新或加入成功记录，
// 对比是否已存在，不存在就记录，
// 返回值表示是否有插入操作。
func (self *Success) UpsertSuccess(hash string) bool {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	if self.old[hash] {
		return false
	}
	if self.new[hash] {
		return false
	}
	self.new[hash] = true
	return true
}

func (self *Success) HasSuccess(hash string) bool {
	self.RWMutex.Lock()
	has := self.old[hash] || self.new[hash]
	self.RWMutex.Unlock()
	return has
}

// 删除成功记录
func (self *Success) DeleteSuccess(hash string) {
	self.RWMutex.Lock()
	delete(self.new, hash)
	self.RWMutex.Unlock()
}

func (self *Success) flush(provider string) (sLen int, err error) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	sLen = len(self.new)
	if sLen == 0 {
		return
	}

	switch provider {
	case "mgo":
		if mgo.Error() != nil {
			err = fmt.Errorf(" *     Fail  [添加成功记录][mgo]: %v 条 [ERROR]  %v\n", sLen, mgo.Error())
			return
		}
		var docs = make([]map[string]interface{}, sLen)
		var i int
		for key := range self.new {
			docs[i] = map[string]interface{}{"_id": key}
			self.old[key] = true
			i++
		}
		err := mgo.Mgo(nil, "insert", map[string]interface{}{
			"Database":   config.DB_NAME,
			"Collection": SUCCESS_SUFFIX + "__" + self.name,
			"Docs":       docs,
		})
		if err != nil {
			err = fmt.Errorf(" *     Fail  [添加成功记录][mgo]: %v 条 [ERROR]  %v\n", sLen, err)
		}

	case "mysql":
		db, err := mysql.DB()
		if err != nil {
			return sLen, fmt.Errorf(" *     Fail  [添加成功记录][mysql]: %v 条 [ERROR]  %v\n", sLen, err)
		}
		table := mysql.New(db).
			SetTableName("`" + SUCCESS_SUFFIX + "__" + self.name + "`").
			CustomPrimaryKey(`id VARCHAR(255) not null primary key`).
			Create()
		for key := range self.new {
			table.AddRow(key).Update()
			self.old[key] = true
		}

	default:
		f, _ := os.OpenFile(SUCCESS_FILE+"__"+self.name, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)

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
