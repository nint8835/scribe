package selection

import (
	"context"

	"github.com/nint8835/scribe/pkg/database"
)

type FirstQuoteSelector func(ctx context.Context, userId string, tiebreaker TiebreakerMethod) (database.Quote, error)
type SecondQuoteSelector func(ctx context.Context, userId string, firstQuote database.Quote, tiebreaker TiebreakerMethod) (database.Quote, error)

type FirstQuoteMethod string
type SecondQuoteMethod string
type TiebreakerMethod string
