package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nint8835/scribe/database"
)

func GenerateAuthorString(authors []*database.Author, guildID string) (string, error) {
	authorNames := []string{}

	for _, author := range authors {
		var name string
		if guildID != "" {
			user, err := Bot.GuildMember(guildID, author.ID)
			if err != nil {
				return "", fmt.Errorf("error getting guild member %s: %w", author.ID, err)
			}

			if user.Nick != "" {
				name = fmt.Sprintf("%s (%s#%s)", user.Nick, user.User.Username, user.User.Discriminator)
			} else {
				name = fmt.Sprintf("%s#%s", user.User.Username, user.User.Discriminator)
			}
		} else {
			user, err := Bot.User(author.ID)
			if err != nil {
				return "", fmt.Errorf("error getting user %s: %w", author.ID, err)
			}

			name = fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
		}
		authorNames = append(authorNames, name)
	}

	return strings.Join(authorNames, ", "), nil
}

func MakeQuoteEmbed(quote *database.Quote, guildID string) (*discordgo.MessageEmbed, error) {
	authors, err := GenerateAuthorString(quote.Authors, guildID)
	if err != nil {
		return &discordgo.MessageEmbed{}, fmt.Errorf("error getting quote authors: %w", err)
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Quote %d", quote.Meta.ID),
		Description: quote.Text,
		Color:       (80 << 16) + (40 << 8) + (200),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Author(s)",
				Value: authors,
			},
		},
	}, nil
}
