package main

import (
	"flag"
	"log"

	"goshop/internal/bootstrap"
)

func main() {
	configPath := flag.String("config", "config.yaml", "config file path")
	flag.Parse()

	app, err := bootstrap.NewApp(*configPath)
	if err != nil {
		log.Fatalf("bootstrap app: %v", err)
	}
	defer app.Close()

	if err := app.Server.Run(); err != nil {
		log.Fatalf("server run: %v", err)
	}
}
