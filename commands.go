package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"
	"pkg.nit.so/switchboard"

	"github.com/nint8835/scribe/database"
)

var MentionListRegexp = regexp.MustCompile(`<@!?(\d{17,})>`)

type AddArgs struct {
	Text   string         `description:"Text for the quote to add. To insert a new line, insert \\n."`
	Author discordgo.User `description:"Author of the quote."`
	Source *string        `description:"Link to a source for the quote, if available (such as a Discord message, screenshot, etc.)"`
}

func AddQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args AddArgs) {
	quote := database.Quote{
		Text:    strings.Replace(args.Text, "\\n", "\n", -1),
		Authors: []*database.Author{{ID: args.Author.ID}},
		Source:  args.Source,
	}

	result := database.Instance.Create(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error),
			},
		})
	}

	result = database.Instance.Save(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error),
			},
		})
	}

	embed, err := MakeQuoteEmbed(&quote, interaction.GuildID)
	if err != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error generating quote embed.\n```\n%s\n```", err),
			},
		})
		return
	}
	embed.Title = fmt.Sprintf("Quote %d added!", quote.Meta.ID)
	embed.Color = (45 << 16) + (200 << 8) + (95)

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	})
}

func AddQuoteMessageCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, message *discordgo.Message) {
	if message.Content == "" {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Color:       (240 << 16) + (85 << 8) + (125),
						Title:       "Error adding quote.",
						Description: "You cannot quote an empty message.",
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	quoteUrl := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", message.GuildID, message.ChannelID, message.ID)

	quote := database.Quote{
		Text:    message.Content,
		Authors: []*database.Author{{ID: message.Author.ID}},
		Source:  &quoteUrl,
	}

	result := database.Instance.Create(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error),
			},
		})
	}

	result = database.Instance.Save(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error),
			},
		})
	}

	embed, err := MakeQuoteEmbed(&quote, interaction.GuildID)
	if err != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error generating quote embed.\n```\n%s\n```", err),
			},
		})
		return
	}
	embed.Title = fmt.Sprintf("Quote %d added!", quote.Meta.ID)
	embed.Color = (45 << 16) + (200 << 8) + (95)

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	})
}

type GetArgs struct {
	ID int `description:"ID of the quote to display."`
}

func GetQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args GetArgs) {
	var quote database.Quote

	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).First(&quote, args.ID)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	embed, err := MakeQuoteEmbed(&quote, interaction.GuildID)
	if err != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", err),
			},
		})
		return
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func RandomQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	var quotes []database.Quote

	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).Find(&quotes)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quotes.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	quote := quotes[rand.Intn(len(quotes))]

	embed, err := MakeQuoteEmbed(&quote, interaction.GuildID)
	if err != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", err),
			},
		})
		return
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

type ListArgs struct {
	Author *discordgo.User `description:"Author to display quotes for. Omit to display quotes from all users."`
	Page   int             `default:"1" description:"Page of quotes to display."`
}

func ListQuotesCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args ListArgs) {
	var quotes []database.Quote

	query := database.Instance.Model(&database.Quote{}).Preload(clause.Associations)

	if args.Author != nil {
		query = query.
			Joins("INNER JOIN quote_authors ON quote_authors.quote_id = quotes.id").
			Where(map[string]interface{}{"quote_authors.author_id": args.Author.ID})
	}

	result := query.Limit(5).Offset(5 * (args.Page - 1)).Find(&quotes)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quotes.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	embed := discordgo.MessageEmbed{
		Title:  "Quotes",
		Color:  (40 << 16) + (120 << 8) + (120),
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, quote := range quotes {
		authors, _, err := GenerateAuthorString(quote.Authors, interaction.GuildID)
		if err != nil {
			Bot.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error getting quote authors.\n```\n%s\n```", result.Error))
		}

		quoteText := quote.Text
		if len(quoteText) >= 900 {
			quoteText = quoteText[:900] + "..."
		}

		quoteBody := fmt.Sprintf("%s\n\n_<t:%d>_", quoteText, quote.Meta.CreatedAt.UTC().Unix())
		if quote.Source != nil {
			quoteBody += fmt.Sprintf(" - [Source](%s)", *quote.Source)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%d - %s", quote.Meta.ID, authors),
			Value: quoteBody,
		})
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
}

type RemoveArgs struct {
	ID int `description:"ID of the quote to remove."`
}

func RemoveQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args RemoveArgs) {
	if interaction.Member.User.ID != config.OwnerId {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not have access to that command.",
			},
		})
		return
	}

	var quote database.Quote
	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).First(&quote, args.ID)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	result = database.Instance.Delete(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error deleting quote.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	embed := discordgo.MessageEmbed{
		Title:       "Quote deleted!",
		Description: fmt.Sprintf("Quote %d has been deleted succesfully.", args.ID),
		Color:       (240 << 16) + (85 << 8) + (125),
	}
	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
}

type EditArgs struct {
	ID   int    `description:"ID of the quote to edit."`
	Text string `description:"New text for the quote."`
}

func EditQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args EditArgs) {
	if interaction.Member.User.ID != config.OwnerId {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not have access to that command.",
			},
		})
		return
	}

	var quote database.Quote
	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).First(&quote, args.ID)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	quote.Text = strings.Replace(args.Text, "\\n", "\n", -1)

	result = database.Instance.Save(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error editing quote.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Quote edited!",
					Color:       (45 << 16) + (200 << 8) + (95),
					Description: "The quote has been edited successfully.",
				},
			},
		},
	})
}

type SearchArgs struct {
	Query string `description:"Keyword / phrase to search for."`
	Page  int    `default:"1" description:"Page of results to display."`
}

func SearchQuotesCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args SearchArgs) {
	if result := database.Instance.Exec("PRAGMA case_sensitive_like = OFF", nil); result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error enabling case-insensitive like.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	var quotes []database.Quote

	query := database.Instance.Model(&database.Quote{}).
		Preload(clause.Associations)

	if strings.Contains(args.Query, "%") {
		query = query.Where("text LIKE ?", args.Query)
	} else {
		query = query.Where("text LIKE ?", "% "+args.Query+" %").
			Or("text LIKE ?", "% "+args.Query).
			Or("text LIKE ?", args.Query+" %")
	}

	result := query.
		Limit(5).
		Offset(int(5 * (args.Page - 1))).
		Find(&quotes)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quotes.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	embed := discordgo.MessageEmbed{
		Title:  "Quotes",
		Color:  (40 << 16) + (120 << 8) + (120),
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, quote := range quotes {
		authors, _, err := GenerateAuthorString(quote.Authors, interaction.GuildID)
		if err != nil {
			Bot.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error getting quote authors.\n```\n%s\n```", result.Error))
		}

		quoteText := quote.Text
		if len(quoteText) >= 900 {
			quoteText = quoteText[:900] + "..."
		}

		quoteBody := fmt.Sprintf("%s\n\n_<t:%d>_", quoteText, quote.Meta.CreatedAt.UTC().Unix())
		if quote.Source != nil {
			quoteBody += fmt.Sprintf(" - [Source](%s)", *quote.Source)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%d - %s", quote.Meta.ID, authors),
			Value: quoteBody,
		})
	}

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
}

func DbCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args struct{}) {
	dbFile, err := os.Open(config.DBPath)
	if err != nil {
		Bot.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error opening database.\n```\n%s\n```", err))
		return
	}

	err = Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: []*discordgo.File{
				{
					Name:        "quotes.sqlite",
					ContentType: "application/x-sqlite3",
					Reader:      dbFile,
				},
			},
		},
	})

	if err != nil {
		Bot.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error sending database.\n```\n%s\n```", err))
	}
}

func RegisterCommands(parser *switchboard.Switchboard) {
	_ = parser.AddCommand(&switchboard.Command{
		Name:        "add",
		Description: "Add a new quote.",
		Handler:     AddQuoteCommand,
		GuildID:     config.GuildId,
	})
	_ = parser.AddCommand(&switchboard.Command{
		Name:    "Quote Message",
		Handler: AddQuoteMessageCommand,
		GuildID: config.GuildId,
		Type:    switchboard.MessageCommand,
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
}
