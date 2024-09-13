package main

import (
	"cmp"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/database"
)

func GenerateAuthorString(authors []*database.Author, guildID string) (string, string, error) {
	authorNames := []string{}

	for _, author := range authors {
		var name string
		if guildID != "" {
			user, err := Bot.GuildMember(guildID, author.ID)
			if err != nil {
				return "", "", fmt.Errorf("error getting guild member %s: %w", author.ID, err)
			}

			nick := cmp.Or(user.Nick, user.User.GlobalName)

			if nick != "" && nick != user.User.Username {
				name = fmt.Sprintf("%s (%s)", nick, user.User.Username)
			} else {
				name = user.User.Username
			}
		} else {
			user, err := Bot.User(author.ID)
			if err != nil {
				return "", "", fmt.Errorf("error getting user %s: %w", author.ID, err)
			}

			if user.GlobalName != "" && user.GlobalName != user.Username {
				name = fmt.Sprintf("%s (%s)", user.GlobalName, user.Username)
			} else {
				name = user.Username
			}
		}
		authorNames = append(authorNames, name)
	}

	var label string
	if len(authors) == 1 {
		label = "Author"
	} else {
		label = "Authors"
	}

	return strings.Join(authorNames, ", "), label, nil
}

func MakeQuoteEmbed(quote *database.Quote, guildID string) (*discordgo.MessageEmbed, error) {
	authors, authorLabel, err := GenerateAuthorString(quote.Authors, guildID)
	if err != nil {
		return &discordgo.MessageEmbed{}, fmt.Errorf("error getting quote authors: %w", err)
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:  authorLabel,
			Value: authors,
		},
	}

	if quote.Source != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "Source",
			Value: *quote.Source,
		})
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Quote %d", quote.Meta.ID),
		Description: quote.Text,
		Color:       (80 << 16) + (40 << 8) + (200),
		Fields:      fields,
		Timestamp:   quote.Meta.CreatedAt.Format(time.RFC3339),
	}, nil
}

func GenerateMessageUrl(message *discordgo.Message) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", message.GuildID, message.ChannelID, message.ID)
}

func addQuote(quote database.Quote, interaction *discordgo.InteractionCreate) {
	result := database.Instance.Create(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error),
			},
		})
	}

	result = database.Instance.Save(&quote)
	if result.Error != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding quote.\n```\n%s\n```", result.Error),
			},
		})
	}

	embed, err := MakeQuoteEmbed(&quote, interaction.GuildID)
	if err != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error generating quote embed.\n```\n%s\n```", err),
			},
		})
		return
	}
	embed.Title = fmt.Sprintf("Quote %d added!", quote.Meta.ID)
	embed.Color = (45 << 16) + (200 << 8) + (95)

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	})
}
