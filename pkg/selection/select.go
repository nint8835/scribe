package selection

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"

	"gorm.io/gorm"

	"github.com/nint8835/scribe/pkg/database"
)

func attemptSelectQuotes(
	ctx context.Context,
	userId string,
	firstMethod FirstQuoteMethod,
	secondMethod SecondQuoteMethod,
	tiebreaker TiebreakerMethod,
) (database.Quote, database.Quote, error) {
	firstQuote, err := selectFirstQuote(ctx, userId, firstMethod, tiebreaker)
	if err != nil {
		return database.Quote{}, database.Quote{}, fmt.Errorf("error selecting first quote: %w", err)
	}

	secondQuote, err := selectSecondQuote(ctx, userId, firstQuote, secondMethod, tiebreaker)
	if err != nil {
		return database.Quote{}, database.Quote{}, fmt.Errorf("error selecting second quote: %w", err)
	}

	if rand.Intn(2) == 0 {
		return secondQuote, firstQuote, nil
	}

	return firstQuote, secondQuote, nil
}

func SelectQuotes(ctx context.Context, userId string, firstMethod FirstQuoteMethod, secondMethod SecondQuoteMethod, tiebreaker TiebreakerMethod) (database.Quote, database.Quote, error) {
	slog.Debug("Selecting quotes",
		"user_id", userId,
		"first_method", firstMethod,
		"second_method", secondMethod,
		"tiebreaker", tiebreaker,
	)

	attempts := 0

	var quoteA database.Quote
	var quoteB database.Quote
	var err error

	for {
		if attempts >= 10 {
			return quoteA, quoteB, ErrTooManyAttempts
		}

		quoteA, quoteB, err = attemptSelectQuotes(ctx, userId, firstMethod, secondMethod, tiebreaker)
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
