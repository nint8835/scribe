package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/embedding"
)

type editArgs struct {
	ID   int    `description:"ID of the quote to edit."`
	Text string `description:"New text for the quote."`
}

func (b *Bot) editQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args editArgs) {
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

	quote.Text = strings.Replace(args.Text, "\\n", "\n", -1)

	result = database.Instance.Save(&quote)
	if result.Error != nil {
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error editing quote.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	encodedEmbedding, err := embedding.EmbedQuote(quote.Text)
	if err != nil {
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error creating quote embedding.\n```\n%s\n```", err),
			},
		})
		return
	}

	err = database.Instance.Exec(
		"UPDATE quote_embeddings SET embedding = ? WHERE rowid = ?",
		encodedEmbedding,
		quote.Meta.ID,
	).Error
	if err != nil {
		b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error saving quote embedding.\n```\n%s\n```", err),
			},
		})
		return
	}

	b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Quote edited!",
					Color:       (45 << 16) + (200 << 8) + (95),
					Description: "The quote has been edited successfully.",
				},
			},
		},
	})
}
