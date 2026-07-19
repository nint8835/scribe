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
	if b.ensureIsOwner(interaction) {
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

type unbanArgs struct {
	User discordgo.User `description:"User to unban from adding quotes."`
}

func (b *Bot) unbanCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args unbanArgs) {
	if b.ensureIsOwner(interaction) {
		return
	}

	unbanned, err := database.UnbanUser(args.User.ID)
	if err != nil {
		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error unbanning user.\n```\n%s\n```", err),
			},
		})
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	embed := discordgo.MessageEmbed{
		Color: (240 << 16) + (85 << 8) + (125),
	}
	if unbanned {
		embed.Title = "User unbanned"
		embed.Description = fmt.Sprintf("%s has been unbanned from adding quotes.", args.User.Mention())
	} else {
		embed.Title = "User not banned"
		embed.Description = fmt.Sprintf("%s was not banned from adding quotes.", args.User.Mention())
	}
	respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
	if respondErr != nil {
		slog.Error("error sending interaction response", "error", respondErr)
	}
}
