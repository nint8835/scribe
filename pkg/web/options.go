package web

import (
	"log/slog"
	"net/http"

	"github.com/nint8835/scribe/pkg/selection"
)

func (s *Server) handleTestSetSession(w http.ResponseWriter, r *http.Request) error {
	session := s.getSession(r)
	session.Values["first_method"] = selection.FirstQuoteMethodRandom

	if err := session.Save(r, w); err != nil {
		slog.Error("Failed to save session", "error", err)
		return &httpError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to save session",
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return nil
}
