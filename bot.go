package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/nint8835/parsley"

	"github.com/nint8835/scribe/database"
)

type Config struct {
	DBPath  string `default:"quotes.sqlite" split_words:"true"`
	Token   string
	Prefix  string `default:"q!"`
	OwnerId string `default:"106162668032802816" split_words:"true"`
}

var config Config
var Bot *discordgo.Session

func Run() error {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Failed to load .env file: %s\n", err.Error())
	}

	err = envconfig.Process("scribe", &config)
	if err != nil {
		return fmt.Errorf("error loading config: %s", err)
	}

	database.Initialize(config.DBPath)
	database.Migrate()

	Bot, err = discordgo.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}
	Bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	parser := parsley.New(config.Prefix)
	parser.RegisterHandler(Bot)
	RegisterCommands(parser)

	if err = Bot.Open(); err != nil {
		return fmt.Errorf("error opening Discord connection: %w", err)
	}

	fmt.Println("Scribe is now running. Press Ctrl-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("Quitting Scribe")

	if err = Bot.Close(); err != nil {
		return fmt.Errorf("error closing Discord connection: %w", err)
	}

	return nil
}

func main() {
	err := Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running Scribe: %s", err)
		os.Exit(1)
	}
}
