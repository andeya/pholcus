package collector

import (
	"fmt"

	mgov2 "gopkg.in/mgo.v2"

	"github.com/henrylee2cn/pholcus/common/mgo"
	"github.com/henrylee2cn/pholcus/common/pool"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

/************************ MongoDB 输出 ***************************/

func init() {
	Output["mgo"] = func(self *Collector, dataIndex int) error {
		//连接数据库
		if mgo.Error() != nil {
			return fmt.Errorf("MongoBD数据库链接失败: %v", mgo.Error())
		}
		return mgo.Call(func(src pool.Src) error {
			var (
				db          = src.(*mgo.MgoSrc).DB(config.DB_NAME)
				namespace   = util.FileNameReplace(self.namespace())
				collections = make(map[string]*mgov2.Collection)
				dataMap     = make(map[string][]interface{})
				err         error
			)

			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				subNamespace := util.FileNameReplace(self.subNamespace(datacell))
				var cName = namespace
				if subNamespace != "" {
					cName += "__" + subNamespace
				}
				if _, ok := collections[subNamespace]; !ok {
					collections[subNamespace] = db.C(cName)
				}
				for k, v := range datacell["Data"].(map[string]interface{}) {
					datacell[k] = v
				}
				delete(datacell, "Data")
				delete(datacell, "RuleName")
				dataMap[subNamespace] = append(dataMap[subNamespace], datacell)
			}

			for k, v := range dataMap {
				err = collections[k].Insert(v...)
				if err != nil {
					logs.Log.Error("%v", err)
				}
			}
			return nil
		})
	}
}
