package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/embedding"
)

type semanticQueryArgs struct {
	Query string `description:"The query to search for."`
}

func (b *Bot) semanticQueryCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args semanticQueryArgs) {
	encodedEmbedding, err := embedding.EmbedQuote(context.TODO(), args.Query)
	if err != nil {
		_, sendErr := b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error creating embedding.\n```\n%s\n```", err))
		if sendErr != nil {
			slog.Error("error sending channel message", "error", sendErr)
		}
		return
	}

	var quotes []database.Quote

	err = database.Instance.Raw(
		`SELECT
			quotes.*
		FROM
			quote_embeddings
		JOIN
			quotes ON quotes.id = quote_embeddings.rowid
		WHERE
			quote_embeddings.embedding MATCH ?
			AND quote_embeddings.k = 5
			AND quotes.deleted_at IS NULL
		ORDER BY
			quote_embeddings.distance ASC
		`,
		encodedEmbedding,
	).Preload("Authors").Take(&quotes).Error

	if err != nil {
		_, sendErr := b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error getting quotes.\n```\n%s\n```", err))
		if sendErr != nil {
			slog.Error("error sending channel message", "error", sendErr)
		}
		return
	}

	embed := discordgo.MessageEmbed{
		Title:  "Quotes",
		Color:  (40 << 16) + (120 << 8) + (120),
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, quote := range quotes {
		authors, _, err := b.generateAuthorString(quote.Authors, interaction.GuildID)
		if err != nil {
			_, sendErr := b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error getting quote authors.\n```\n%s\n```", err))
			if sendErr != nil {
				slog.Error("error sending channel message", "error", sendErr)
			}
		}
		embed.Fields = append(embed.Fields, quoteListField(quote, authors))
	}

	err = b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
	if err != nil {
		slog.Error("error sending interaction response", "error", err)
	}
}
