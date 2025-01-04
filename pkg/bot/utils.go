package bot

import (
	"cmp"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/pkg/database"
)

func formatDiscordMember(member *discordgo.Member) string {
	nick := cmp.Or(member.Nick, member.User.GlobalName)

	if nick != "" && nick != member.User.Username {
		return fmt.Sprintf("%s (%s)", nick, member.User.Username)
	}

	return member.User.Username
}

func formatDiscordUser(user *discordgo.User) string {
	if user.GlobalName != "" && user.GlobalName != user.Username {
		return fmt.Sprintf("%s (%s)", user.GlobalName, user.Username)
	}

	return user.Username
}

func GenerateAuthorString(authors []*database.Author, guildID string) (string, string, error) {
	authorNames := []string{}

	for _, author := range authors {
		var name string
		if guildID != "" {
			member, err := Bot.GuildMember(guildID, author.ID)
			if err != nil {
				user, err := Bot.User(author.ID)
				if err != nil {
					return "", "", fmt.Errorf("error getting user %s: %w", author.ID, err)
				}

				name = formatDiscordUser(user)
			} else {
				name = formatDiscordMember(member)
			}
		} else {
			user, err := Bot.User(author.ID)
			if err != nil {
				return "", "", fmt.Errorf("error getting user %s: %w", author.ID, err)
			}

			name = formatDiscordUser(user)
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
