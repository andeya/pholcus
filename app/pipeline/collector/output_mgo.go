//go:build !coverage

package collector

import (
	mgov2 "gopkg.in/mgo.v2"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/mgo"
	"github.com/andeya/pholcus/common/pool"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
)

// --- MongoDB Output ---

func init() {
	DataOutput["mgo"] = func(col *Collector) result.VoidResult {
		if mgo.Error() != nil {
			mgo.Refresh()
			if mgo.Error() != nil {
				return result.FmtErrVoid("MongoDB connection failed: %v", mgo.Error())
			}
		}
		return mgo.Call(func(src pool.Src) error {
			var (
				db          = src.(*mgo.MgoSrc).DB(config.Conf().DBName)
				namespace   = util.FileNameReplace(col.namespace())
				collections = make(map[string]*mgov2.Collection)
				dataMap     = make(map[string][]interface{})
				err         error
			)

			for _, datacell := range col.dataBuf {
				subNamespace := util.FileNameReplace(col.subNamespace(datacell))
				cName := joinNamespaces(namespace, subNamespace)

				if _, ok := collections[subNamespace]; !ok {
					collections[subNamespace] = db.C(cName)
				}
				for k, v := range datacell["Data"].(map[string]interface{}) {
					datacell[k] = v
				}
				delete(datacell, "Data")
				delete(datacell, "RuleName")
				if !col.Spider.OutDefaultField() {
					delete(datacell, "Url")
					delete(datacell, "ParentUrl")
					delete(datacell, "DownloadTime")
				}
				dataMap[subNamespace] = append(dataMap[subNamespace], datacell)
			}

			for collection, docs := range dataMap {
				c := collections[collection]
				count := len(docs)
				loop := count / mgo.MaxLen
				for i := 0; i < loop; i++ {
					err = c.Insert(docs[i*mgo.MaxLen : (i+1)*mgo.MaxLen]...)
					if err != nil {
						logs.Log().Error("%v", err)
					}
				}
				if count%mgo.MaxLen == 0 {
					continue
				}
				err = c.Insert(docs[loop*mgo.MaxLen:]...)
				if err != nil {
					logs.Log().Error("%v", err)
				}
			}

			return nil
		})
	}
}
