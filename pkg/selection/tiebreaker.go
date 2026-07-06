package selection

import "encoding/gob"

const (
	TiebreakerMethodRandom        TiebreakerMethod = "random"
	TiebreakerMethodHighestRanked TiebreakerMethod = "highest_ranked"
	TiebreakerMethodLowestRanked  TiebreakerMethod = "lowest_ranked"
	TiebreakerMethodNewest        TiebreakerMethod = "newest"
	TiebreakerMethodOldest        TiebreakerMethod = "oldest"
)

var DefaultTiebreakerMethod = TiebreakerMethodHighestRanked

var TiebreakerMethods = []TiebreakerMethod{
	TiebreakerMethodRandom,
	TiebreakerMethodHighestRanked,
	TiebreakerMethodLowestRanked,
	TiebreakerMethodNewest,
	TiebreakerMethodOldest,
}

var tiebreakerSQL = map[TiebreakerMethod]string{
	TiebreakerMethodRandom:        "RANDOM()",
	TiebreakerMethodHighestRanked: "elo DESC",
	TiebreakerMethodLowestRanked:  "elo ASC",
	TiebreakerMethodNewest:        "created_at DESC",
	TiebreakerMethodOldest:        "created_at ASC",
}

func (m TiebreakerMethod) String() string {
	return string(m)
}

func (m TiebreakerMethod) DisplayName() string {
	switch m {
	case TiebreakerMethodRandom:
		return "Random"
	case TiebreakerMethodHighestRanked:
		return "Highest ranked"
	case TiebreakerMethodLowestRanked:
		return "Lowest ranked"
	case TiebreakerMethodNewest:
		return "Newest"
	case TiebreakerMethodOldest:
		return "Oldest"
	default:
		return "Unknown tiebreaker"
	}
}

func (m TiebreakerMethod) Description() string {
	switch m {
	case TiebreakerMethodRandom:
		return "Breaks ties randomly."
	case TiebreakerMethodHighestRanked:
		return "Selects the tied quote with the highest Elo rating."
	case TiebreakerMethodLowestRanked:
		return "Selects the tied quote with the lowest Elo rating."
	case TiebreakerMethodNewest:
		return "Selects the newest tied quote."
	case TiebreakerMethodOldest:
		return "Selects the oldest tied quote."
	default:
		return "Unknown tiebreaker"
	}
}

func init() {
	gob.Register(TiebreakerMethodRandom)
	gob.Register(TiebreakerMethodHighestRanked)
	gob.Register(TiebreakerMethodLowestRanked)
	gob.Register(TiebreakerMethodNewest)
	gob.Register(TiebreakerMethodOldest)
}
