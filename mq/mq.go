package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
)

const (
	defaultTopic       = "billing-engine"
	defaultChannel     = "billing-engine-consumer"
	defaultNSQDTCPAddr = "127.0.0.1:4150"
)

// billingHandler processes each message received from the topic/channel.
// Replace with real billing engine logic.
type billingHandler struct{}

func (h *billingHandler) HandleMessage(m *nsq.Message) error {
	log.Printf("received message: %s", string(m.Body))
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	topic := envOrDefault("MQ_TOPIC", defaultTopic)
	channel := envOrDefault("MQ_CHANNEL", defaultChannel)
	nsqdAddr := envOrDefault("NSQD_TCP_ADDR", defaultNSQDTCPAddr)

	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		log.Fatal(err)
	}

	consumer.AddHandler(&billingHandler{})

	if err := consumer.ConnectToNSQD(nsqdAddr); err != nil {
		log.Fatal(err)
	}
	log.Printf("consuming topic=%q channel=%q via nsqd=%q", topic, channel, nsqdAddr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("stopping consumer")
	consumer.Stop()
	<-consumer.StopChan
}
