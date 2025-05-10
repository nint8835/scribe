package web

import (
	"bytes"
	"context"
	"errors"
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

// TODO: See if there is a better way to select a pair of quotes
func attemptPickRandomQuotePair(ctx context.Context, userId string) (database.Quote, database.Quote, error) {
	var quoteA database.Quote
	var quoteB database.Quote

	err := database.Instance.WithContext(ctx).Model(&database.Quote{}).Order("RANDOM()").Take(&quoteA).Error
	if err != nil {
		return quoteA, quoteB, fmt.Errorf("error getting first random quote: %w", err)
	}

	// Needs the following indices, some of which currently aren't in code:
	// idx_comparisons_user_b_a	FALSE	user_id,quote_b_id,quote_a_id
	// idx_comparisons_user_a_b	FALSE	user_id,quote_a_id,quote_b_id

	// Find the quote with the closest ELO rating to the first quote that hasn't been compared by this user
	err = database.Instance.WithContext(ctx).Raw(
		`WITH
			compared_quotes AS (
				SELECT
					CASE
						WHEN c.quote_a_id = ? THEN c.quote_b_id
						WHEN c.quote_b_id = ? THEN c.quote_a_id
					END AS compared_quote_id
			FROM
				completed_comparisons c
			WHERE
				c.user_id = ?
				AND (
					c.quote_a_id = ?
					OR c.quote_b_id = ?
				)
			),
			filtered_quotes AS (
				SELECT
					q.*,
					ABS(q.elo - ?) AS elo_diff
				FROM
					quotes q
					LEFT JOIN compared_quotes cq ON q.id = cq.compared_quote_id
				WHERE
					q.deleted_at IS NULL
					AND q.id != ?
					AND cq.compared_quote_id IS NULL
			)
		SELECT
			q.*
		FROM
			filtered_quotes q
		ORDER BY
			elo_diff ASC
		LIMIT 1`,
		quoteA.Meta.ID,
		quoteA.Meta.ID,
		userId,
		quoteA.Meta.ID,
		quoteA.Meta.ID,
		quoteA.Elo,
		quoteA.Meta.ID,
	).Take(&quoteB).Error
	if err != nil {
		return quoteA, quoteB, fmt.Errorf("error getting second random quote: %w", err)
	}

	return quoteA, quoteB, nil
}

func fetchRandomQuotePair(ctx context.Context, userId string) (database.Quote, database.Quote, error) {
	attempts := 0

	var quoteA database.Quote
	var quoteB database.Quote
	var err error

	for {
		if attempts >= 10 {
			return quoteA, quoteB, errors.New("too many attempts to get random quotes")
		}

		quoteA, quoteB, err = attemptPickRandomQuotePair(ctx, userId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				attempts++
				continue
			}

			return quoteA, quoteB, err
		}

		break
	}

	return quoteA, quoteB, nil
}

func (s *Server) getRankStatsDisplayProps(id string) (components.RankStatsDisplayProps, error) {
	var stats components.RankStatsDisplayProps

	err := database.Instance.Model(&database.CompletedComparison{}).Where("user_id = ?", id).Count(&stats.UserRankCount).Error
	if err != nil {
		return stats, fmt.Errorf("error getting user rank count: %w", err)
	}

	err = database.Instance.Model(&database.CompletedComparison{}).Count(&stats.TotalRankCount).Error
	if err != nil {
		return stats, fmt.Errorf("error getting total rank count: %w", err)
	}

	var quoteCount int64
	err = database.Instance.Model(&database.Quote{}).Count(&quoteCount).Error
	if err != nil {
		return stats, fmt.Errorf("error getting quote count: %w", err)
	}

	stats.MaxRankCount = quoteCount * (quoteCount - 1) / 2

	return stats, nil
}

func (s *Server) getRankFormProps(quoteA database.Quote, quoteB database.Quote) (components.RankProps, error) {
	quoteAContent, err := s.renderQuoteText(quoteA)
	if err != nil {
		return components.RankProps{}, fmt.Errorf("error rendering quote A: %w", err)
	}

	quoteBContent, err := s.renderQuoteText(quoteB)
	if err != nil {
		return components.RankProps{}, fmt.Errorf("error rendering quote B: %w", err)
	}

	return components.RankProps{
		QuoteAID:      fmt.Sprintf("%d", quoteA.Meta.ID),
		QuoteBID:      fmt.Sprintf("%d", quoteB.Meta.ID),
		QuoteAContent: quoteAContent,
		QuoteBContent: quoteBContent,
	}, nil
}

func (s *Server) renderQuoteText(quote database.Quote) (string, error) {
	var buf bytes.Buffer
	err := s.md.Convert([]byte(quote.Text), &buf)
	return buf.String(), err
}

func (s *Server) handleGetRank(w http.ResponseWriter, r *http.Request) error {
	userId := s.getCurrentUserId(r)

	quoteA, quoteB, err := fetchRandomQuotePair(r.Context(), userId)
	if err != nil {
		return fmt.Errorf("error getting quotes: %w", err)
	}

	props, err := s.getRankFormProps(quoteA, quoteB)
	if err != nil {
		return fmt.Errorf("error getting rank form props: %w", err)
	}

	stats, err := s.getRankStatsDisplayProps(userId)
	if err != nil {
		return fmt.Errorf("error getting rank stats props: %w", err)
	}

	pages.Rank(props, stats).Render(r.Context(), w)

	return nil
}

func (s *Server) handlePostRank(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()

	userId := s.getCurrentUserId(r)

	quoteAID := r.PostForm.Get("quote_a_id")
	quoteBID := r.PostForm.Get("quote_b_id")
	winner := r.PostForm.Get("winner")

	if quoteAID == "" || quoteBID == "" || winner == "" {
		return httpError{
			StatusCode: http.StatusBadRequest,
			Message:    "Missing required fields.",
		}
	}

	if quoteAID == quoteBID {
		return httpError{
			StatusCode: http.StatusBadRequest,
			Message:    "Cannot rank the same quote against itself.",
		}
	}

	quoteAIDInt, err := strconv.ParseUint(quoteAID, 10, 64)
	if err != nil {
		return httpError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid quote A ID.",
		}
	}

	quoteBIDInt, err := strconv.ParseUint(quoteBID, 10, 64)
	if err != nil {
		return httpError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid quote B ID.",
		}
	}

	winnerInt, err := strconv.ParseUint(winner, 10, 64)
	if err != nil {
		return httpError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid winner.",
		}
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

		var comparisonExists bool
		matchingComparisonSubquery := tx.Model(&database.CompletedComparison{}).
			Where(
				"(quote_a_id = ? AND quote_b_id = ? AND user_id = ?) OR (quote_a_id = ? AND quote_b_id = ? AND user_id = ?)",
				quoteAIDInt,
				quoteBIDInt,
				userId,
				quoteBIDInt,
				quoteAIDInt,
				userId,
			).Find(&database.CompletedComparison{})
		err = tx.Raw("SELECT EXISTS (?)", matchingComparisonSubquery).Scan(&comparisonExists).Error
		if err != nil {
			return fmt.Errorf("error checking if comparison exists: %w", err)
		}

		if comparisonExists {
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
		return fmt.Errorf("error ranking quotes: %w", err)
	}

	quoteA, quoteB, err := fetchRandomQuotePair(r.Context(), userId)
	if err != nil {
		return fmt.Errorf("error getting next quotes: %w", err)
	}

	props, err := s.getRankFormProps(quoteA, quoteB)
	if err != nil {
		return fmt.Errorf("error getting rank form props: %w", err)
	}

	stats, err := s.getRankStatsDisplayProps(userId)
	if err != nil {
		return fmt.Errorf("error getting rank stats props: %w", err)
	}
	stats.ShouldSwap = true

	components.RankStatsDisplay(stats).Render(r.Context(), w)
	components.RankForm(props).Render(r.Context(), w)

	return nil
}

func (s *Server) handleRankStats(w http.ResponseWriter, r *http.Request) error {
	userId := s.getCurrentUserId(r)

	stats, err := s.getRankStatsDisplayProps(userId)
	if err != nil {
		return fmt.Errorf("error getting rank stats props: %w", err)
	}

	components.RankStatsDisplay(stats).Render(r.Context(), w)

	return nil
}
