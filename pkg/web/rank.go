package web

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	elogo "github.com/kortemy/elo-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/selection"
	"github.com/nint8835/scribe/pkg/web/ui/components"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

var elo *elogo.Elo = elogo.NewElo()

func (s *Server) getQuoteRank(quoteID uint) (int, error) {
	var rank int64
	err := database.Instance.Model(&database.Quote{}).
		Where("elo > (SELECT elo FROM quotes WHERE id = ?)", quoteID).
		Count(&rank).Error
	if err != nil {
		return 0, fmt.Errorf("error calculating quote rank: %w", err)
	}
	return int(rank) + 1, nil
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

func (s *Server) getRankResultProps(quoteA database.Quote, quoteB database.Quote, quoteARankBefore, quoteARankAfter, quoteBRankBefore, quoteBRankAfter int) (components.RankResultProps, error) {
	quoteAContent, err := s.renderQuoteText(quoteA)
	if err != nil {
		return components.RankResultProps{}, fmt.Errorf("error rendering quote A: %w", err)
	}

	quoteBContent, err := s.renderQuoteText(quoteB)
	if err != nil {
		return components.RankResultProps{}, fmt.Errorf("error rendering quote B: %w", err)
	}

	quoteAAuthors, err := s.formatAuthors(quoteA)
	if err != nil {
		return components.RankResultProps{}, fmt.Errorf("error formatting author names for quote A: %w", err)
	}

	quoteBAuthors, err := s.formatAuthors(quoteB)
	if err != nil {
		return components.RankResultProps{}, fmt.Errorf("error formatting author names for quote B: %w", err)
	}

	return components.RankResultProps{
		QuoteAID:         quoteA.Meta.ID,
		QuoteBID:         quoteB.Meta.ID,
		QuoteAContent:    quoteAContent,
		QuoteBContent:    quoteBContent,
		QuoteARankBefore: quoteARankBefore,
		QuoteBRankBefore: quoteBRankBefore,
		QuoteARankAfter:  quoteARankAfter,
		QuoteBRankAfter:  quoteBRankAfter,
		QuoteAAuthors:    quoteAAuthors,
		QuoteBAuthors:    quoteBAuthors,
	}, nil
}

func (s *Server) renderQuoteText(quote database.Quote) (string, error) {
	var buf bytes.Buffer
	err := s.md.Convert([]byte(quote.Text), &buf)
	return buf.String(), err
}

func (s *Server) selectQuotes(r *http.Request) (database.Quote, database.Quote, error) {
	session := s.getSession(r)
	userId := s.getCurrentUserId(r)

	firstMethod, firstMethodSet := session.Values["first_method"].(selection.FirstQuoteMethod)
	secondMethod, secondMethodSet := session.Values["second_method"].(selection.SecondQuoteMethod)

	if !firstMethodSet {
		firstMethod = selection.FirstQuoteMethodLeastSeen
	}

	if !secondMethodSet {
		secondMethod = selection.SecondQuoteMethodClosestRank
	}

	return selection.SelectQuotes(r.Context(), userId, firstMethod, secondMethod)
}

func (s *Server) handleGetRank(w http.ResponseWriter, r *http.Request) error {
	userId := s.getCurrentUserId(r)

	quoteA, quoteB, err := s.selectQuotes(r)
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

	pages.Rank(props, stats, nil).Render(r.Context(), w)

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

	quoteARankBefore, err := s.getQuoteRank(uint(quoteAIDInt))
	if err != nil {
		return fmt.Errorf("error getting quote A rank before: %w", err)
	}

	quoteBRankBefore, err := s.getQuoteRank(uint(quoteBIDInt))
	if err != nil {
		return fmt.Errorf("error getting quote B rank before: %w", err)
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

	var previousQuoteA, previousQuoteB database.Quote
	err = database.Instance.Model(&database.Quote{}).Where("id = ?", quoteAIDInt).Preload("Authors").First(&previousQuoteA).Error
	if err != nil {
		return fmt.Errorf("error getting previous quote A: %w", err)
	}

	err = database.Instance.Model(&database.Quote{}).Where("id = ?", quoteBIDInt).Preload("Authors").First(&previousQuoteB).Error
	if err != nil {
		return fmt.Errorf("error getting previous quote B: %w", err)
	}

	quoteARankAfter, err := s.getQuoteRank(uint(quoteAIDInt))
	if err != nil {
		return fmt.Errorf("error getting quote A rank after: %w", err)
	}

	quoteBRankAfter, err := s.getQuoteRank(uint(quoteBIDInt))
	if err != nil {
		return fmt.Errorf("error getting quote B rank after: %w", err)
	}

	nextQuoteA, nextQuoteB, err := s.selectQuotes(r)
	if err != nil {
		return fmt.Errorf("error getting next quotes: %w", err)
	}

	props, err := s.getRankFormProps(nextQuoteA, nextQuoteB)
	if err != nil {
		return fmt.Errorf("error getting rank form props: %w", err)
	}

	stats, err := s.getRankStatsDisplayProps(userId)
	if err != nil {
		return fmt.Errorf("error getting rank stats props: %w", err)
	}
	stats.ShouldSwap = true

	rankResultProps, err := s.getRankResultProps(previousQuoteA, previousQuoteB, quoteARankBefore, quoteARankAfter, quoteBRankBefore, quoteBRankAfter)
	if err != nil {
		return fmt.Errorf("error getting rank result props: %w", err)
	}

	components.RankStatsDisplay(stats).Render(r.Context(), w)
	components.RankForm(props).Render(r.Context(), w)
	components.RankResult(rankResultProps).Render(r.Context(), w)

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
