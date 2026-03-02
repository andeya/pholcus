package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/mgo"
	"github.com/andeya/pholcus/common/mysql"
	"github.com/andeya/pholcus/config"
)

// Success tracks successfully crawled request IDs for deduplication.
type Success struct {
	tabName     string
	fileName    string
	new         map[string]bool
	old         map[string]bool
	inheritable bool
	sync.RWMutex
}

// UpsertSuccess updates or adds a success record. Returns true if an insert occurred.
func (self *Success) UpsertSuccess(reqUnique string) bool {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	if self.old[reqUnique] {
		return false
	}
	if self.new[reqUnique] {
		return false
	}
	self.new[reqUnique] = true
	return true
}

func (self *Success) HasSuccess(reqUnique string) bool {
	self.RWMutex.Lock()
	has := self.old[reqUnique] || self.new[reqUnique]
	self.RWMutex.Unlock()
	return has
}

// DeleteSuccess removes a success record.
func (self *Success) DeleteSuccess(reqUnique string) {
	self.RWMutex.Lock()
	delete(self.new, reqUnique)
	self.RWMutex.Unlock()
}

func (self *Success) flush(provider string) result.Result[int] {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	sLen := len(self.new)
	if sLen == 0 {
		return result.Ok(0)
	}

	switch provider {
	case "mgo":
		if mgo.Error() != nil {
			return result.TryErr[int](fmt.Errorf(" *     Fail  [add success record][mgo]: %v [ERROR]  %v\n", sLen, mgo.Error()))
		}
		var docs = make([]map[string]interface{}, sLen)
		var i int
		for key := range self.new {
			docs[i] = map[string]interface{}{"_id": key}
			self.old[key] = true
			i++
		}
		r := mgo.Mgo(nil, "insert", map[string]interface{}{
			"Database":   config.DB_NAME,
			"Collection": self.tabName,
			"Docs":       docs,
		})
		if r.IsErr() {
			return result.TryErr[int](fmt.Errorf(" *     Fail  [add success record][mgo]: %v [ERROR]  %v\n", sLen, r.UnwrapErr()))
		}

	case "mysql":
		_, err := mysql.DB()
		if err != nil {
			return result.TryErr[int](fmt.Errorf(" *     Fail  [add success record][mysql]: %v [ERROR]  %v\n", sLen, err))
		}
		table, ok := getWriteMysqlTable(self.tabName)
		if !ok {
			table = mysql.New().Unwrap()
			table.SetTableName(self.tabName).CustomPrimaryKey(`id VARCHAR(255) NOT NULL PRIMARY KEY`)
			if r := table.Create(); r.IsErr() {
				return result.TryErr[int](fmt.Errorf(" *     Fail  [add success record][mysql]: %v [ERROR]  %v\n", sLen, r.UnwrapErr()))
			}
			setWriteMysqlTable(self.tabName, table)
		}
		for key := range self.new {
			table.AutoInsert([]string{key})
			self.old[key] = true
		}
		if r := table.FlushInsert(); r.IsErr() {
			return result.TryErr[int](fmt.Errorf(" *     Fail  [add success record][mysql]: %v [ERROR]  %v\n", sLen, r.UnwrapErr()))
		}

	default:
		f, _ := os.OpenFile(self.fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)

		b, _ := json.Marshal(self.new)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()

		for key := range self.new {
			self.old[key] = true
		}
	}
	self.new = make(map[string]bool)
	return result.Ok(sLen)
}
