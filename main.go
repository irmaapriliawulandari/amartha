package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"amartha-test/handler"
	"amartha-test/helper"
	"amartha-test/repo"
	"amartha-test/usecase"
)

// panicHandler stops the service on any handler panic instead of
// letting the request fail silently while the process keeps running.
func panicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v\n%s", rec, debug.Stack())
				os.Exit(1)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := helper.InitDB()
	if err != nil {
		panic(err)
	}
	log.Println("db connected")

	repo := repo.NewLoanRepo(db)

	billingEngineUsecase := usecase.NewBillingEngineUsecase(repo)

	billingEngineHandler := handler.NewBillingEngineHandler(billingEngineUsecase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, billingEngineHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: panicHandler(mux),
	}

	go func() {
		log.Println("listening on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
