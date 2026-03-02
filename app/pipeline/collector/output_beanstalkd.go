package collector

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/beanstalkd"
	"github.com/andeya/pholcus/common/util"
)

// --- Beanstalkd Output ---

func init() {
	DataOutput["beanstalkd"] = func(self *Collector) (r result.VoidResult) {
		defer r.Catch()
		client := beanstalkd.New().Unwrap()
		defer client.Close()

		namespace := fmt.Sprintf("%v__%v-%v", util.FileNameReplace(self.namespace()), self.sum[0], self.sum[1])
		createtime := fmt.Sprintf("%d", time.Now().Unix())

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
			result.RetVoid(err).Unwrap()
			data.Add("createtime", createtime)
			data.Add("type", namespace+"__"+subNamespace)
			data.Add("content", string(res))
			client.Send(data).Unwrap()
		}
		return result.OkVoid()
	}
}
