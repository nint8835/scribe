package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
)

type removeArgs struct {
	ID int `description:"ID of the quote to remove."`
}

func (b *Bot) removeQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args removeArgs) {
	if interaction.Member.User.ID != config.Instance.OwnerId {
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	result = database.Instance.Delete(&quote)
	if result.Error != nil {
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
	b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
}
