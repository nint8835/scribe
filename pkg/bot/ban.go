package bot

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/pkg/database"
)

type banArgs struct {
	User discordgo.User `description:"User to ban from adding quotes."`
}

func (b *Bot) banCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args banArgs) {
	if b.ensureCanDeleteQuotes(interaction) {
		return
	}

	banned := database.BannedUser{ID: args.User.ID}
	result := database.Instance.FirstOrCreate(&banned, banned)
	if result.Error != nil {
		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error banning user.\n```\n%s\n```", result.Error),
			},
		})
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	embed := discordgo.MessageEmbed{
		Title:       "User banned",
		Description: fmt.Sprintf("%s has been banned from adding quotes.", args.User.Mention()),
		Color:       (240 << 16) + (85 << 8) + (125),
	}
	err := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
	if err != nil {
		slog.Error("error sending interaction response", "error", err)
	}
}