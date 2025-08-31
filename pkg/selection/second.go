package selection

import (
	"context"
	"encoding/gob"
	"fmt"

	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/embedding"
)

func secondQuoteClosestRank(ctx context.Context, userId string, firstQuote database.Quote) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
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

func secondQuoteFurthestRank(ctx context.Context, userId string, firstQuote database.Quote) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	err := db.Raw(
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
			elo_diff DESC
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

func secondQuoteSemanticSimilarity(ctx context.Context, userId string, firstQuote database.Quote) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)

	encodedEmbedding, err := embedding.EmbedQuote(firstQuote.Text)
	if err != nil {
		return database.Quote{}, fmt.Errorf("failed to serialize embedding: %w", err)
	}

	err = db.Raw(
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
					q.*
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
			quote_embeddings qe
			JOIN filtered_quotes q ON q.id = qe.rowid
		WHERE
			qe.rowid IN (SELECT id FROM filtered_quotes)
			AND qe.embedding MATCH ?
			AND qe.k = 1
		ORDER BY
			distance`,
		firstQuote.Meta.ID,
		firstQuote.Meta.ID,
		userId,
		firstQuote.Meta.ID,
		firstQuote.Meta.ID,
		firstQuote.Meta.ID,
		encodedEmbedding,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

func secondQuoteRandom(ctx context.Context, userId string, firstQuote database.Quote) (database.Quote, error) {
	var quote database.Quote
	db := database.Instance.WithContext(ctx)
	err := db.Raw(
		`SELECT
			*		
		FROM
			quotes
		WHERE
			deleted_at IS NULL
			AND id != ?
			AND NOT EXISTS (
				SELECT
					1
				FROM
					completed_comparisons
				WHERE
					user_id = ?
					AND quote_a_id = ?
					AND quote_b_id = quotes.id
				UNION ALL
				SELECT
					1
				FROM
					completed_comparisons
				WHERE
					user_id = ?
					AND quote_b_id = ?
					AND quote_a_id = quotes.id
			)
		ORDER BY
			random()
		LIMIT 1`,
		firstQuote.Meta.ID,
		userId,
		firstQuote.Meta.ID,
		userId,
		firstQuote.Meta.ID,
	).Take(&quote).Error

	if err != nil {
		return database.Quote{}, err
	}

	return quote, nil
}

const (
	SecondQuoteMethodClosestRank        SecondQuoteMethod = "closest_rank"
	SecondQuoteMethodFurthestRank       SecondQuoteMethod = "furthest_rank"
	SecondQuoteMethodSemanticSimilarity SecondQuoteMethod = "semantic_similarity"
	SecondQuoteMethodRandom             SecondQuoteMethod = "random"
)

func (m SecondQuoteMethod) String() string {
	return string(m)
}

func (m SecondQuoteMethod) DisplayName() string {
	switch m {
	case SecondQuoteMethodClosestRank:
		return "Closest rank"
	case SecondQuoteMethodFurthestRank:
		return "Furthest rank"
	case SecondQuoteMethodSemanticSimilarity:
		return "Semantic similarity"
	case SecondQuoteMethodRandom:
		return "Random"
	default:
		return "Unknown method"
	}
}

func (m SecondQuoteMethod) Description() string {
	switch m {
	case SecondQuoteMethodClosestRank:
		return "Selects a quote that is closest in Elo rating to the first quote, that you have not already ranked."
	case SecondQuoteMethodFurthestRank:
		return "Selects a quote that is furthest in Elo rating from the first quote, that you have not already ranked."
	case SecondQuoteMethodSemanticSimilarity:
		return "Selects a quote that is semantically similar to the first quote, that you have not already ranked."
	case SecondQuoteMethodRandom:
		return "Selects a random quote that you have not already ranked against the first quote."
	default:
		return "Unknown method"
	}
}

var SecondQuoteMethods = []SecondQuoteMethod{
	SecondQuoteMethodClosestRank,
	SecondQuoteMethodFurthestRank,
	SecondQuoteMethodSemanticSimilarity,
	SecondQuoteMethodRandom,
}

var secondQuoteSelectors = map[SecondQuoteMethod]SecondQuoteSelector{
	SecondQuoteMethodClosestRank:        secondQuoteClosestRank,
	SecondQuoteMethodFurthestRank:       secondQuoteFurthestRank,
	SecondQuoteMethodSemanticSimilarity: secondQuoteSemanticSimilarity,
	SecondQuoteMethodRandom:             secondQuoteRandom,
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
