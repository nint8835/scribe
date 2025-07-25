package bot

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/pkg/database"
)

type randomArgs struct {
	Author *discordgo.User `description:"Author to pick a random quote from. Omit to pick a quote from any user."`
}

func (b *Bot) randomQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args randomArgs) {
	var quotes []database.Quote

	query := database.Instance.Model(&database.Quote{}).Preload(clause.Associations)

	if args.Author != nil {
		query = query.
			Joins("INNER JOIN quote_authors ON quote_authors.quote_id = quotes.id").
			Where("quote_authors.author_id = ?", args.Author.ID)
	}

	result := query.Find(&quotes)
	if result.Error != nil {
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quotes.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	quote := quotes[rand.Intn(len(quotes))]

	embed, err := b.makeQuoteEmbed(&quote, interaction.GuildID)
	if err != nil {
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", err),
			},
		})
		return
	}

	b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
