package web

import (
	"fmt"
	"net/http"

	"github.com/nint8835/scribe/pkg/web/ui/components"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

func (s *Server) handleGetRank(w http.ResponseWriter, r *http.Request) {
	_ = s.getCurrentUserId(r)

	fakeQuotes := components.RankProps{
		QuoteAID:      "1",
		QuoteBID:      "2",
		QuoteAContent: "Example quote A",
		QuoteBContent: "Example quote B",
	}

	pages.Rank(fakeQuotes).Render(r.Context(), w)
}

func (s *Server) handlePostRank(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	_ = s.getCurrentUserId(r)
	fmt.Printf("%#+v\n", r.PostForm)

	fakeQuotes := components.RankProps{
		QuoteAID:      "1",
		QuoteBID:      "2",
		QuoteAContent: "Example quote A",
		QuoteBContent: "Example quote B",
	}

	components.RankForm(fakeQuotes).Render(r.Context(), w)
}
