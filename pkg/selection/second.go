package selection

import (
	"context"
	"encoding/gob"

	"github.com/nint8835/scribe/pkg/database"
)

func secondQuoteClosestRank(ctx context.Context, userId string, firstQuote database.Quote) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(`WITH
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
		firstQuote.Meta.ID,
		firstQuote.Meta.ID,
		userId,
		firstQuote.Meta.ID,
		firstQuote.Meta.ID,
		firstQuote.Elo,
		firstQuote.Meta.ID,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

const (
	SecondQuoteMethodClosestRank SecondQuoteMethod = "closest_rank"
)

var secondQuoteSelectors = map[SecondQuoteMethod]SecondQuoteSelector{
	SecondQuoteMethodClosestRank: secondQuoteClosestRank,
}

func selectSecondQuote(ctx context.Context, userId string, firstQuote database.Quote, method SecondQuoteMethod) (database.Quote, error) {
	selector, ok := secondQuoteSelectors[method]
	if !ok {
		return database.Quote{}, ErrUnknownMethod
	}
	return selector(ctx, userId, firstQuote)
}

func init() {
	gob.Register(SecondQuoteMethodClosestRank)
}
