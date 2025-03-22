package web

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/nint8835/scribe/pkg/bot"
	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web/ui/components"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

var nlLoc, _ = time.LoadLocation("America/St_Johns")

func generatePageURL(baseURL *url.URL, query url.Values, page int) string {
	newURL := *baseURL
	newURL.RawQuery = ""

	newQuery := query
	newQuery.Del("page")

	if page > 1 {
		newQuery.Set("page", strconv.Itoa(page))
	}

	newURL.RawQuery = newQuery.Encode()

	return newURL.String()
}

func (s *Server) handleGetList(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()

	page := 1

	if pageStr := q.Get("page"); pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err != nil {
			return fmt.Errorf("error parsing page number: %w", err)
		}

		page = pageInt
	}

	opts := database.SearchOptions{
		Page:  page,
		Limit: bot.WEB_LIST_QUOTES_PER_PAGE,
	}

	if author := q.Get("author"); author != "" {
		opts.Author = &author
	}

	if query := q.Get("query"); query != "" {
		opts.Query = &query
	}

	quotes, totalCount, err := database.Search(opts)
	if err != nil {
		return fmt.Errorf("error searching quotes: %w", err)
	}

	formattedQuotes := make([]pages.ListQuote, len(quotes))
	for i, quote := range quotes {
		content, err := s.renderQuoteText(quote)
		if err != nil {
			return fmt.Errorf("error rendering quote: %w", err)
		}

		authorMentions := make([]string, len(quote.Authors))
		for i, author := range quote.Authors {
			authorMentions[i] = fmt.Sprintf("<@%s>", author.ID)
		}
		authorNames, err := s.resolveAuthorIDs(strings.Join(authorMentions, ", "))
		if err != nil {
			return fmt.Errorf("error resolving author IDs: %w", err)
		}

		quoteLabel := fmt.Sprintf("#%d • %s • %s", quote.Meta.ID, authorNames, quote.Meta.CreatedAt.In(nlLoc).Format("January 2 2006, 3:04 PM"))

		if quote.Source != nil {
			var sourceLinkBuf bytes.Buffer
			err := s.md.Convert([]byte(fmt.Sprintf("[Source](%s)", *quote.Source)), &sourceLinkBuf)
			if err != nil {
				return fmt.Errorf("error converting source link: %w", err)
			}

			quoteLabel += " • " + sourceLinkBuf.String()
		}

		formattedQuotes[i] = pages.ListQuote{
			Content: content,
			Label:   quoteLabel,
		}
	}

	props := pages.ListProps{
		Quotes: formattedQuotes,
		PaginationProps: components.PaginationProps{
			UrlBase:     "/list",
			Target:      "#list-content",
			Page:        page,
			TotalPages:  int((totalCount + bot.WEB_LIST_QUOTES_PER_PAGE - 1) / bot.WEB_LIST_QUOTES_PER_PAGE),
			PrevPageUrl: generatePageURL(r.URL, q, page-1),
			NextPageUrl: generatePageURL(r.URL, q, page+1),
		},
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.ListContent(props).Render(r.Context(), w)
	} else {
		pages.List(props).Render(r.Context(), w)
	}

	return nil
}
