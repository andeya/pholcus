package collector

import (
	"fmt"
	"sync"

	"github.com/henrylee2cn/pholcus/common/kafka"
	"github.com/henrylee2cn/pholcus/common/util"
)

/************************ Kafka 输出 ***************************/
func init() {
	var (
		kafkaSenders    = map[string]*kafka.KafkaSender{}
		kafkaSenderLock sync.RWMutex
	)

	var getKafkaSender = func(name string) (*kafka.KafkaSender, bool) {
		kafkaSenderLock.RLock()
		tab, ok := kafkaSenders[name]
		kafkaSenderLock.RUnlock()
		return tab, ok
	}

	var setKafkaSender = func(name string, tab *kafka.KafkaSender) {
		kafkaSenderLock.Lock()
		kafkaSenders[name] = tab
		kafkaSenderLock.Unlock()
	}

	DataOutput["kafka"] = func(self *Collector) error {
		_, err := kafka.GetProducer()
		if err != nil {
			return fmt.Errorf("kafka producer失败: %v", err)
		}
		var (
			kafkas    = make(map[string]*kafka.KafkaSender)
			namespace = util.FileNameReplace(self.namespace())
		)
		for _, datacell := range self.dataDocker {
			subNamespace := util.FileNameReplace(self.subNamespace(datacell))
			topicName := joinNamespaces(namespace, subNamespace)
			sender, ok := kafkas[topicName]
			if !ok {
				sender, ok = getKafkaSender(topicName)
				if ok {
					kafkas[topicName] = sender
				} else {
					sender = kafka.New()
					sender.SetTopic(topicName)
					setKafkaSender(topicName, sender)
					kafkas[topicName] = sender
				}
			}
			data := make(map[string]interface{})
			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					data[title] = v
				} else {
					data[title] = util.JsonString(vd[title])
				}
			}
			if self.Spider.OutDefaultField() {
				data["url"] = datacell["Url"].(string)
				data["parent_url"] = datacell["ParentUrl"].(string)
				data["download_time"] = datacell["DownloadTime"].(string)
			}
			err := sender.Push(data)
			util.CheckErr(err)
		}
		kafkas = nil
		return nil
	}
}
