package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"pkg.nit.so/switchboard"

	"github.com/nint8835/scribe/pkg/config"
)

type Bot struct {
	Session *discordgo.Session

	parser   *switchboard.Switchboard
	quitChan chan struct{}
}

var Instance *Bot

func (b *Bot) Run() error {
	b.Session.AddHandler(b.parser.HandleInteractionCreate)
	b.registerCommands()

	err := b.parser.SyncCommands(b.Session, config.Instance.AppId)
	if err != nil {
		return fmt.Errorf("error syncing commands: %w", err)
	}

	if err = b.Session.Open(); err != nil {
		return fmt.Errorf("error opening Discord connection: %w", err)
	}

	<-b.quitChan
	fmt.Println("Stopping bot...")

	if err = b.Session.Close(); err != nil {
		return fmt.Errorf("error closing Discord connection: %w", err)
	}

	return nil
}

func (b *Bot) Stop() {
	b.quitChan <- struct{}{}
}

func (b *Bot) registerCommands() {
	// Slash commands
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "add",
		Description: "Add a new quote.",
		Handler:     b.addQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "get",
		Description: "Display an individual quote by ID.",
		Handler:     b.getQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "random",
		Description: "Get a random quote.",
		Handler:     b.randomQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "list",
		Description: "List quotes.",
		Handler:     b.listQuotesCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "remove",
		Description: "Remove a quote.",
		Handler:     b.removeQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "edit",
		Description: "Edit a quote.",
		Handler:     b.editQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "db",
		Description: "Get a copy of the Scribe database.",
		Handler:     b.dbCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "cancelmulti",
		Description: "Cancel the current multi-message quote.",
		Handler:     b.cancelMultiMessageQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:        "savemulti",
		Description: "Save the current multi-message quote.",
		Handler:     b.saveMultiMessageQuoteCommand,
		GuildID:     config.Instance.GuildId,
	})

	// Message commands
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:    "Quote Message",
		Handler: b.addQuoteMessageCommand,
		GuildID: config.Instance.GuildId,
		Type:    switchboard.MessageCommand,
	})
	_ = b.parser.AddCommand(&switchboard.Command{
		Name:    "Add to Multi-Message Quote",
		Handler: b.addToMultiMessageQuoteCommand,
		GuildID: config.Instance.GuildId,
		Type:    switchboard.MessageCommand,
	})
}

func New() (*Bot, error) {
	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.Instance.Token))
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	bot := &Bot{
		Session:  session,
		quitChan: make(chan struct{}),
		parser:   &switchboard.Switchboard{},
	}

	return bot, nil
}
