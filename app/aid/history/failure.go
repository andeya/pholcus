package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/henrylee2cn/pholcus/app/downloader/request"
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/pool"
)

type Failure struct {
	// [spiderName][reqJson]
	list        map[string]map[string]bool
	inheritable bool
	sync.RWMutex
}

// 获取指定蜘蛛在上一次运行时失败的请求
func (self *Failure) PullFailure(spiderName string) (reqs []*request.Request) {
	if len(self.list[spiderName]) == 0 {
		return
	}

	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	for failure, _ := range self.list[spiderName] {
		req, err := request.UnSerialize(failure)
		if err == nil {
			reqs = append(reqs, req)
		}
	}
	self.list[spiderName] = make(map[string]bool)
	return
}

// 更新或加入失败记录
// 对比是否已存在，不存在就记录
func (self *Failure) UpsertFailure(req *request.Request) bool {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()

	spName := req.GetSpiderName()
	s := req.Serialize()

	if failures, ok := self.list[spName]; !ok {
		self.list[spName] = make(map[string]bool)
	} else if failures[s] {
		return false
	}

	self.list[spName][s] = true
	return true
}

// 删除失败记录
func (self *Failure) DeleteFailure(req *request.Request) {
	self.RWMutex.Lock()
	s := req.Serialize()
	delete(self.list[req.GetSpiderName()], s)
	self.RWMutex.Unlock()
}

func (self *Failure) flush(provider string) (fLen int, err error) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	for _, val := range self.list {
		fLen += len(val)
	}

	switch provider {
	case "mgo":
		if mgo.Error() != nil {
			err = fmt.Errorf(" *     Fail  [添加失败记录][mgo]: %v 条 [ERROR]  %v\n", fLen, mgo.Error())
			return
		}
		mgo.Call(func(src pool.Src) error {
			c := src.(*mgo.MgoSrc).DB(MGO_DB).C(FAILURE_FILE)
			// 删除失败记录文件
			c.DropCollection()
			if fLen == 0 {
				return nil
			}

			var docs = []interface{}{}
			for _, val := range self.list {
				for key := range val {
					docs = append(docs, map[string]interface{}{"_id": key})
				}
			}
			c.Insert(docs...)
			return nil
		})

	case "mysql":
		db, err := mysql.DB()
		if err != nil {
			return fLen, fmt.Errorf(" *     Fail  [添加失败记录][mysql]: %v 条 [ERROR]  %v\n", fLen, err)
		}

		// 删除失败记录文件
		stmt, err := db.Prepare(`DROP TABLE ` + FAILURE_FILE)
		if err != nil {
			return fLen, fmt.Errorf(" *     Fail  [添加失败记录][mysql]: %v 条 [ERROR]  %v\n", fLen, err)
		}
		stmt.Exec()
		if fLen == 0 {
			return fLen, nil
		}

		table := mysql.New(db).
			SetTableName(FAILURE_FILE).
			AddColumn(`failure MEDIUMTEXT`).
			Create()
		for _, val := range self.list {
			for key := range val {
				table.AddRow(key).Update()
			}
		}

	default:
		// 删除失败记录文件
		os.Remove(FAILURE_FILE_FULL)
		if fLen == 0 {
			return
		}

		f, _ := os.OpenFile(FAILURE_FILE_FULL, os.O_CREATE|os.O_WRONLY, 0660)

		b, _ := json.Marshal(self.list)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()
	}
	return
}
