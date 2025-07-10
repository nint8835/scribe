package web

import (
	"fmt"
	"net/http"

	"github.com/nint8835/scribe/pkg/selection"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

func (s *Server) handleGetOptions(w http.ResponseWriter, r *http.Request) error {
	session := s.getSession(r)

	firstMethod, firstMethodSet := session.Values["first_method"].(selection.FirstQuoteMethod)
	secondMethod, secondMethodSet := session.Values["second_method"].(selection.SecondQuoteMethod)

	if !firstMethodSet {
		firstMethod = selection.FirstQuoteMethodLeastSeen
	}

	if !secondMethodSet {
		secondMethod = selection.SecondQuoteMethodClosestRank
	}

	pages.Options(
		pages.OptionsProps{
			FirstMethod:  firstMethod,
			SecondMethod: secondMethod,
		},
	).Render(r.Context(), w)

	return nil
}

func (s *Server) handlePostOptions(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()
	session := s.getSession(r)

	firstMethod := selection.FirstQuoteMethod(r.PostForm.Get("first_method"))
	secondMethod := selection.SecondQuoteMethod(r.PostForm.Get("second_method"))

	session.Values["first_method"] = firstMethod
	session.Values["second_method"] = secondMethod
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("error saving session: %w", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
