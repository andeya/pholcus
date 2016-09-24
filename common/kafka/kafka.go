package kafka

import (
	"strings"
	"sync"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"

	"github.com/Shopify/sarama"
)

var (
	err      error
	producer sarama.SyncProducer
	lock     sync.RWMutex
	once     sync.Once
)

type KafkaSender struct {
	topic string
}

func GetProducer() (sarama.SyncProducer, error) {
	return producer, err
}

//刷新producer
func Refresh() {
	once.Do(func() {
		conf := sarama.NewConfig()
		conf.Producer.RequiredAcks = sarama.WaitForAll //等待所有备份返回ack
		conf.Producer.Retry.Max = 10                   // 重试次数
		brokerList := config.KAFKA_BORKERS
		producer, err = sarama.NewSyncProducer(strings.Split(brokerList, ","), conf)
		if err != nil {
			logs.Log.Error("Kafka:%v\n", err)
		}
	})
}

func New() *KafkaSender {
	return &KafkaSender{}
}

func (p *KafkaSender) SetTopic(topic string) {
	p.topic = topic
}

func (p *KafkaSender) Push(data map[string]interface{}) error {
	val := util.JsonString(data)
	_, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(val),
	})
	return err
}
