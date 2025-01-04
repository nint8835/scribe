package bot

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"

	database2 "github.com/nint8835/scribe/pkg/database"
)

type RandomArgs struct {
	Author *discordgo.User `description:"Author to pick a random quote from. Omit to pick a quote from any user."`
}

func RandomQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args RandomArgs) {
	var quotes []database2.Quote

	query := database2.Instance.Model(&database2.Quote{}).Preload(clause.Associations)

	if args.Author != nil {
		query = query.
			Joins("INNER JOIN quote_authors ON quote_authors.quote_id = quotes.id").
			Where(map[string]interface{}{"quote_authors.author_id": args.Author.ID})
	}

	result := query.Find(&quotes)
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
