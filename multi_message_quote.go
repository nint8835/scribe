package main

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/database"
)

var pendingMultiMessageQuotes = map[string][]*discordgo.Message{}

func generateWIPMultiMessageQuoteEmbed(memberId string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       (95 << 16) + (155 << 8) + (200),
		Title:       "Pending Multi-Message Quote",
		Description: "`/savemulti` to save, `/cancelmulti` to cancel.",
	}

	for _, message := range pendingMultiMessageQuotes[memberId] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (%s)", message.Author.Username, GenerateMessageUrl(message)),
			Value:  message.Content,
			Inline: false,
		})
	}

	return embed
}

func AddToMultiMessageQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, message *discordgo.Message) {
	memberId := interaction.Member.User.ID

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

	if _, ok := pendingMultiMessageQuotes[memberId]; !ok {
		pendingMultiMessageQuotes[memberId] = []*discordgo.Message{}
	}

	for _, existingMessage := range pendingMultiMessageQuotes[memberId] {
		if existingMessage.ID == message.ID {
			Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Color:       (240 << 16) + (85 << 8) + (125),
							Title:       "Error adding quote.",
							Description: "You cannot quote the same message twice.",
						},
					},
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
	}

	pendingMultiMessageQuotes[memberId] = append(pendingMultiMessageQuotes[memberId], message)
	sort.Slice(pendingMultiMessageQuotes[memberId], func(i, j int) bool {
		return pendingMultiMessageQuotes[memberId][i].Timestamp.Before(pendingMultiMessageQuotes[memberId][j].Timestamp)
	})

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{generateWIPMultiMessageQuoteEmbed(memberId)},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func CancelMultiMessageQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	memberId := interaction.Member.User.ID

	if _, ok := pendingMultiMessageQuotes[memberId]; !ok {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Color:       (240 << 16) + (85 << 8) + (125),
						Title:       "Error cancelling quote.",
						Description: "You do not have a pending multi-message quote.",
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	delete(pendingMultiMessageQuotes, memberId)

	Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Color:       (45 << 16) + (200 << 8) + (95),
					Title:       "Quote cancelled!",
					Description: "Your pending multi-message quote has been cancelled.",
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func SaveMultiMessageQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	memberId := interaction.Member.User.ID

	if _, ok := pendingMultiMessageQuotes[memberId]; !ok {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Color:       (240 << 16) + (85 << 8) + (125),
						Title:       "Error saving quote.",
						Description: "You do not have a pending multi-message quote.",
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	quoteUrl := GenerateMessageUrl(pendingMultiMessageQuotes[memberId][0])

	authorIds := map[string]bool{}

	for _, message := range pendingMultiMessageQuotes[memberId] {
		authorIds[message.Author.ID] = true
	}

	var authors []*database.Author

	for authorId := range authorIds {
		authors = append(authors, &database.Author{ID: authorId})
	}

	quoteContent := ""

	for _, message := range pendingMultiMessageQuotes[memberId] {
		if len(authors) > 1 {
			quoteContent += fmt.Sprintf("%s: %s\n", message.Author.Mention(), message.Content)
		} else {
			quoteContent += message.Content + "\n"
		}
	}

	quote := database.Quote{
		Text:    quoteContent,
		Authors: authors,
		Source:  &quoteUrl,
	}

	addQuote(quote, interaction)

	delete(pendingMultiMessageQuotes, memberId)
}
