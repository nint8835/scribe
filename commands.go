package main

import (
	"fmt"
	"math/rand"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/nint8835/scribe/database"
	"gorm.io/gorm/clause"
)

var MentionListRegexp = regexp.MustCompile(`<@!?(\d{17,})>`)

type AddArgs struct {
	Text    string `description:"Text for the quote to add. Can be multi-line, by wrapping in quotes."`
	Authors string `description:"Author(s) of the quote, as a list of Discord mentions."`
	Source  string `default:"" description:"Link to a source for the quote, if available (such as a Discord message, screenshot, etc.)"`
}

func AddQuoteCommand(message *discordgo.MessageCreate, args AddArgs) {
	addedAuthors := map[string]bool{}
	authors := []*database.Author{}
	authorMatches := MentionListRegexp.FindAllStringSubmatch(args.Authors, -1)

	for _, match := range authorMatches {
		if _, found := addedAuthors[match[1]]; found {
			continue
		}
		_, err := Bot.User(match[1])
		if err != nil {
			Bot.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
				Title:       "Error adding quote",
				Color:       (200 << 16) + (45 << 8) + (95),
				Description: "One or more of the provided authors is invalid.",
			})
			return
		}
		authors = append(authors, &database.Author{ID: match[1]})
		addedAuthors[match[1]] = true
	}

	if len(authors) == 0 {
		Bot.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
			Title:       "Error adding quote",
			Color:       (200 << 16) + (45 << 8) + (95),
			Description: "One or more quote authors must be provided.",
		})
		return
	}

	var source *string = nil

	if args.Source != "" {
		source = &args.Source
	}

	quote := database.Quote{
		Text:    args.Text,
		Authors: authors,
		Source:  source,
	}

	result := database.Instance.Create(&quote)
	if result.Error != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error))
	}

	result = database.Instance.Save(&quote)
	if result.Error != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error))
	}

	Bot.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		Title:       "Quote added!",
		Color:       (45 << 16) + (200 << 8) + (95),
		Description: fmt.Sprintf("Your quote was added. It has been assigned ID %d.", quote.Meta.ID),
	})
}

type GetArgs struct {
	ID uint `description:"ID of the quote to display."`
}

func GetQuoteCommand(message *discordgo.MessageCreate, args GetArgs) {
	var quote database.Quote

	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).First(&quote, args.ID)
	if result.Error != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error))
	}

	embed, err := MakeQuoteEmbed(&quote, message.GuildID)
	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quote.\n```\n%s\n```", err))
	}

	Bot.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func RandomQuoteCommand(message *discordgo.MessageCreate, args struct{}) {
	var quote database.Quote
	var quoteCount int64

	result := database.Instance.Model(&database.Quote{}).Count(&quoteCount)
	if result.Error != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting random quote number.\n```\n%s\n```", result.Error))
	}

	result = database.Instance.Model(&database.Quote{}).Preload(clause.Associations).First(&quote, rand.Intn(int(quoteCount+1)))
	if result.Error != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error))
	}

	embed, err := MakeQuoteEmbed(&quote, message.GuildID)
	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quote.\n```\n%s\n```", err))
	}

	Bot.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func RegisterCommands(parser *parsley.Parser) {
	parser.NewCommand("add", "Add a new quote.", AddQuoteCommand)
	parser.NewCommand("get", "Display an individual quote by ID.", GetQuoteCommand)
	parser.NewCommand("random", "Get a random quote.", RandomQuoteCommand)
}
