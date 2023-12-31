package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"pkg.nit.so/switchboard"

	"github.com/nint8835/scribe/database"
)

type Config struct {
	DBPath  string `default:"quotes.sqlite" split_words:"true"`
	Token   string
	OwnerId string `default:"106162668032802816" split_words:"true"`
	GuildId string `default:"497544520695808000" split_words:"true"`
	AppId   string `default:"862525831552172045" split_words:"true"`
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

	parser := &switchboard.Switchboard{}
	Bot.AddHandler(parser.HandleInteractionCreate)
	RegisterCommands(parser)
	err = parser.SyncCommands(Bot, config.AppId)
	if err != nil {
		return fmt.Errorf("error syncing commands: %w", err)
	}

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

func RegisterCommands(parser *switchboard.Switchboard) {
	// Slash commands
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "add",
		Description: "Add a new quote.",
		Handler:     AddQuoteCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "get",
		Description: "Display an individual quote by ID.",
		Handler:     GetQuoteCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "random",
		Description: "Get a random quote.",
		Handler:     RandomQuoteCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "list",
		Description: "List quotes.",
		Handler:     ListQuotesCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "remove",
		Description: "Remove a quote.",
		Handler:     RemoveQuoteCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "edit",
		Description: "Edit a quote.",
		Handler:     EditQuoteCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "search",
		Description: "Search for quotes.",
		Handler:     SearchQuotesCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "db",
		Description: "Get a copy of the Scribe database.",
		Handler:     DbCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "cancelmulti",
		Description: "Cancel the current multi-message quote.",
		Handler:     CancelMultiMessageQuoteCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "savemulti",
		Description: "Save the current multi-message quote.",
		Handler:     SaveMultiMessageQuoteCommand,
		GuildID:     config.GuildId,
	})

	// Message commands
	_ = parser.AddCommand(&switchboard.Command{
		Name:    "Quote Message",
		Handler: AddQuoteMessageCommand,
		GuildID: config.GuildId,
		Type:    switchboard.MessageCommand,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:    "Add to Multi-Message Quote",
		Handler: AddToMultiMessageQuoteCommand,
		GuildID: config.GuildId,
		Type:    switchboard.MessageCommand,
	})
}

func main() {
	err := Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running Scribe: %s\n", err)
		os.Exit(1)
	}
}
