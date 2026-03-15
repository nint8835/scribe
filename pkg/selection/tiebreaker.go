package selection

import "encoding/gob"

const (
	TiebreakerMethodRandom        TiebreakerMethod = "random"
	TiebreakerMethodHighestRanked TiebreakerMethod = "highest_ranked"
)

var DefaultTiebreakerMethod = TiebreakerMethodHighestRanked

var TiebreakerMethods = []TiebreakerMethod{
	TiebreakerMethodRandom,
	TiebreakerMethodHighestRanked,
}

var tiebreakerSQL = map[TiebreakerMethod]string{
	TiebreakerMethodRandom:        "RANDOM()",
	TiebreakerMethodHighestRanked: "elo DESC",
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
	default:
		return "Unknown tiebreaker"
	}
}

func (m TiebreakerMethod) Description() string {
	switch m {
	case TiebreakerMethodRandom:
		return "Breaks ties randomly."
	case TiebreakerMethodHighestRanked:
		return "When there is a tie, selects the quote with the highest Elo rating."
	default:
		return "Unknown tiebreaker"
	}
}

func init() {
	gob.Register(TiebreakerMethodHighestRanked)
}
