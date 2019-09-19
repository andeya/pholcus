package collector

import (
	"fmt"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/common/beanstalkd"
	"net/url"
	"encoding/json"
	"time"
)

/************************ beanstalkd 输出 ***************************/
func init() {
	DataOutput["beanstalkd"] = func(self *Collector) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("%v", p)
			}
		}()

		client, err := beanstalkd.New()
		if err != nil {
			return err
		}
		defer client.Close()

		namespace := fmt.Sprintf("%v__%v-%v", util.FileNameReplace(self.namespace()), self.sum[0], self.sum[1])
		createtime := fmt.Sprintf("%d", time.Now().Unix())

		// 添加分类数据工作表
		for _, datacell := range self.dataDocker {
			var subNamespace = util.FileNameReplace(self.subNamespace(datacell))

			tmp := make(map[string]interface{})
			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					tmp[title] = v
				} else {
					tmp[title] = util.JsonString(vd[title])
				}
			}
			if self.Spider.OutDefaultField() {
				tmp["Url"] = datacell["Url"].(string)
				tmp["ParentUrl"] = datacell["ParentUrl"].(string)
				tmp["DownloadTime"] = datacell["DownloadTime"].(string)
			}

			data := url.Values{}
			res, err := json.Marshal(tmp)
			if err != nil {
				return err
			}

			data.Add("createtime", createtime)
			data.Add("type", namespace+"__"+subNamespace)
			data.Add("content", string(res))
			client.Send(data)
		}

		return
	}
}
