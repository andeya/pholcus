package kafka

import (
	"strings"
	"sync"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"

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

// GetProducer returns the Kafka sync producer and any initialization error.
func GetProducer() result.Result[sarama.SyncProducer] {
	return result.Ret(producer, err)
}

// Refresh initializes or reconnects the Kafka producer.
func Refresh() {
	once.Do(func() {
		conf := sarama.NewConfig()
		conf.Producer.RequiredAcks = sarama.WaitForAll
		conf.Producer.Retry.Max = 10
		brokerList := config.Conf().Kafka.Brokers
		producer, err = sarama.NewSyncProducer(strings.Split(brokerList, ","), conf)
		if err != nil {
			logs.Log().Error("Kafka: %v\n", err)
		}
	})
}

// New creates a new KafkaSender.
func New() *KafkaSender {
	return &KafkaSender{}
}

// SetTopic sets the Kafka topic for sending messages.
func (p *KafkaSender) SetTopic(topic string) {
	p.topic = topic
}

// Push sends data as a JSON message to the configured topic.
func (p *KafkaSender) Push(data map[string]interface{}) result.VoidResult {
	val := util.JSONString(data)
	_, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(val),
	})
	return result.RetVoid(err)
}
