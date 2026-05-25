package bot

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/pkg/config"
)

func (b *Bot) dbCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args struct{}) {
	err := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		_, sendErr := b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error acknowledging command.\n```\n%s\n```", err))
		if sendErr != nil {
			slog.Error("error sending channel message", "error", sendErr)
		}
		return
	}

	dbFile, err := os.Open(config.Instance.DBPath)
	if err != nil {
		content := fmt.Sprintf("Error opening database.\n```\n%s\n```", err)
		_, editErr := b.Session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if editErr != nil {
			slog.Error("error editing interaction response", "error", editErr, "original_error", err)
		}
		return
	}
	defer dbFile.Close()

	content := ""
	_, err = b.Session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Files: []*discordgo.File{
			{
				Name:        "quotes.sqlite",
				ContentType: "application/x-sqlite3",
				Reader:      dbFile,
			},
		},
	})

	if err != nil {
		content := fmt.Sprintf("Error sending database.\n```\n%s\n```", err)
		_, editErr := b.Session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if editErr != nil {
			slog.Error("error editing interaction response", "error", editErr, "original_error", err)
		}
	}
}
