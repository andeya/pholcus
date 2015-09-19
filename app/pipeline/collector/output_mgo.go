package collector

import (
	"gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"

	"github.com/henrylee2cn/pholcus/common/pool"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

/************************ MongoDB 输出 ***************************/
var mgoPool = pool.NewPool(new(mgoFish), 1024)

type mgoFish struct {
	*mgo.Session
}

func (self *mgoFish) New() pool.Fish {
	mgoSession, err := mgo.Dial(config.MGO_OUTPUT.Host)
	if err != nil {
		panic(err)
	}
	mgoSession.SetMode(mgo.Monotonic, true)
	return &mgoFish{Session: mgoSession}
}

// 判断连接有效性
func (self *mgoFish) Usable() bool {
	if self.Session.Ping() != nil {
		return false
	}
	return true
}

// 自毁方法，在被资源池删除时调用
func (self *mgoFish) Close() {
	self.Session.Close()
}

func (*mgoFish) Clean() {}

func init() {
	Output["mgo"] = func(self *Collector, dataIndex int) {
		var err error
		//连接数据库
		mgoSession := mgoPool.GetOne().(*mgoFish)
		defer mgoPool.Free(mgoSession)

		dbname, tabname := dbOrTabName(self)
		db := mgoSession.DB(dbname)

		if tabname == "" {
			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				tabname = tabName(self, datacell["RuleName"].(string))
				collection := db.C(tabname)
				err = collection.Insert(datacell)
				if err != nil {
					logs.Log.Error("%v", err)
				}
			}
			return
		}

		collection := db.C(tabname)

		for i, count := 0, len(self.DockerQueue.Dockers[dataIndex]); i < count; i++ {
			err = collection.Insert((interface{})(self.DockerQueue.Dockers[dataIndex][i]))
			if err != nil {
				logs.Log.Error("%v", err)
			}
		}
	}
}
