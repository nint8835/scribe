package bot

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"

	"github.com/nint8835/scribe/pkg/database"
)

var pendingMultiMessageQuotes = map[string][]*discordgo.Message{}

func generateWIPMultiMessageQuoteEmbed(memberId string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       (95 << 16) + (155 << 8) + (200),
		Title:       "Pending Multi-Message Quote",
		Description: "`/savemulti` to save, `/cancelmulti` to cancel.",
	}

	for _, message := range pendingMultiMessageQuotes[memberId] {
		messageTitle := message.Author.Username

		// Include a link to the message if this is a real message
		if message.ID != "" {
			messageTitle += fmt.Sprintf(" (%s)", GenerateMessageUrl(message))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   messageTitle,
			Value:  message.Content,
			Inline: false,
		})
	}

	return embed
}

func (b *Bot) addToMultiMessageQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, message *discordgo.Message) {
	memberId := interaction.Member.User.ID

	if message.Content == "" {
		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	if _, ok := pendingMultiMessageQuotes[memberId]; !ok {
		pendingMultiMessageQuotes[memberId] = []*discordgo.Message{}
	}

	for _, existingMessage := range pendingMultiMessageQuotes[memberId] {
		if existingMessage.ID == message.ID {
			respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
			if respondErr != nil {
				slog.Error("error sending interaction response", "error", respondErr)
			}
			return
		}
	}

	pendingMultiMessageQuotes[memberId] = append(pendingMultiMessageQuotes[memberId], message)
	sort.Slice(pendingMultiMessageQuotes[memberId], func(i, j int) bool {
		return pendingMultiMessageQuotes[memberId][i].Timestamp.Before(pendingMultiMessageQuotes[memberId][j].Timestamp)
	})

	respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{generateWIPMultiMessageQuoteEmbed(memberId)},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if respondErr != nil {
		slog.Error("error sending interaction response", "error", respondErr)
	}
}

type slashAddMultiMessageQuoteArgs struct {
	Text   string         `description:"Text for the line to add. To insert a new line, insert \\n."`
	Author discordgo.User `description:"Author of the line."`
}

func (b *Bot) slashAddMultiMessageQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args slashAddMultiMessageQuoteArgs) {
	memberId := interaction.Member.User.ID

	if _, ok := pendingMultiMessageQuotes[memberId]; !ok {
		pendingMultiMessageQuotes[memberId] = []*discordgo.Message{}
	}

	// Create a fake message for the data provided, to enable re-using all the same quote generation logic
	lineMessage := &discordgo.Message{
		Content:   strings.ReplaceAll(args.Text, "\\n", "\n"),
		Author:    &args.Author,
		Timestamp: time.Now(),
	}

	pendingMultiMessageQuotes[memberId] = append(pendingMultiMessageQuotes[memberId], lineMessage)
	sort.Slice(pendingMultiMessageQuotes[memberId], func(i, j int) bool {
		return pendingMultiMessageQuotes[memberId][i].Timestamp.Before(pendingMultiMessageQuotes[memberId][j].Timestamp)
	})

	respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{generateWIPMultiMessageQuoteEmbed(memberId)},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if respondErr != nil {
		slog.Error("error sending interaction response", "error", respondErr)
	}
}

func (b *Bot) cancelMultiMessageQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	memberId := interaction.Member.User.ID

	if _, ok := pendingMultiMessageQuotes[memberId]; !ok {
		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	delete(pendingMultiMessageQuotes, memberId)

	respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
	if respondErr != nil {
		slog.Error("error sending interaction response", "error", respondErr)
	}
}

func (b *Bot) saveMultiMessageQuoteCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, _ struct{}) {
	memberId := interaction.Member.User.ID

	if _, ok := pendingMultiMessageQuotes[memberId]; !ok {
		respondErr := b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
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
		if respondErr != nil {
			slog.Error("error sending interaction response", "error", respondErr)
		}
		return
	}

	var quoteUrl *string

	if pendingMultiMessageQuotes[memberId][0].ID != "" {
		url := GenerateMessageUrl(pendingMultiMessageQuotes[memberId][0])
		quoteUrl = &url
	}

	authorIds := map[string]bool{}

	for _, message := range pendingMultiMessageQuotes[memberId] {
		authorIds[message.Author.ID] = true
	}

	var authors []*database.Author

	for authorId := range authorIds {
		authors = append(authors, &database.Author{ID: authorId})
	}

	var quoteContent strings.Builder

	for _, message := range pendingMultiMessageQuotes[memberId] {
		if len(authors) > 1 {
			fmt.Fprintf(&quoteContent, "%s: %s\n", message.Author.Mention(), message.Content)
		} else {
			quoteContent.WriteString(message.Content + "\n")
		}
	}

	quote := database.Quote{
		Text:    quoteContent.String(),
		Authors: authors,
		Source:  quoteUrl,
		Meta: gorm.Model{
			CreatedAt: pendingMultiMessageQuotes[memberId][0].Timestamp,
		},
	}

	b.addQuote(quote, interaction)

	delete(pendingMultiMessageQuotes, memberId)
}
