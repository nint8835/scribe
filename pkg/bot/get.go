package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"

	database2 "github.com/nint8835/scribe/pkg/database"
)

type GetArgs struct {
	ID int `description:"ID of the quote to display."`
}

func GetQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args GetArgs) {
	var quote database2.Quote

	result := database2.Instance.Model(&database2.Quote{}).Preload(clause.Associations).First(&quote, args.ID)
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
