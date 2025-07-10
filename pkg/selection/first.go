package selection

import (
	"context"
	"encoding/gob"

	"github.com/nint8835/scribe/pkg/database"
)

func firstQuoteLeastSeen(ctx context.Context, userId string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(`SELECT
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
		1`, userId, userId).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func firstQuoteRandom(ctx context.Context, _ string) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(`SELECT
			*
		FROM
			quotes
		WHERE
			deleted_at IS NULL
		ORDER BY
			random()
		LIMIT
			1`).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

const (
	FirstQuoteMethodLeastSeen FirstQuoteMethod = "least_seen"
	FirstQuoteMethodRandom    FirstQuoteMethod = "random"
)

func (m FirstQuoteMethod) String() string {
	return string(m)
}

func (m FirstQuoteMethod) DisplayName() string {
	switch m {
	case FirstQuoteMethodLeastSeen:
		return "Least seen"
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
	case FirstQuoteMethodRandom:
		return "Selects a quote completely at random."
	default:
		return "Unknown method"
	}
}

var FirstQuoteMethods = []FirstQuoteMethod{
	FirstQuoteMethodLeastSeen,
	FirstQuoteMethodRandom,
}

var firstQuoteSelectors = map[FirstQuoteMethod]FirstQuoteSelector{
	FirstQuoteMethodLeastSeen: firstQuoteLeastSeen,
	FirstQuoteMethodRandom:    firstQuoteRandom,
}

func selectFirstQuote(ctx context.Context, userId string, method FirstQuoteMethod) (database.Quote, error) {
	selector, ok := firstQuoteSelectors[method]
	if !ok {
		return database.Quote{}, ErrUnknownMethod
	}
	return selector(ctx, userId)
}

func init() {
	gob.Register(FirstQuoteMethodLeastSeen)
}
