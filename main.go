package main

import (
	"log"

	"github.com/nint8835/scribe/pkg/bot"
)

func main() {
	err := bot.Run()
	if err != nil {
		log.Fatalf("Error running Scribe: %s", err)
	}
}
