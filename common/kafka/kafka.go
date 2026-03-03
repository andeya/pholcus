// Package kafka 提供了 Kafka 消息队列的发送封装。
package kafka

import (
	"errors"
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

// KafkaSender 向指定 topic 发送消息。
type KafkaSender struct {
	topic string
}

// GetProducer 返回 Kafka 同步生产者及初始化错误。
func GetProducer() result.Result[sarama.SyncProducer] {
	return result.Ret(producer, err)
}

// Refresh 初始化或重连 Kafka 生产者。
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

// New 创建 KafkaSender 实例。
func New() *KafkaSender {
	return &KafkaSender{}
}

// SetTopic 设置发送消息的 topic。
func (p *KafkaSender) SetTopic(topic string) {
	p.topic = topic
}

// Push 将数据以 JSON 格式发送到已配置的 topic。
func (p *KafkaSender) Push(data map[string]interface{}) result.VoidResult {
	if producer == nil {
		return result.TryErrVoid(errors.New("kafka producer not initialized"))
	}
	val := util.JSONString(data)
	_, _, sendErr := producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(val),
	})
	return result.RetVoid(sendErr)
}
