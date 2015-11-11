package collector

import (
	"github.com/henrylee2cn/pholcus/common/util"
)

// 命名空间相对于数据库名，不依赖具体数据内容，可选
func (self *Collector) namespace() string {
	if self.Spider.Namespace == nil {
		if self.Spider.GetKeyword() == "" {
			return self.Spider.GetName()
		}
		if len([]rune(self.Spider.GetKeyword())) > 10 {
			return self.Spider.GetName() + "__" + util.MakeHash(self.Spider.GetKeyword())
		}
		return self.Spider.GetName() + "__" + self.Spider.GetKeyword()
	}
	return self.Spider.Namespace(self.Spider)
}

// 子命名空间相对于表名，可依赖具体数据内容，可选
func (self *Collector) subNamespace(dataCell map[string]interface{}) string {
	if self.Spider.SubNamespace == nil {
		return dataCell["RuleName"].(string)
	}
	return self.Spider.SubNamespace(self.Spider, dataCell)
}
