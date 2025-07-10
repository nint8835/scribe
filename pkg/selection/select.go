package selection

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/nint8835/scribe/pkg/database"
)

func attemptSelectQuotes(
	ctx context.Context,
	userId string,
	firstMethod FirstQuoteMethod,
	secondMethod SecondQuoteMethod,
) (database.Quote, database.Quote, error) {
	firstQuote, err := selectFirstQuote(ctx, userId, firstMethod)
	if err != nil {
		return database.Quote{}, database.Quote{}, fmt.Errorf("error selecting first quote: %w", err)
	}

	secondQuote, err := selectSecondQuote(ctx, userId, firstQuote, secondMethod)
	if err != nil {
		return database.Quote{}, database.Quote{}, fmt.Errorf("error selecting second quote: %w", err)
	}

	return firstQuote, secondQuote, nil
}

func SelectQuotes(ctx context.Context, userId string, firstMethod FirstQuoteMethod, secondMethod SecondQuoteMethod) (database.Quote, database.Quote, error) {
	slog.Debug("Selecting quotes",
		"user_id", userId,
		"first_method", firstMethod,
		"second_method", secondMethod,
	)

	attempts := 0

	var quoteA database.Quote
	var quoteB database.Quote
	var err error

	for {
		if attempts >= 10 {
			return quoteA, quoteB, ErrTooManyAttempts
		}

		quoteA, quoteB, err = attemptSelectQuotes(ctx, userId, firstMethod, secondMethod)
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
