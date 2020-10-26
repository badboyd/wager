package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wager/config"
	"wager/internal/app"
	"wager/internal/repository/postgres"

	"github.com/jmoiron/sqlx"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Panicf("Cannot load configuration: %s\n", err.Error())
	}

	dbConfig := fmt.Sprintf("user=%s dbname=%s host=%s port=%d sslmode=disable",
		cfg.Database.Username, cfg.Database.Database, cfg.Database.Host, cfg.Database.Port)
	log.Printf("Init db with these param %v", dbConfig)

	app := app.New(postgres.New(sqlx.MustConnect("postgres", dbConfig)))

	// run app in another routine
	go func() {
		if err := app.Run(cfg.Service.Port); err != nil {
			log.Printf("app run failed: %s\n", err.Error())
		}
	}()

	// graceful shutdown will be handled here
	// wait for the signal
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, os.Kill)

	log.Printf("Received signal %s", <-ch)
	defer cancel()

	if err = app.Close(ctx); err != nil {
		panic(err)
	}
}
