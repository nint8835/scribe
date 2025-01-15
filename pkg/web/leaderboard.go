package web

import (
	"net/http"
	"strconv"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

const QUOTES_PER_PAGE = 10

func (s *Server) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil {
			page = pageNum
		}
	}

	var quotes []database.Quote
	var total int64

	query := database.Instance.Model(&database.Quote{}).Where("elo != 1000").Order("elo desc")

	if err := query.Count(&total).Error; err != nil {
		http.Error(w, "Error fetching quotes", http.StatusInternalServerError)
		return
	}

	if err := query.Offset((page - 1) * QUOTES_PER_PAGE).Limit(QUOTES_PER_PAGE).Find(&quotes).Error; err != nil {
		http.Error(w, "Error fetching quotes", http.StatusInternalServerError)
		return
	}

	formattedQuotes := make([]pages.LeaderboardQuote, len(quotes))
	for i, quote := range quotes {
		content, err := s.renderQuoteText(quote)
		if err != nil {
			http.Error(w, "Error rendering quotes", http.StatusInternalServerError)
			return
		}
		formattedQuotes[i] = pages.LeaderboardQuote{
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
}
