package main

import (
	"log"

	"wager/internal/app"
)

func main() {
	app := app.New()
	if app == nil {
		panic("Cannot init app")
	}

	// app will be run, graceful shutdown will be handled inside the Run function
	if err := app.Run(); err != nil {
		log.Printf("app run failed: %s\n", err.Error())
	}
}
