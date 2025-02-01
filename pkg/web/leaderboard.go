package web

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

const QUOTES_PER_PAGE = 10

func (s *Server) resolveAuthorIDs(ids string) (string, error) {
	var buf bytes.Buffer
	err := s.md.Convert([]byte(ids), &buf)
	return buf.String(), err
}

func (s *Server) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) error {
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil {
			page = pageNum
		}
	}

	var quotes []database.Quote
	var total int64

	query := database.Instance.Model(&database.Quote{}).
		Preload("Authors").
		Joins("LEFT JOIN completed_comparisons ON quotes.id = completed_comparisons.quote_a_id OR quotes.id = completed_comparisons.quote_b_id").
		Group("quotes.id").
		Having("COUNT(completed_comparisons.id) > 0").
		Order("elo desc")

	if err := query.Count(&total).Error; err != nil {
		return fmt.Errorf("error counting quotes: %w", err)
	}

	if err := query.Offset((page - 1) * QUOTES_PER_PAGE).Limit(QUOTES_PER_PAGE).Find(&quotes).Error; err != nil {
		return fmt.Errorf("error fetching quotes: %w", err)
	}

	formattedQuotes := make([]pages.LeaderboardQuote, len(quotes))
	for i, quote := range quotes {
		content, err := s.renderQuoteText(quote)
		if err != nil {
			return fmt.Errorf("error rendering quote: %w", err)
		}

		authorIDs := make([]string, len(quote.Authors))
		for i, author := range quote.Authors {
			authorIDs[i] = author.ID
		}
		authorNames, err := s.resolveAuthorIDs("<@" + strings.Join(authorIDs, ">, <@") + ">")

		formattedQuotes[i] = pages.LeaderboardQuote{
			Author:  authorNames,
			Content: content,
			Elo:     quote.Elo,
			Rank:    (page-1)*QUOTES_PER_PAGE + i + 1,
		}
	}

	props := pages.LeaderboardProps{
		Quotes:     formattedQuotes,
		Page:       page,
		TotalPages: int((total + QUOTES_PER_PAGE - 1) / QUOTES_PER_PAGE),
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.LeaderboardContent(props).Render(r.Context(), w)
	} else {
		pages.Leaderboard(props).Render(r.Context(), w)
	}

	return nil
}
