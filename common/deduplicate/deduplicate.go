package deduplicate

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/mysql"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type Deduplicate interface {
	// 采集非重复样本并返回对比结果，重复为true
	Compare(obj interface{}) bool
	// 提交去重记录至指定输出方式(数据库或文件)
	Submit(provider string)
	// 从指定输出方式(数据库或文件)读取更新去重记录
	Update(provider string, inherit bool)
	// 取消指定去重样本
	Remove(obj interface{})
	// 清空样本缓存
	CleanCache()
}

type Deduplication struct {
	sampling struct {
		new map[string]bool
		old map[string]bool
	}
	provider      string
	lastIsInherit bool
	sync.Mutex
}

func New() Deduplicate {
	return &Deduplication{
		sampling: struct {
			new map[string]bool
			old map[string]bool
		}{
			new: make(map[string]bool),
			old: make(map[string]bool),
		},
	}
}

// 对比是否已存在，不存在则采样
func (self *Deduplication) Compare(obj interface{}) (duplicate bool) {
	self.Mutex.Lock()
	defer self.Unlock()

	s := util.MakeUnique(obj)
	if self.sampling.old[s] {
		return true
	}
	if self.sampling.new[s] {
		return true
	}
	self.sampling.new[s] = true
	return false
}

// 取消指定去重样本
func (self *Deduplication) Remove(obj interface{}) {
	self.Mutex.Lock()
	defer self.Unlock()

	s := util.MakeUnique(obj)
	if self.sampling.new[s] {
		delete(self.sampling.new, s)
	}
}

func (self *Deduplication) Submit(provider string) {
	self.Mutex.Lock()
	defer self.Unlock()

	self.provider = provider

	if len(self.sampling.new) == 0 {
		return
	}

	switch self.provider {
	case "mgo":
		var docs = make([]map[string]interface{}, len(self.sampling.new))
		var i int
		for key := range self.sampling.new {
			docs[i] = map[string]interface{}{"_id": key}
			self.sampling.old[key] = true
			i++
		}
		mgo.Mgo(nil, "insert", map[string]interface{}{
			"Database":   config.MGO.DB,
			"Collection": config.DEDUPLICATION.FILE_NAME,
			"Docs":       docs,
		})

	case "mysql":
		db, ok := mysql.MysqlPool.GetOne().(*mysql.MysqlSrc)
		if !ok || db == nil {
			logs.Log.Error("链接Mysql数据库超时，无法保存去重记录！")
			return
		}
		defer mysql.MysqlPool.Free(db)
		table := mysql.New(db.DB).
			SetTableName(config.DEDUPLICATION.FILE_NAME).
			CustomPrimaryKey(`id VARCHAR(255) not null primary key`).
			Create()
		for key := range self.sampling.new {
			table.AddRow(key).Update()
			self.sampling.old[key] = true
		}

	default:
		p, _ := path.Split(config.COMM_PATH.CACHE + "/" + config.DEDUPLICATION.FILE_NAME)
		// 创建/打开目录
		d, err := os.Stat(p)
		if err != nil || !d.IsDir() {
			if err := os.MkdirAll(p, 0777); err != nil {
				logs.Log.Error("Error: %v\n", err)
			}
		}

		f, _ := os.OpenFile(config.COMM_PATH.CACHE+"/"+config.DEDUPLICATION.FILE_NAME, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)

		b, _ := json.Marshal(self.sampling.new)
		b[0] = ','
		f.Write(b[:len(b)-1])
		f.Close()

		for key := range self.sampling.new {
			self.sampling.old[key] = true
		}
	}
	logs.Log.Informational(" *     新增 %v 条去重样本\n", len(self.sampling.new))
	self.sampling.new = make(map[string]bool)
}

func (self *Deduplication) Update(provider string, inherit bool) {
	self.Mutex.Lock()
	defer self.Unlock()

	self.provider = provider

	if !inherit {
		// 不继承历史记录时
		self.sampling.old = make(map[string]bool)
		self.sampling.new = make(map[string]bool)
		self.lastIsInherit = false
		return

	} else if self.lastIsInherit {
		// 本次与上次均继承历史记录时
		return

	} else {
		// 上次没有继承历史记录，但本次继承时
		self.sampling.old = make(map[string]bool)
		self.sampling.new = make(map[string]bool)
		self.lastIsInherit = true
	}

	switch self.provider {
	case "mgo":
		var docs = map[string]interface{}{}
		err := mgo.Mgo(&docs, "find", map[string]interface{}{
			"Database":   config.MGO.DB,
			"Collection": config.DEDUPLICATION.FILE_NAME,
		})
		if err != nil {
			logs.Log.Error("去重读取mgo: %v", err)
			return
		}
		for _, v := range docs["Docs"].([]interface{}) {
			self.sampling.old[v.(bson.M)["_id"].(string)] = true
		}

	case "mysql":
		db, ok := mysql.MysqlPool.GetOne().(*mysql.MysqlSrc)
		if !ok || db == nil {
			logs.Log.Error("链接Mysql数据库超时，无法读取去重记录！")
			return
		}
		defer mysql.MysqlPool.Free(db)
		rows, err := mysql.New(db.DB).
			SetTableName("`" + config.DEDUPLICATION.FILE_NAME + "`").
			SelectAll()
		if err != nil {
			// logs.Log.Error("读取Mysql数据库中去重记录失败：%v", err)
			return
		}

		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			self.sampling.old[id] = true
		}

	default:
		f, err := os.Open(config.COMM_PATH.CACHE + "/" + config.DEDUPLICATION.FILE_NAME)
		if err != nil {
			return
		}
		defer f.Close()
		b, _ := ioutil.ReadAll(f)
		b[0] = '{'
		json.Unmarshal(
			append(b, '}'),
			&self.sampling.old,
		)
	}
	logs.Log.Informational(" *     读出 %v 条去重样本\n", len(self.sampling.old))
}

func (self *Deduplication) CleanCache() {
	self.Mutex.Lock()
	defer self.Unlock()
	self.sampling.old = make(map[string]bool)
	self.sampling.new = make(map[string]bool)
}
