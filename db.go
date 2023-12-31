package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

func DbCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args struct{}) {
	dbFile, err := os.Open(config.DBPath)
	if err != nil {
		Bot.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error opening database.\n```\n%s\n```", err))
		return
	}

	err = Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
		Bot.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error sending database.\n```\n%s\n```", err))
	}
}
