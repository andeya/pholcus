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

func (f *Failure) PullFailure() map[string]*request.Request {
	list := f.list
	f.list = make(map[string]*request.Request)
	return list
}

// UpsertFailure updates or adds a failure record. Returns true if an insert occurred.
func (f *Failure) UpsertFailure(req *request.Request) bool {
	f.RWMutex.Lock()
	defer f.RWMutex.Unlock()
	if f.list[req.Unique()] != nil {
		return false
	}
	f.list[req.Unique()] = req
	return true
}

// DeleteFailure removes a failure record.
func (f *Failure) DeleteFailure(req *request.Request) {
	f.RWMutex.Lock()
	delete(f.list, req.Unique())
	f.RWMutex.Unlock()
}

// flush clears historical failure records first, then updates.
func (f *Failure) flush(provider string) (r result.Result[int]) {
	defer r.Catch()
	f.RWMutex.Lock()
	defer f.RWMutex.Unlock()
	fLen := len(f.list)

	switch provider {
	case "mgo":
		result.RetVoid(mgo.Error()).Unwrap()
		mgo.Call(func(src pool.Src) error {
			c := src.(*mgo.MgoSrc).DB(config.Conf().DBName).C(f.tabName)
			c.DropCollection()
			if fLen == 0 {
				return nil
			}
			var docs = []interface{}{}
			for key, req := range f.list {
				docs = append(docs, map[string]interface{}{"_id": key, "failure": req.Serialize().Unwrap()})
			}
			c.Insert(docs...)
			return nil
		}).Unwrap()

	case "mysql":
		_, err := mysql.DB()
		result.RetVoid(err).Unwrap()
		table, ok := getWriteMysqlTable(f.tabName)
		if !ok {
			table = mysql.New().Unwrap()
			table.SetTableName(f.tabName).CustomPrimaryKey(`id VARCHAR(255) NOT NULL PRIMARY KEY`).AddColumn(`failure MEDIUMTEXT`)
			setWriteMysqlTable(f.tabName, table)
			table.Create().Unwrap()
		} else {
			table.Truncate().Unwrap()
		}
		for key, req := range f.list {
			table.AutoInsert([]string{key, req.Serialize().Unwrap()})
			table.FlushInsert().Unwrap()
		}

	default:
		os.Remove(f.fileName)
		if fLen == 0 {
			return result.Ok(0)
		}
		file, err := os.OpenFile(f.fileName, os.O_CREATE|os.O_WRONLY, 0777)
		result.RetVoid(err).Unwrap()
		docs := make(map[string]string, len(f.list))
		for key, req := range f.list {
			docs[key] = req.Serialize().Unwrap()
		}
		b, _ := json.Marshal(docs)
		b = bytes.Replace(b, []byte(`\u0026`), []byte(`&`), -1)
		file.Write(b)
		file.Close()
	}
	return result.Ok(fLen)
}
