package bot

import (
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/utils"
)

const WEB_LIST_QUOTES_PER_PAGE = 10
const BOT_LIST_QUOTES_PER_PAGE = 5

type listArgs struct {
	Author *discordgo.User `description:"Author to display quotes for. Omit to display quotes from all users."`
	Query  *string         `description:"Optional keyword / phrase to search for."`
	Page   int             `default:"1" description:"Page of quotes to display."`
}

func (b *Bot) listQuotesCommand(_ *discordgo.Session, interaction *discordgo.InteractionCreate, args listArgs) {
	opts := database.SearchOptions{
		Page:  args.Page,
		Limit: BOT_LIST_QUOTES_PER_PAGE,
		Query: args.Query,
	}

	if args.Author != nil {
		opts.Author = utils.PtrTo(args.Author.ID)
	}

	// TODO: Use quote count
	quotes, totalCount, err := database.Search(opts)
	if err != nil {
		b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error getting quotes.\n```\n%s\n```", err))
		return
	}

	webUrl := *config.BaseUrl
	webUrl.Path = "/list"
	newQuery := webUrl.Query()

	if args.Author != nil {
		newQuery.Set("author", args.Author.ID)
	}

	if args.Query != nil {
		newQuery.Set("query", *args.Query)
	}

	webPageNumber := int(math.Ceil(float64(args.Page*BOT_LIST_QUOTES_PER_PAGE) / float64(WEB_LIST_QUOTES_PER_PAGE)))
	if webPageNumber > 1 {
		newQuery.Set("page", fmt.Sprintf("%d", webPageNumber))
	}

	webUrl.RawQuery = newQuery.Encode()

	embed := discordgo.MessageEmbed{
		Title:       "Quotes",
		Color:       (40 << 16) + (120 << 8) + (120),
		Fields:      []*discordgo.MessageEmbedField{},
		Description: fmt.Sprintf("[View in browser](%s)", webUrl.String()),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page %d of %d", args.Page, int(math.Ceil(float64(totalCount)/float64(BOT_LIST_QUOTES_PER_PAGE)))),
		},
	}

	for _, quote := range quotes {
		authors, _, err := b.generateAuthorString(quote.Authors, interaction.GuildID)
		if err != nil {
			b.Session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("Error getting quote authors.\n```\n%s\n```", err))
		}

		quoteText := quote.Text
		if len(quoteText) >= 900 {
			quoteText = quoteText[:900] + "..."
		}

		quoteBody := fmt.Sprintf("%s\n\n_<t:%d>_", quoteText, quote.Meta.CreatedAt.UTC().Unix())
		if quote.Source != nil {
			quoteBody += fmt.Sprintf(" - [Source](%s)", *quote.Source)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%d - %s", quote.Meta.ID, authors),
			Value: quoteBody,
		})
	}

	b.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
}
