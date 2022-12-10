package main

import (
	"fmt"
	"math/rand"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/parsley"
	"gorm.io/gorm/clause"

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
	var quotes []database.Quote

	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).Find(&quotes)
	if result.Error != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quotes.\n```\n%s\n```", result.Error))
	}

	quote := quotes[rand.Intn(len(quotes))]

	embed, err := MakeQuoteEmbed(&quote, message.GuildID)
	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quote.\n```\n%s\n```", err))
	}

	Bot.ChannelMessageSendEmbed(message.ChannelID, embed)
}

type ListArgs struct {
	Authors string `description:"Author(s) to display, as a list of Discord mentions, or all for all quotes."`
	Page    uint   `default:"1" description:"Page of quotes to display."`
}

func ListQuotesCommand(message *discordgo.MessageCreate, args ListArgs) {
	var quotes []database.Quote

	query := database.Instance.Model(&database.Quote{}).Preload(clause.Associations)

	if args.Authors != "all" {
		authors := []string{}
		authorMatches := MentionListRegexp.FindAllStringSubmatch(args.Authors, -1)

		for _, match := range authorMatches {
			authors = append(authors, match[1])
		}

		query = query.
			Joins("INNER JOIN quote_authors ON quote_authors.quote_id = quotes.id", authorMatches).
			Where(map[string]interface{}{"quote_authors.author_id": authors})
	}

	result := query.Limit(5).Offset(int(5 * (args.Page - 1))).Find(&quotes)
	if result.Error != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quotes.\n```\n%s\n```", result.Error))
	}

	embed := discordgo.MessageEmbed{
		Title:  "Quotes",
		Color:  (40 << 16) + (120 << 8) + (120),
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, quote := range quotes {
		authors, _, err := GenerateAuthorString(quote.Authors, message.GuildID)
		if err != nil {
			Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Error getting quote authors.\n```\n%s\n```", result.Error))
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

	Bot.ChannelMessageSendEmbed(message.ChannelID, &embed)
}

type RemoveArgs struct {
	ID uint `description:"ID of the quote to remove."`
}

func RemoveQuoteCommand(message *discordgo.MessageCreate, args RemoveArgs) {
	if message.Author.ID != config.OwnerId {
		Bot.ChannelMessageSendReply(message.ChannelID, "You do not have access to that command.", message.Reference())
		return
	}

	var quote database.Quote
	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).First(&quote, args.ID)
	if result.Error != nil {
		Bot.ChannelMessageSendReply(message.ChannelID, fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error), message.Reference())
		return
	}

	result = database.Instance.Delete(&quote)
	if result.Error != nil {
		Bot.ChannelMessageSendReply(message.ChannelID, fmt.Sprintf("Error deleting quote.\n```\n%s\n```", result.Error), message.Reference())
	}

	embed := discordgo.MessageEmbed{
		Title:       "Quote deleted!",
		Description: fmt.Sprintf("Quote %d has been deleted succesfully.", args.ID),
		Color:       (240 << 16) + (85 << 8) + (125),
	}
	Bot.ChannelMessageSendEmbedReply(message.ChannelID, &embed, message.Reference())
}

type EditArgs struct {
	ID   uint   `description:"ID of the quote to edit."`
	Text string `description:"New text for the quote."`
}

func EditQuoteCommand(message *discordgo.MessageCreate, args EditArgs) {
	if message.Author.ID != config.OwnerId {
		Bot.ChannelMessageSendReply(message.ChannelID, "You do not have access to that command.", message.Reference())
		return
	}

	var quote database.Quote
	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).First(&quote, args.ID)
	if result.Error != nil {
		Bot.ChannelMessageSendReply(message.ChannelID, fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error), message.Reference())
		return
	}

	quote.Text = args.Text

	result = database.Instance.Save(&quote)
	if result.Error != nil {
		Bot.ChannelMessageSendReply(message.ChannelID, fmt.Sprintf("Error editing quote.\n```\n%s\n```", result.Error), message.Reference())
		return
	}

	Bot.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		Title:       "Quote edited!",
		Color:       (45 << 16) + (200 << 8) + (95),
		Description: "The quote has been edited successfully.",
	})
}

func RegisterCommands(parser *parsley.Parser) {
	parser.NewCommand("add", "Add a new quote.", AddQuoteCommand)
	parser.NewCommand("get", "Display an individual quote by ID.", GetQuoteCommand)
	parser.NewCommand("random", "Get a random quote.", RandomQuoteCommand)
	parser.NewCommand("list", "List quotes.", ListQuotesCommand)
	parser.NewCommand("remove", "Remove a quote.", RemoveQuoteCommand)
	parser.NewCommand("edit", "Edit a quote.", EditQuoteCommand)
}
