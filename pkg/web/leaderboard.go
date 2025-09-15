package web

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

const LEADERBOARD_QUOTES_PER_PAGE = 10

// Number of comparisons a quote must have to be included in the leaderboard
const USER_LEADERBOARD_COMPARISON_THRESHOLD = 10

// Number of quotes a user must have to be included in the leaderboard
const USER_LEADERBOARD_QUOTE_THRESHOLD = 5

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

	if err := query.Offset((page - 1) * LEADERBOARD_QUOTES_PER_PAGE).Limit(LEADERBOARD_QUOTES_PER_PAGE).Find(&quotes).Error; err != nil {
		return fmt.Errorf("error fetching quotes: %w", err)
	}

	formattedQuotes := make([]pages.LeaderboardQuote, len(quotes))
	for i, quote := range quotes {
		content, err := s.renderQuoteText(quote)
		if err != nil {
			return fmt.Errorf("error rendering quote: %w", err)
		}

		authorNames, err := s.formatAuthors(quote)
		if err != nil {
			return fmt.Errorf("error formatting author names: %w", err)
		}

		formattedQuotes[i] = pages.LeaderboardQuote{
			Author:  authorNames,
			Content: content,
			Elo:     quote.Elo,
			Rank:    (page-1)*LEADERBOARD_QUOTES_PER_PAGE + i + 1,
		}
	}

	props := pages.LeaderboardProps{
		Quotes:     formattedQuotes,
		Page:       page,
		TotalPages: int((total + LEADERBOARD_QUOTES_PER_PAGE - 1) / LEADERBOARD_QUOTES_PER_PAGE),
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.LeaderboardContent(props).Render(r.Context(), w)
	} else {
		pages.Leaderboard(props).Render(r.Context(), w)
	}

	return nil
}

func (s *Server) handleGetUserLeaderboard(w http.ResponseWriter, r *http.Request) error {
	var users []struct {
		AuthorID   string  `gorm:"column:author_id"`
		AvgElo     float64 `gorm:"column:avg_elo"`
		QuoteCount int     `gorm:"column:quote_count"`
	}

	err := database.Instance.Raw(`WITH
		sufficiently_ranked_quotes AS (
			SELECT
				quotes.id,
				COUNT(completed_comparisons.id) AS comparison_count
			FROM
				quotes
				LEFT JOIN completed_comparisons ON (
					completed_comparisons.quote_a_id = quotes.id
					OR completed_comparisons.quote_b_id = quotes.id
				)
			GROUP BY
				quotes.id
			HAVING
				comparison_count >= ?
		)
		SELECT
			quote_authors.author_id,
			AVG(quotes.elo) AS avg_elo,
			COUNT(quotes.id) AS quote_count
		FROM
			quote_authors
			JOIN sufficiently_ranked_quotes ON sufficiently_ranked_quotes.id = quote_authors.quote_id
			JOIN quotes ON quotes.id = sufficiently_ranked_quotes.id
		GROUP BY
			quote_authors.author_id
		HAVING
			quote_count >= ?
		ORDER BY
			avg_elo DESC`, USER_LEADERBOARD_COMPARISON_THRESHOLD, USER_LEADERBOARD_QUOTE_THRESHOLD).Scan(&users).Error
	if err != nil {
		return fmt.Errorf("error fetching users: %w", err)
	}

	formattedUsers := make([]pages.UserLeaderboardEntry, len(users))
	for i, user := range users {
		username, err := s.resolveAuthorIDs(fmt.Sprintf("<@%s>", user.AuthorID))
		if err != nil {
			return fmt.Errorf("error resolving author IDs: %w", err)
		}

		formattedUsers[i] = pages.UserLeaderboardEntry{
			Username: username,
			Elo:      user.AvgElo,
			Quotes:   user.QuoteCount,
		}
	}

	pages.UserLeaderboard(pages.UserLeaderboardProps{
		Users:          formattedUsers,
		RequiredRanks:  USER_LEADERBOARD_COMPARISON_THRESHOLD,
		RequiredQuotes: USER_LEADERBOARD_QUOTE_THRESHOLD,
	}).Render(r.Context(), w)

	return nil
}
