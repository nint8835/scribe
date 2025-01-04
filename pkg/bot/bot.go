package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"pkg.nit.so/switchboard"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
)

var Bot *discordgo.Session

func Run() error {
	err := config.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	database.Initialize(config.Instance.DBPath)
	database.Migrate()

	Bot, err = discordgo.New(fmt.Sprintf("Bot %s", config.Instance.Token))
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}
	Bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	parser := &switchboard.Switchboard{}
	Bot.AddHandler(parser.HandleInteractionCreate)
	RegisterCommands(parser)
	err = parser.SyncCommands(Bot, config.Instance.AppId)
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
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "get",
		Description: "Display an individual quote by ID.",
		Handler:     GetQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "random",
		Description: "Get a random quote.",
		Handler:     RandomQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "list",
		Description: "List quotes.",
		Handler:     ListQuotesCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "remove",
		Description: "Remove a quote.",
		Handler:     RemoveQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "edit",
		Description: "Edit a quote.",
		Handler:     EditQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "db",
		Description: "Get a copy of the Scribe database.",
		Handler:     DbCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "cancelmulti",
		Description: "Cancel the current multi-message quote.",
		Handler:     CancelMultiMessageQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "savemulti",
		Description: "Save the current multi-message quote.",
		Handler:     SaveMultiMessageQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})

	// Message commands
	_ = parser.AddCommand(&switchboard.Command{
		Name:    "Quote Message",
		Handler: AddQuoteMessageCommand,
		GuildID: config.Instance.GuildId,
		Type:    switchboard.MessageCommand,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:    "Add to Multi-Message Quote",
		Handler: AddToMultiMessageQuoteCommand,
		GuildID: config.Instance.GuildId,
		Type:    switchboard.MessageCommand,
	})
}
