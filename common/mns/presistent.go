package mns

import (
	"github.com/souriki/ali_mns"
)

type MNSPresistent struct {
	queue ali_mns.AliMNSQueue
}

func (self *MNSPresistent) Push(bs []byte) {
	msg := ali_mns.MessageSendRequest{
		MessageBody:  string(bs),
		DelaySeconds: 0,
		Priority:     1,
	}
	self.queue.SendMessage(msg)

}
func (self *MNSPresistent) Pull() (bs []byte) {
	respChan := make(chan ali_mns.MessageReceiveResponse)
	errChan := make(chan error)
	endChan := make(chan int)
	go func() {
		select {
		case resp := <-respChan:
			{
				if ret, e := self.queue.ChangeMessageVisibility(resp.ReceiptHandle, 5); e != nil {
					endChan <- 1
					return
				} else {
					if e := self.queue.DeleteMessage(ret.ReceiptHandle); e != nil {
						endChan <- 1
						return
					}

				}
				bs = []byte(resp.MessageBody)
				endChan <- 1
				return
			}
		case <-errChan:
			{
				endChan <- 1
				return
			}

		}
	}()
	self.queue.ReceiveMessage(respChan, errChan, 2)
	<-endChan
	return
}

type MNSPresistentFactory struct {
	Prefix       string
	MNSClient    ali_mns.MNSClient
	QueueManager ali_mns.AliQueueManager
}

func NewMNSPresistentFactory(prefix string, c ali_mns.MNSClient) *MNSPresistentFactory {
	f := &MNSPresistentFactory{Prefix: prefix, MNSClient: c}
	f.QueueManager = ali_mns.NewMNSQueueManager(c)
	return f
}

func (self *MNSPresistentFactory) New(name string) (matrix *MNSPresistent, err error) {
	queueName := self.Prefix + name
	err = self.QueueManager.CreateSimpleQueue(queueName)
	// 队列已经存在
	if err != nil && !ali_mns.ERR_MNS_QUEUE_ALREADY_EXIST_AND_HAVE_SAME_ATTR.IsEqual(err) {
		return
	}
	// Init
	err = nil
	queue := ali_mns.NewMNSQueue(queueName, self.MNSClient)
	matrix = &MNSPresistent{queue}
	return
}
