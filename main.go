package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nint8835/scribe/pkg/bot"
	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	database.Initialize(config.Instance.DBPath)
	database.Migrate()

	botInst, err := bot.New()
	if err != nil {
		log.Fatalf("Error creating bot: %s", err)
	}

	go func() {
		err = botInst.Run()
		if err != nil {
			log.Fatalf("Error running bot: %s", err)
		}
	}()

	fmt.Println("Scribe is now running. Press Ctrl-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	botInst.Stop()
}
