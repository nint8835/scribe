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

// Cache to store input/output pairs
var cache = make(map[string]string)
var cacheMutex sync.Mutex // Mutex to handle concurrent access to the cache

// GetGlobalName queries the DiscordLookup API and retrieves the global_name.
func GetGlobalName(id string) (string, error) {
	// Check if the result is already in the cache
	cacheMutex.Lock()
	if cachedName, found := cache[id]; found {
		cacheMutex.Unlock()
		return cachedName, nil
	}
	cacheMutex.Unlock()

	// Construct the API URL
	url := fmt.Sprintf("https://discordlookup.mesalytic.moe/v1/user/%s", id)

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a non-200 status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the JSON response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %w", err)
	}

	// Extract the global_name key
	name, ok := responseData["global_name"].(string)
	if !ok {
		// Extract the username key if that fails
		name, ok = responseData["username"].(string)
		if !ok {
			return "", fmt.Errorf("global_name or username key not found or not a string")
		}
	}

	// Store the result in the cache
	cacheMutex.Lock()
	cache[id] = name
	cacheMutex.Unlock()

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

		authorIDs := make([]string, len(quote.Authors))
		for i, author := range quote.Authors {
			authorName, err := GetGlobalName(author.ID)

			if err == nil {
				authorIDs[i] = authorName
			} else {
				authorIDs[i] = author.ID
			}
		}

		joinedIDs := strings.Join(authorIDs, ", ")

		formattedQuotes[i] = pages.LeaderboardQuote{
			Author:  joinedIDs,
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
