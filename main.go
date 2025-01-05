package main

import (
	"log"

	"github.com/nint8835/scribe/pkg/bot"
	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	database.Initialize(config.Instance.DBPath)
	database.Migrate()

	var botInst *bot.Bot

	if !config.Instance.RunBot {
		botInst, err = bot.New()
		if err != nil {
			log.Fatalf("Error creating bot: %s", err)
		}

		go func() {
			err = botInst.Run()
			if err != nil {
				log.Fatalf("Error running bot: %s", err)
			}
		}()
	}

	webServer, err := web.New()
	if err != nil {
		log.Fatalf("Error creating web server: %s", err)
	}

	err = webServer.Run()
	if err != nil {
		log.Fatalf("Error running web server: %s", err)
	}

	if botInst != nil {
		botInst.Stop()
	}
}
