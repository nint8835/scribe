package main

import (
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"github.com/nint8835/scribe/database"
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

func RegisterCommands(parser *parsley.Parser) {
	parser.NewCommand("add", "Add a new quote.", AddQuoteCommand)
}
