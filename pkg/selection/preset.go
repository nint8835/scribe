package selection

type SelectionPreset struct {
	ID               string
	DisplayName      string
	Description      string
	FirstMethod      FirstQuoteMethod
	SecondMethod     SecondQuoteMethod
	TiebreakerMethod TiebreakerMethod
}

var SelectionPresets = []SelectionPreset{
	{
		ID:               "default",
		DisplayName:      "Default",
		Description:      "Use the standard quote selection settings.",
		FirstMethod:      DefaultFirstQuoteMethod,
		SecondMethod:     DefaultSecondQuoteMethod,
		TiebreakerMethod: DefaultTiebreakerMethod,
	},
	{
		ID:               "tiebreaker",
		DisplayName:      "Tiebreaker",
		Description:      "Focus on closely ranked quotes to separate leaderboard ties.",
		FirstMethod:      FirstQuoteMethodClosestElo,
		SecondMethod:     SecondQuoteMethodClosestRank,
		TiebreakerMethod: TiebreakerMethodHighestRanked,
	},
}

func FindSelectionPreset(id string) (SelectionPreset, bool) {
	for _, preset := range SelectionPresets {
		if preset.ID == id {
			return preset, true
		}
	}

	return SelectionPreset{}, false
}

func (p SelectionPreset) Matches(first FirstQuoteMethod, second SecondQuoteMethod, tiebreaker TiebreakerMethod) bool {
	return p.FirstMethod == first &&
		p.SecondMethod == second &&
		p.TiebreakerMethod == tiebreaker
}
