package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
)

// runBillingJob is the scheduled job body. Replace with real billing engine logic.
func runBillingJob() {
	log.Println("running billing job")
}

func main() {
	c := cron.New()

	if _, err := c.AddFunc("@daily", runBillingJob); err != nil {
		log.Fatal(err)
	}

	c.Start()
	log.Println("cron started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("stopping cron")
	ctx := c.Stop()
	<-ctx.Done()
}
