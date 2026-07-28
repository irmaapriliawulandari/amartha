package helper

import (
	"fmt"

	"github.com/nsqio/go-nsq"
)

var (
	nsqdAddr = "127.0.0.1:9150"
)

// NSQPublisher wraps an nsq.Producer to publish messages to an nsqd instance.
type NSQPublisher struct {
	producer *nsq.Producer
}

// NewNSQPublisher connects to the nsqd instance at addr (e.g. "localhost:4150")
// and returns a ready-to-use NSQPublisher.
func NewNSQPublisher() (*NSQPublisher, error) {
	producer, err := nsq.NewProducer(nsqdAddr, nsq.NewConfig())
	if err != nil {
		return nil, fmt.Errorf("create nsq producer: %w", err)
	}

	if err := producer.Ping(); err != nil {
		return nil, fmt.Errorf("ping nsqd at %s: %w", nsqdAddr, err)
	}

	return &NSQPublisher{producer: producer}, nil
}

// Publish sends body to the given NSQ topic.
func (p *NSQPublisher) Publish(topic string, body []byte) error {
	if err := p.producer.Publish(topic, body); err != nil {
		return fmt.Errorf("publish to topic %s: %w", topic, err)
	}

	return nil
}

// Stop gracefully closes the underlying connection to nsqd.
func (p *NSQPublisher) Stop() {
	p.producer.Stop()
}
