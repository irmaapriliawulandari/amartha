package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	handlercron "amartha-test/handler/cron"
	"amartha-test/helper"
	"amartha-test/repo"
	"amartha-test/usecase"

	"github.com/robfig/cron/v3"
)

func main() {
	db, err := helper.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("db connected")

	loanRepo := repo.NewLoanRepo(db)
	billingEngineUsecase := usecase.NewBillingEngineUsecase(loanRepo)

	c := cron.New()

	handler := handlercron.NewCronHandler(billingEngineUsecase)
	if _, err := c.AddFunc("0 1 * * *", handler.MarkOverdue); err != nil {
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
