package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/database"
)

func AddQuoteMessageCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, message *discordgo.Message) {
	if message.Content == "" {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Color:       (240 << 16) + (85 << 8) + (125),
						Title:       "Error adding quote.",
						Description: "You cannot quote an empty message.",
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	quoteUrl := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", message.GuildID, message.ChannelID, message.ID)

	quote := database.Quote{
		Text:    message.Content,
		Authors: []*database.Author{{ID: message.Author.ID}},
		Source:  &quoteUrl,
	}

	addQuote(quote, interaction)
}
