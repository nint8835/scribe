package bot

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
)

func sqliteTimeZoneModifier(t time.Time) string {
	return t.Format("-07:00")
}

func onThisDayQuery(query *gorm.DB, day time.Time) *gorm.DB {
	day = day.In(config.Location)
	timeZoneModifier := sqliteTimeZoneModifier(day)

	return query.Where(
		"strftime('%m', created_at, ?) = ? AND strftime('%d', created_at, ?) = ? AND strftime('%Y', created_at, ?) != ?",
		timeZoneModifier,
		day.Format("01"),
		timeZoneModifier,
		day.Format("02"),
		timeZoneModifier,
		day.Format("2006"),
	)
}

func (b *Bot) makeOnThisDayEmbed(guildID string) (*discordgo.MessageEmbed, error) {
	now := time.Now().In(config.Location)
	query := onThisDayQuery(database.Instance.Model(&database.Quote{}), now)
	var quotes []database.Quote
	result := query.
		Preload(clause.Associations).
		Order("created_at ASC").
		Find(&quotes)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting quotes: %w", result.Error)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes found from this day in history (%s): %w", now.Format("January 2"), gorm.ErrRecordNotFound)
	}

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("On This Day: %s", now.Format("January 2")),
		Color:  (80 << 16) + (40 << 8) + 200,
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, quote := range quotes {
		authors, _, err := b.generateAuthorString(quote.Authors, guildID)
		if err != nil {
			return nil, fmt.Errorf("error getting quote authors: %w", err)
		}

		field := quoteListField(quote, authors)
		footer := &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Showing %d of %d quotes from this day in history", len(embed.Fields)+1, len(quotes))}
		embed.Fields = append(embed.Fields, field)
		embed.Footer = footer
		if len(embed.Fields) > discordEmbedFieldLimit || embedTextLength(embed) > discordEmbedTotalLimit {
			embed.Fields = embed.Fields[:len(embed.Fields)-1]
			break
		}
	}

	quoteLabel := "quotes"
	if len(quotes) == 1 {
		quoteLabel = "quote"
	}
	if len(embed.Fields) == len(quotes) {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("%d %s from this day in history", len(quotes), quoteLabel)}
	} else {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Showing %d of %d quotes from this day in history", len(embed.Fields), len(quotes))}
	}

	return embed, nil
}

func (b *Bot) onThisDayQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	now := time.Now().In(config.Location)

	embed, err := b.makeOnThisDayEmbed(interaction.GuildID)
	if err != nil {
		var content string
		if errors.Is(err, gorm.ErrRecordNotFound) {
			content = fmt.Sprintf("No quotes found from this day in history (%s).", now.Format("January 2"))
		} else {
			content = fmt.Sprintf("Error getting quote.\n```\n%s\n```", err)
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
