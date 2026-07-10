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
	tiebreakerMethod, tiebreakerMethodSet := session.Values["tiebreaker_method"].(selection.TiebreakerMethod)

	if !firstMethodSet {
		firstMethod = selection.DefaultFirstQuoteMethod
	}

	if !secondMethodSet {
		secondMethod = selection.DefaultSecondQuoteMethod
	}

	if !tiebreakerMethodSet {
		tiebreakerMethod = selection.DefaultTiebreakerMethod
	}

	err := pages.Options(
		pages.OptionsProps{
			FirstMethod:      firstMethod,
			SecondMethod:     secondMethod,
			TiebreakerMethod: tiebreakerMethod,
		},
	).Render(r.Context(), w)
	if err != nil {
		return fmt.Errorf("error rendering options: %w", err)
	}

	return nil
}

func (s *Server) handlePostOptions(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("error parsing form: %w", err)
	}

	session := s.getSession(r)

	var firstMethod selection.FirstQuoteMethod
	var secondMethod selection.SecondQuoteMethod
	var tiebreakerMethod selection.TiebreakerMethod

	preset, presetFound := selection.FindSelectionPreset(r.PostForm.Get("preset"))
	if presetFound {
		firstMethod = preset.FirstMethod
		secondMethod = preset.SecondMethod
		tiebreakerMethod = preset.TiebreakerMethod
	} else {
		firstMethod = selection.FirstQuoteMethod(r.PostForm.Get("first_method"))
		secondMethod = selection.SecondQuoteMethod(r.PostForm.Get("second_method"))
		tiebreakerMethod = selection.TiebreakerMethod(r.PostForm.Get("tiebreaker_method"))
	}

	session.Values["first_method"] = firstMethod
	session.Values["second_method"] = secondMethod
	session.Values["tiebreaker_method"] = tiebreakerMethod

	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("error saving session: %w", err)
	}

	http.Redirect(w, r, "/rank", http.StatusSeeOther)
	return nil
}
