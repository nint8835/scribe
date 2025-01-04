package bot

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/pkg/config"
)

func (b *Bot) dbCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args struct{}) {
	dbFile, err := os.Open(config.Instance.DBPath)
	if err != nil {
		b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error opening database.\n```\n%s\n```", err))
		return
	}

	err = b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: []*discordgo.File{
				{
					Name:        "quotes.sqlite",
					ContentType: "application/x-sqlite3",
					Reader:      dbFile,
				},
			},
		},
	})

	if err != nil {
		b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error sending database.\n```\n%s\n```", err))
	}
}
