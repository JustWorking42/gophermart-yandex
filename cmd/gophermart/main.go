package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JustWorking42/gophermart-yandex/internal/app/app"
	"github.com/JustWorking42/gophermart-yandex/internal/app/config"
	"github.com/JustWorking42/gophermart-yandex/internal/app/handlers"
	"github.com/JustWorking42/gophermart-yandex/internal/app/updater"
)

func main() {
	mainContext, MainCancel := context.WithCancel(context.Background())
	defer MainCancel()

	config, err := config.ConfigService()

	if err != nil {
		log.Fatalf("ConfigService init err: %v err", err)
	}

	app, err := app.NewApp(mainContext, config.DatabaseURI, config.LogLevel)

	if err != nil {
		log.Fatalf("App init err: %v err", err)
	}

	updater := updater.NewUpdater(app.Repository, config.AccrualAdress)
	wg, errChan := updater.SubcribeOnTask(mainContext)

	go func() {
		wg.Wait()
	}()

	server := http.Server{
		Addr:         config.ServerAdress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handlers.Webhooks(app),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("ListenAndServe error: %v\n", err)
		}
	}()

	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Printf("Error: %v", err)
			}
		case <-stop:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				fmt.Printf("Server Shutdown Failed:%+v", err)
			} else {
				fmt.Println("Server stopped")
			}
			return
		}
	}
}
