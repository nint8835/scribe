package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	elogo "github.com/kortemy/elo-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/web/ui/components"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

var elo *elogo.Elo = elogo.NewElo()

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

	quoteAID := r.PostForm.Get("quote_a_id")
	quoteBID := r.PostForm.Get("quote_b_id")
	winner := r.PostForm.Get("winner")

	if quoteAID == "" || quoteBID == "" || winner == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if quoteAID == quoteBID {
		http.Error(w, "Cannot rank the same quote against itself", http.StatusBadRequest)
		return
	}

	quoteAIDInt, err := strconv.ParseUint(quoteAID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid quote A ID", http.StatusBadRequest)
		return
	}

	quoteBIDInt, err := strconv.ParseUint(quoteBID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid quote B ID", http.StatusBadRequest)
		return
	}

	winnerInt, err := strconv.ParseUint(winner, 10, 64)
	if err != nil {
		http.Error(w, "Invalid winner", http.StatusBadRequest)
		return
	}

	err = database.Instance.Transaction(func(tx *gorm.DB) error {
		var quoteA database.Quote
		var quoteB database.Quote

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&database.Quote{}).Where("id = ?", quoteAIDInt).First(&quoteA).Error
		if err != nil {
			return fmt.Errorf("error getting quote A: %w", err)
		}

		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&database.Quote{}).Where("id = ?", quoteBIDInt).First(&quoteB).Error
		if err != nil {
			return fmt.Errorf("error getting quote B: %w", err)
		}

		// TODO: This query logs an error in the success case. See if there's a better way that won't result in an error log
		var existingComparison database.CompletedComparison
		err = tx.Model(&database.CompletedComparison{}).
			Where("(quote_a_id = ? AND quote_b_id = ? AND user_id = ?) OR (quote_a_id = ? AND quote_b_id = ? AND user_id = ?)", quoteAIDInt, quoteBIDInt, userId, quoteBIDInt, quoteAIDInt, userId).
			First(&existingComparison).Error
		if err == nil {
			return fmt.Errorf("comparison already exists")
		}

		err = tx.Create(&database.CompletedComparison{
			QuoteAID: quoteA.Meta.ID,
			QuoteBID: quoteB.Meta.ID,
			UserID:   userId,
			WinnerID: uint(winnerInt),
		}).Error
		if err != nil {
			return fmt.Errorf("error creating comparison: %w", err)
		}

		var score float64
		if quoteAID == winner {
			score = 1
		} else {
			score = 0
		}

		outcomeA, outcomeB := elo.Outcome(quoteA.Elo, quoteB.Elo, score)
		quoteA.Elo = outcomeA.Rating
		quoteB.Elo = outcomeB.Rating

		err = tx.Save(&quoteA).Error
		if err != nil {
			return fmt.Errorf("error saving quote A: %w", err)
		}

		err = tx.Save(&quoteB).Error
		if err != nil {
			return fmt.Errorf("error saving quote B: %w", err)
		}

		return nil
	})
	if err != nil {
		http.Error(w, "Error ranking quotes", http.StatusInternalServerError)
		return
	}

	quoteA, quoteB, err := pickRandomQuotePair(r.Context(), userId)
	if err != nil {
		http.Error(w, "Error getting quotes", http.StatusInternalServerError)
		return
	}

	components.RankForm(getRankFormProps(quoteA, quoteB)).Render(r.Context(), w)
}
