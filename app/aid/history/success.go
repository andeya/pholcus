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
func (s *Success) UpsertSuccess(reqUnique string) bool {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	if s.old[reqUnique] {
		return false
	}
	if s.new[reqUnique] {
		return false
	}
	s.new[reqUnique] = true
	return true
}

func (s *Success) HasSuccess(reqUnique string) bool {
	s.RWMutex.Lock()
	has := s.old[reqUnique] || s.new[reqUnique]
	s.RWMutex.Unlock()
	return has
}

// DeleteSuccess removes a success record.
func (s *Success) DeleteSuccess(reqUnique string) {
	s.RWMutex.Lock()
	delete(s.new, reqUnique)
	s.RWMutex.Unlock()
}

func (s *Success) flush(provider string) result.Result[int] {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	sLen := len(s.new)
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
		for key := range s.new {
			docs[i] = map[string]interface{}{"_id": key}
			s.old[key] = true
			i++
		}
		r := mgo.Mgo(nil, "insert", map[string]interface{}{
			"Database":   config.Conf().DBName,
			"Collection": s.tabName,
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
		table, ok := getWriteMysqlTable(s.tabName)
		if !ok {
			table = mysql.New().Unwrap()
			table.SetTableName(s.tabName).CustomPrimaryKey(`id VARCHAR(255) NOT NULL PRIMARY KEY`)
			if r := table.Create(); r.IsErr() {
				return result.TryErr[int](fmt.Errorf(" *     Fail  [add success record][mysql]: %v [ERROR]  %v\n", sLen, r.UnwrapErr()))
			}
			setWriteMysqlTable(s.tabName, table)
		}
		for key := range s.new {
			table.AutoInsert([]string{key})
			s.old[key] = true
		}
		if r := table.FlushInsert(); r.IsErr() {
			return result.TryErr[int](fmt.Errorf(" *     Fail  [add success record][mysql]: %v [ERROR]  %v\n", sLen, r.UnwrapErr()))
		}

	default:
		f, _ := os.OpenFile(s.fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)

		b, _ := json.Marshal(s.new)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()

		for key := range s.new {
			s.old[key] = true
		}
	}
	s.new = make(map[string]bool)
	return result.Ok(sLen)
}
