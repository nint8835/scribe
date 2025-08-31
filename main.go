package main

import (
	"log/slog"
	"os"

	"github.com/nint8835/scribe/pkg/bot"
	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/embedding"
	"github.com/nint8835/scribe/pkg/web"
)

func main() {
	err := config.Load()
	if err != nil {
		slog.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	err = embedding.Initialize()
	if err != nil {
		slog.Error("Error initializing embedding", "error", err)
		os.Exit(1)
	}

	database.Initialize(config.Instance.DBPath)

	botInst, err := bot.New()
	if err != nil {
		slog.Error("Error creating bot", "error", err)
		os.Exit(1)
	}

	if config.Instance.RunBot {
		go func() {
			err = botInst.Run()
			if err != nil {
				slog.Error("Error running bot", "error", err)
				os.Exit(1)
			}
		}()
	}

	webServer, err := web.New()
	if err != nil {
		slog.Error("Error creating web server", "error", err)
		os.Exit(1)
	}

	err = webServer.Run()
	if err != nil {
		slog.Error("Error running web server", "error", err)
		os.Exit(1)
	}

	if botInst != nil {
		botInst.Stop()
	}
}
