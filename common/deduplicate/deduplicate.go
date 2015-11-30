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
	Submit(provider ...string)
	// 从指定输出方式(数据库或文件)读取更新去重记录
	Update(provider ...string)
	// 取消指定去重样本
	Remove(obj interface{})
	// 清空样本缓存
	CleanCache()
}

type Deduplication struct {
	sampling map[string]bool
	provider string
	sync.Mutex
}

func New() Deduplicate {
	return &Deduplication{
		sampling: make(map[string]bool),
	}
}

// 对比是否已存在，不存在则采样
func (self *Deduplication) Compare(obj interface{}) bool {
	self.Mutex.Lock()
	defer self.Unlock()

	s := util.MakeUnique(obj)
	if !self.sampling[s] {
		self.sampling[s] = true
		return false
	}
	return true
}

// 取消指定去重样本
func (self *Deduplication) Remove(obj interface{}) {
	self.Mutex.Lock()
	defer self.Unlock()

	s := util.MakeUnique(obj)
	if self.sampling[s] {
		delete(self.sampling, s)
	}
}

func (self *Deduplication) Submit(provider ...string) {
	if len(provider) > 0 && provider[0] != "" {
		self.provider = provider[0]
	}
	if len(self.sampling) == 0 {
		return
	}
	switch self.provider {
	case "mgo":
		var docs = make([]map[string]interface{}, len(self.sampling))
		var i int
		for key := range self.sampling {
			docs[i] = map[string]interface{}{"_id": key}
			i++
		}
		// fmt.Println(docs)
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
		for key := range self.sampling {
			table.AddRow(key).Update()
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

		// 创建并写入文件
		f, _ := os.Create(config.COMM_PATH.CACHE + "/" + config.DEDUPLICATION.FILE_NAME)
		b, _ := json.Marshal(self.sampling)
		f.Write(b)
		f.Close()
	}
}

func (self *Deduplication) Update(provider ...string) {
	if len(provider) > 0 && provider[0] != self.provider {
		self.provider = provider[0]
		self.CleanCache()
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
			self.sampling[v.(bson.M)["_id"].(string)] = true
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
			self.sampling[id] = true
		}

	default:
		f, err := os.Open(config.COMM_PATH.CACHE + "/" + config.DEDUPLICATION.FILE_NAME)
		if err != nil {
			return
		}
		defer f.Close()
		b, _ := ioutil.ReadAll(f)
		json.Unmarshal(b, &self.sampling)
	}
}

func (self *Deduplication) CleanCache() {
	self.sampling = make(map[string]bool)
}
