package collector

import (
	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/logs"
)

/************************ MongoDB 输出 ***************************/

func init() {
	Output["mgo"] = func(self *Collector, dataIndex int) {
		var err error
		//连接数据库
		mgoSession := mgo.MgoPool.GetOne().(*mgo.MgoFish)
		defer mgo.MgoPool.Free(mgoSession)

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
