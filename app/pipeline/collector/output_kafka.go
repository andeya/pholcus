package collector

import (
	"regexp"
	"sync"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/kafka"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/logs"
)

// --- Kafka Output ---

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

	var topic = regexp.MustCompile("^[0-9a-zA-Z_-]+$")

	DataOutput["kafka"] = func(self *Collector) (r result.VoidResult) {
		defer r.Catch()
		kafka.GetProducer().Unwrap()
		var (
			kafkas    = make(map[string]*kafka.KafkaSender)
			namespace = util.FileNameReplace(self.namespace())
		)
		for _, datacell := range self.dataDocker {
			subNamespace := util.FileNameReplace(self.subNamespace(datacell))
			topicName := joinNamespaces(namespace, subNamespace)
			if !topic.MatchString(topicName) {
				logs.Log.Error("topic must match '^[0-9a-zA-Z_-]+$', got: %s", topicName)
				continue
			}
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
			sender.Push(data).Unwrap()
		}
		kafkas = nil
		return result.OkVoid()
	}
}
