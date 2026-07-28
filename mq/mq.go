package main

import (
	"amartha-test/handler/mq"
	"amartha-test/helper"
	"amartha-test/repo"
	"amartha-test/usecase"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
)

const (
	defaultNSQDTCPAddr = "127.0.0.1:9150"
)

func main() {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic: %v\n%s", rec, debug.Stack())
			os.Exit(1)
		}
	}()

	db, err := helper.InitDB()
	if err != nil {
		panic(err)
	}
	log.Println("db connected")

	billingEngineRepo := repo.NewLoanRepo(db)
	billingEngineUC := usecase.NewBillingEngineUsecase(billingEngineRepo)

	consumers := []*nsq.Consumer{}
	consumers = append(consumers, register("disburse_loan", "billing_engine", mq.NewDisburseLoanHandler(billingEngineUC)))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("stopping consumer")
	for _, c := range consumers {
		c.Stop()
		<-c.StopChan
	}

}

func register(topic, channel string, handler nsq.Handler) *nsq.Consumer {
	config := nsq.NewConfig()
	config.DefaultRequeueDelay = 30 * time.Second
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		log.Fatal(err)
	}

	consumer.AddHandler(handler)

	if err := consumer.ConnectToNSQD(defaultNSQDTCPAddr); err != nil {
		log.Fatal(err)
	}
	log.Printf("consuming topic=%q channel=%q via nsqd=%q", topic, channel, defaultNSQDTCPAddr)

	return consumer
}
