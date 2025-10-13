package selection

import (
	"context"
	"encoding/gob"

	"github.com/nint8835/scribe/pkg/database"
)

func firstQuoteLeastSeen(ctx context.Context, userId string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
		`SELECT
			q.*,
			COUNT(DISTINCT ca.id) + COUNT(DISTINCT cb.id) AS comparison_count
		FROM
			quotes q
			LEFT JOIN completed_comparisons ca ON (
				ca.user_id = ?
				AND ca.quote_a_id = q.id
			)
			LEFT JOIN completed_comparisons cb ON (
				cb.user_id = ?
				AND cb.quote_b_id = q.id
			)
		WHERE
			q.deleted_at IS NULL
		GROUP BY
			q.id
		ORDER BY
			comparison_count ASC,
			random()
		LIMIT
		1`,
		userId,
		userId,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func firstQuoteLeastSeenGlobal(ctx context.Context, userId string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
		`SELECT
			q.*,
			COUNT(DISTINCT ca.id) + COUNT(DISTINCT cb.id) AS comparison_count
		FROM
			quotes q
			LEFT JOIN completed_comparisons ca ON ca.quote_a_id = q.id
			LEFT JOIN completed_comparisons cb ON cb.quote_b_id = q.id
		WHERE
			q.deleted_at IS NULL
		GROUP BY
			q.id
		ORDER BY
			comparison_count ASC,
			random()
		LIMIT
		1`,
		userId,
		userId,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func firstQuoteMostAverage(ctx context.Context, userId string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
		`SELECT
			q.*,
			ABS(q.elo - 1000) AS elo_diff
		FROM
			quotes q
		WHERE
			q.deleted_at IS NULL
		ORDER BY
			elo_diff ASC,
			random()
		LIMIT
			1`,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func firstQuoteClosestElo(ctx context.Context, userId string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
		`WITH
			elo_diffs AS (
				SELECT
					*,
					LEAD(elo, 1) OVER (
						ORDER BY
							elo
					) - elo AS next_diff,
					elo - LAG(elo, 1) OVER (
						ORDER BY
							elo
					) AS prev_diff
				FROM
					quotes
				WHERE
					deleted_at IS NULL
			)
		SELECT
			*
		FROM
			elo_diffs
		ORDER BY
			MIN(COALESCE(next_diff, 9999999), COALESCE(prev_diff, 9999999)) ASC,
			RANDOM()
		LIMIT
			1`,
	).Take(&quote).Error
	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func firstQuoteClosestTiebreakHighest(ctx context.Context, userId string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
		`WITH
			elo_diffs AS (
				SELECT
					*,
					LEAD(elo, 1) OVER (
						ORDER BY
							elo
					) - elo AS next_diff,
					elo - LAG(elo, 1) OVER (
						ORDER BY
							elo
					) AS prev_diff
				FROM
					quotes
				WHERE
					deleted_at IS NULL
			)
		SELECT
			*
		FROM
			elo_diffs
		ORDER BY
			MIN(COALESCE(next_diff, 9999999), COALESCE(prev_diff, 9999999)) ASC,
			elo DESC
		LIMIT
			1`,
	).Take(&quote).Error
	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func firstQuoteLastWinner(ctx context.Context, userId string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
		`SELECT
			*
		FROM
			completed_comparisons
			JOIN quotes ON quotes.id = completed_comparisons.winner_id
		WHERE
			completed_comparisons.user_id = ?
			AND completed_comparisons.deleted_at IS NULL
			AND quotes.deleted_at IS NULL
		ORDER BY
			completed_comparisons.id DESC
		LIMIT
			1`,
		userId,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func firstQuoteRandom(ctx context.Context, _ string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
		`SELECT
			*
		FROM
			quotes
		WHERE
			deleted_at IS NULL
		ORDER BY
			random()
		LIMIT
			1`,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

const (
	FirstQuoteMethodLeastSeen              FirstQuoteMethod = "least_seen"
	FirstQuoteMethodLeastSeenGlobal        FirstQuoteMethod = "least_seen_global"
	FirstQuoteMethodMostAverage            FirstQuoteMethod = "most_average"
	FirstQuoteMethodClosestElo             FirstQuoteMethod = "closest_elo"
	FirstQuoteMethodClosestTiebreakHighest FirstQuoteMethod = "closest_elo_tiebreak_highest"
	FirstQuoteMethodLastWinner             FirstQuoteMethod = "last_winner"
	FirstQuoteMethodRandom                 FirstQuoteMethod = "random"
)

func (m FirstQuoteMethod) String() string {
	return string(m)
}

func (m FirstQuoteMethod) DisplayName() string {
	switch m {
	case FirstQuoteMethodLeastSeen:
		return "Least seen"
	case FirstQuoteMethodLeastSeenGlobal:
		return "Least seen (global)"
	case FirstQuoteMethodMostAverage:
		return "Most average"
	case FirstQuoteMethodClosestElo:
		return "Closest Elo"
	case FirstQuoteMethodClosestTiebreakHighest:
		return "Closest Elo (tiebreak highest)"
	case FirstQuoteMethodLastWinner:
		return "Last winner"
	case FirstQuoteMethodRandom:
		return "Random"
	default:
		return "Unknown method"
	}
}

func (m FirstQuoteMethod) Description() string {
	switch m {
	case FirstQuoteMethodLeastSeen:
		return "Selects the quote that you have completed the least comparisons for."
	case FirstQuoteMethodLeastSeenGlobal:
		return "Selects the quote that has had the least comparisons completed for, across all users."
	case FirstQuoteMethodMostAverage:
		return "Selects the quote with an Elo rating closest to the average / starting Elo rating."
	case FirstQuoteMethodClosestElo:
		return "Selects the quote with the Elo closest to its neighbours in the leaderboard."
	case FirstQuoteMethodClosestTiebreakHighest:
		return "Selects the quote with the Elo closest to its neighbours in the leaderboard. In the case of a tie, selects the quote with the highest Elo."
	case FirstQuoteMethodLastWinner:
		return "Selects the quote that you last voted as the winner in a comparison."
	case FirstQuoteMethodRandom:
		return "Selects a quote completely at random."
	default:
		return "Unknown method"
	}
}

var FirstQuoteMethods = []FirstQuoteMethod{
	FirstQuoteMethodLeastSeen,
	FirstQuoteMethodLeastSeenGlobal,
	FirstQuoteMethodMostAverage,
	FirstQuoteMethodClosestElo,
	FirstQuoteMethodClosestTiebreakHighest,
	FirstQuoteMethodLastWinner,
	FirstQuoteMethodRandom,
}

var firstQuoteSelectors = map[FirstQuoteMethod]FirstQuoteSelector{
	FirstQuoteMethodLeastSeen:              firstQuoteLeastSeen,
	FirstQuoteMethodLeastSeenGlobal:        firstQuoteLeastSeenGlobal,
	FirstQuoteMethodMostAverage:            firstQuoteMostAverage,
	FirstQuoteMethodClosestElo:             firstQuoteClosestElo,
	FirstQuoteMethodClosestTiebreakHighest: firstQuoteClosestTiebreakHighest,
	FirstQuoteMethodLastWinner:             firstQuoteLastWinner,
	FirstQuoteMethodRandom:                 firstQuoteRandom,
}

var DefaultFirstQuoteMethod = FirstQuoteMethodLeastSeen

func selectFirstQuote(ctx context.Context, userId string, method FirstQuoteMethod) (database.Quote, error) {
	selector, ok := firstQuoteSelectors[method]
	if !ok {
		selector = firstQuoteSelectors[DefaultFirstQuoteMethod]
	}
	return selector(ctx, userId)
}

func init() {
	gob.Register(FirstQuoteMethodLeastSeen)
}
