package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/database"
)

type ListArgs struct {
	Author *discordgo.User `description:"Author to display quotes for. Omit to display quotes from all users."`
	Query  *string         `description:"Optional keyword / phrase to search for."`
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

	if args.Query != nil {
		queryString := *args.Query

		filterQuery := database.Instance.Raw("SELECT ROWID FROM quotes_fts WHERE quotes_fts MATCH ?", queryString)
		query.Where("quotes.ROWID IN (?)", filterQuery)
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
