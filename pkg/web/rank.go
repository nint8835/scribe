package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web/ui/components"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

func pickRandomQuotePair(ctx context.Context, userId string) (database.Quote, database.Quote, error) {
	var quoteA database.Quote
	var quoteB database.Quote

	err := database.Instance.WithContext(ctx).Model(&database.Quote{}).Order("RANDOM()").Take(&quoteA).Error
	if err != nil {
		return quoteA, quoteB, fmt.Errorf("error getting first random quote: %w", err)
	}

	// TODO: In the event there are no quotes left to compare for the given first quote, this will error - in the event this occurs, it should draw a new first quote
	// There's _probably_ a better way to do this, but with 1000 quotes it will take a long time for this to be a problem
	err = database.Instance.WithContext(ctx).Model(&database.Quote{}).
		Joins(
			"LEFT JOIN completed_comparisons ON (completed_comparisons.quote_a_id = quotes.id AND completed_comparisons.quote_b_id = ? AND completed_comparisons.user_id = ?) OR (completed_comparisons.quote_a_id = ? AND completed_comparisons.quote_b_id = quotes.id AND completed_comparisons.user_id = ?)",
			quoteA.Meta.ID,
			userId,
			quoteA.Meta.ID,
			userId,
		).
		Where("completed_comparisons.id IS NULL AND quotes.id != ?", quoteA.Meta.ID).
		Order("RANDOM()").Take(&quoteB).Error
	if err != nil {
		return quoteA, quoteB, fmt.Errorf("error getting second random quote: %w", err)
	}

	return quoteA, quoteB, nil
}

func getRankFormProps(quoteA database.Quote, quoteB database.Quote) components.RankProps {
	return components.RankProps{
		QuoteAID:      fmt.Sprintf("%d", quoteA.Meta.ID),
		QuoteBID:      fmt.Sprintf("%d", quoteB.Meta.ID),
		QuoteAContent: quoteA.Text,
		QuoteBContent: quoteB.Text,
	}
}

func (s *Server) handleGetRank(w http.ResponseWriter, r *http.Request) {
	userId := s.getCurrentUserId(r)

	quoteA, quoteB, err := pickRandomQuotePair(r.Context(), userId)
	if err != nil {
		http.Error(w, "Error getting quotes", http.StatusInternalServerError)
		return
	}

	pages.Rank(getRankFormProps(quoteA, quoteB)).Render(r.Context(), w)
}

func (s *Server) handlePostRank(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userId := s.getCurrentUserId(r)
	fmt.Printf("%#+v\n", r.PostForm)

	quoteA, quoteB, err := pickRandomQuotePair(r.Context(), userId)
	if err != nil {
		http.Error(w, "Error getting quotes", http.StatusInternalServerError)
		return
	}

	components.RankForm(getRankFormProps(quoteA, quoteB)).Render(r.Context(), w)
}
