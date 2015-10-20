package deduplicate

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

type Deduplicate interface {
	// 采集非重复样本并返回对比结果，重复为true
	Compare(obj interface{}) bool
	// 保存去重记录到provider(status.FILE or status.MGO)
	Write(provider string)
	// 从provider(status.FILE or status.MGO)读取去重记录
	ReRead(provider string)
	// 取消指定去重样本
	Remove(obj interface{})
	// 清空样本记录
	CleanRead()
}

type Deduplication struct {
	sampling map[string]bool
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

func (self *Deduplication) Write(provider string) {
	switch strings.ToLower(provider) {
	case status.MGO:
		var docs = make([]map[string]interface{}, len(self.sampling))
		var i int
		for key := range self.sampling {
			docs[i] = map[string]interface{}{"_id": key}
			i++
		}
		mgo.Mgo(nil, "insert", map[string]interface{}{
			"Database":   config.DEDUPLICATION.DB,
			"Collection": config.DEDUPLICATION.COLLECTION,
			"Docs":       docs,
		})

	case status.FILE:
		fallthrough
	default:
		p, _ := path.Split(config.DEDUPLICATION.FULL_FILE_NAME)
		// 创建/打开目录
		d, err := os.Stat(p)
		if err != nil || !d.IsDir() {
			if err := os.MkdirAll(p, 0777); err != nil {
				logs.Log.Error("Error: %v\n", err)
			}
		}

		// 创建并写入文件
		f, _ := os.Create(config.DEDUPLICATION.FULL_FILE_NAME)
		b, _ := json.Marshal(self.sampling)
		f.Write(b)
		f.Close()
	}
}

func (self *Deduplication) ReRead(provider string) {
	self.CleanRead()

	switch strings.ToLower(provider) {
	case status.MGO:
		var docs = map[string]interface{}{}
		err := mgo.Mgo(&docs, "find", map[string]interface{}{
			"Database":   config.DEDUPLICATION.DB,
			"Collection": config.DEDUPLICATION.COLLECTION,
		})
		if err != nil {
			logs.Log.Error("去重读取mgo: %v", err)
			return
		}
		for _, v := range docs["Docs"].([]interface{}) {
			self.sampling[v.(bson.M)["_id"].(string)] = true
		}

	case status.FILE:
		fallthrough
	default:
		f, err := os.Open(config.DEDUPLICATION.FULL_FILE_NAME)
		if err != nil {
			return
		}
		defer f.Close()
		b, _ := ioutil.ReadAll(f)
		json.Unmarshal(b, &self.sampling)
	}
	// fmt.Printf("%#v", self.sampling)
}

func (self *Deduplication) CleanRead() {
	self.sampling = make(map[string]bool)
}
