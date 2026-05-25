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
	query := database.Instance.
		Model(&database.Quote{}).
		Where("strftime('%m', created_at) = ? AND strftime('%d', created_at) = ?", now.Format("01"), now.Format("02"))

	var quoteCount int64
	result := query.Count(&quoteCount)
	if result.Error != nil {
		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error counting quotes.\n```\n%s\n```", result.Error),
			},
		})
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	var quote database.Quote
	result = query.
		Preload(clause.Associations).
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
	quoteCountLabel := "quotes"
	if quoteCount == 1 {
		quoteCountLabel = "quote"
	}
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("%d %s from this day in history", quoteCount, quoteCountLabel),
	}

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
