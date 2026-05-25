package bot

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/pkg/database"
)

func (b *Bot) onThisDayQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	now := time.Now()

	var quote database.Quote
	result := database.Instance.
		Model(&database.Quote{}).
		Preload(clause.Associations).
		Where("strftime('%m', created_at) = ? AND strftime('%d', created_at) = ?", now.Format("01"), now.Format("02")).
		Order("RANDOM()").
		Take(&quote)
	if result.Error != nil {
		var content string
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			content = fmt.Sprintf("No quotes found from this day in history (%s).", now.Format("January 2"))
		} else {
			content = fmt.Sprintf("Error getting quote.\n```\n%s\n```", result.Error)
		}

		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	embed, err := b.makeQuoteEmbed(&quote, interaction.GuildID)
	if err != nil {
		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quote.\n```\n%s\n```", err),
			},
		})
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	embed.Title = fmt.Sprintf("On This Day: %s", now.Format("January 2"))

	err = b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		slog.Error("error sending interaction response", "error", err)
	}
}
