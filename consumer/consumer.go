package consumer

import (
	"fmt"
	"log"
	"time"
)

//MessageIterator is the consumer API.
type MessageIterator interface {
	//At each call returns the next batch of messages
	NextMessages() ([]Message, error)
}

//DefaultIterator is the default implementation of the MessageIterator interface.
//Calling the NewIterator(QueueConfig) a new instance of DefaultIterator is returned.
//NOTE: DefaultIterator is not thread-safe! If you call NextMessages() from different go routines concurrently, you doing it wrong.
type DefaultIterator struct {
	config   QueueConfig
	queue    queueCaller
	consumer *consumer
}

//QueueConfig represents the configuration of the queue, consumer group and topic the consumer interested about.
type QueueConfig struct {
	Addr  string `json:"address"`
	Group string `json:"group"`
	Topic string `json:"topic"`
	Queue string `json:"queue"`
}

//Message is the higher-level representation of messages from the queue.
type Message struct {
	Headers map[string]string
	Body    string
}

//NewIterator returns a pointer to a freshly created DefaultIterator.
func NewIterator(config QueueConfig) MessageIterator {
	queue := defaultQueueCaller{
		addr:   config.Addr,
		group:  config.Group,
		topic:  config.Topic,
		caller: defaultHTTPCaller{config.Queue},
	}
	return &DefaultIterator{config, queue, nil}
}

const backoffPeriod = 8

//NextMessages returns the next batch of messages from the queue.
func (c *DefaultIterator) NextMessages() (msgs []Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Error: recovered from panic: %v", r)
			}
		}
	}()
	msgs, err = c.consume()
	time.Sleep(time.Duration(backoffPeriod) * time.Second)
	return msgs, err
}

func (c *DefaultIterator) consume() ([]Message, error) {
	q := c.queue
	if c.consumer == nil {
		cInst, err := q.createConsumerInstance()
		if err != nil {
			log.Printf("ERROR - creating consumer instance: %s", err.Error())
			return nil, err
		}
		c.consumer = &cInst
	}
	msgs, err := q.consumeMessages(*c.consumer)
	if err != nil {
		log.Printf("ERROR - consuming messages: %s", err.Error())
		cInst := *c.consumer
		c.consumer = nil
		errD := q.destroyConsumerInstance(cInst)
		if errD != nil {
			log.Printf("ERROR - deleting consumer instance: %s", errD.Error())
		}
		return nil, err
	}
	return msgs, nil
}