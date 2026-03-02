package history

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/downloader/request"
	"github.com/andeya/pholcus/common/mgo"
	"github.com/andeya/pholcus/common/mysql"
	"github.com/andeya/pholcus/common/pool"
	"github.com/andeya/pholcus/config"
)

// Failure tracks failed requests for retry.
type Failure struct {
	tabName     string
	fileName    string
	list        map[string]*request.Request
	inheritable bool
	sync.RWMutex
}

func (self *Failure) PullFailure() map[string]*request.Request {
	list := self.list
	self.list = make(map[string]*request.Request)
	return list
}

// UpsertFailure updates or adds a failure record. Returns true if an insert occurred.
func (self *Failure) UpsertFailure(req *request.Request) bool {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	if self.list[req.Unique()] != nil {
		return false
	}
	self.list[req.Unique()] = req
	return true
}

// DeleteFailure removes a failure record.
func (self *Failure) DeleteFailure(req *request.Request) {
	self.RWMutex.Lock()
	delete(self.list, req.Unique())
	self.RWMutex.Unlock()
}

// flush clears historical failure records first, then updates.
func (self *Failure) flush(provider string) (r result.Result[int]) {
	defer r.Catch()
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	fLen := len(self.list)

	switch provider {
	case "mgo":
		result.RetVoid(mgo.Error()).Unwrap()
		mgo.Call(func(src pool.Src) error {
			c := src.(*mgo.MgoSrc).DB(config.DB_NAME).C(self.tabName)
			c.DropCollection()
			if fLen == 0 {
				return nil
			}
			var docs = []interface{}{}
			for key, req := range self.list {
				docs = append(docs, map[string]interface{}{"_id": key, "failure": req.Serialize().Unwrap()})
			}
			c.Insert(docs...)
			return nil
		}).Unwrap()

	case "mysql":
		_, err := mysql.DB()
		result.RetVoid(err).Unwrap()
		table, ok := getWriteMysqlTable(self.tabName)
		if !ok {
			table = mysql.New().Unwrap()
			table.SetTableName(self.tabName).CustomPrimaryKey(`id VARCHAR(255) NOT NULL PRIMARY KEY`).AddColumn(`failure MEDIUMTEXT`)
			setWriteMysqlTable(self.tabName, table)
			table.Create().Unwrap()
		} else {
			table.Truncate().Unwrap()
		}
		for key, req := range self.list {
			table.AutoInsert([]string{key, req.Serialize().Unwrap()})
			table.FlushInsert().Unwrap()
		}

	default:
		os.Remove(self.fileName)
		if fLen == 0 {
			return result.Ok(0)
		}
		f, err := os.OpenFile(self.fileName, os.O_CREATE|os.O_WRONLY, 0777)
		result.RetVoid(err).Unwrap()
		docs := make(map[string]string, len(self.list))
		for key, req := range self.list {
			docs[key] = req.Serialize().Unwrap()
		}
		b, _ := json.Marshal(docs)
		b = bytes.Replace(b, []byte(`\u0026`), []byte(`&`), -1)
		f.Write(b)
		f.Close()
	}
	return result.Ok(fLen)
}
