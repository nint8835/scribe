package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

const QUOTES_PER_PAGE = 10

var nameCache = make(map[string]string)
var nameCacheMutex sync.Mutex

func GetGlobalName(id string) (string, error) {
	nameCacheMutex.Lock()
	if cachedName, found := nameCache[id]; found {
		nameCacheMutex.Unlock()
		return cachedName, nil
	}
	nameCacheMutex.Unlock()

	url := fmt.Sprintf("https://discordlookup.mesalytic.moe/v1/user/%s", id)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %w", err)
	}

	name, ok := responseData["global_name"].(string)
	if !ok {
		name, ok = responseData["username"].(string)
		if !ok {
			return "", fmt.Errorf("global_name or username key not found or not a string")
		}
	}

	nameCacheMutex.Lock()
	nameCache[id] = name
	nameCacheMutex.Unlock()

	return name, nil
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

		authorNames := make([]string, len(quote.Authors))
		for i, author := range quote.Authors {
			authorName, err := GetGlobalName(author.ID)

			if err == nil {
				authorNames[i] = authorName
			} else {
				authorNames[i] = author.ID
			}
		}

		formattedQuotes[i] = pages.LeaderboardQuote{
			Author:  strings.Join(authorNames, ", "),
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
