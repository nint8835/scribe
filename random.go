package main

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/database"
)

func RandomQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	var quotes []database.Quote

	result := database.Instance.Model(&database.Quote{}).Preload(clause.Associations).Find(&quotes)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error getting quotes.\n```\n%s\n```", result.Error),
			},
		})
		return
	}

	quote := quotes[rand.Intn(len(quotes))]

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
