package web

import (
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/sessions"

	"github.com/nint8835/scribe/pkg/config"
)

func (s *Server) getSession(r *http.Request) *sessions.Session {
	session, _ := s.sessionStore.Get(r, "session")
	return session
}

func (s *Server) getCurrentUserId(r *http.Request) string {
	session := s.getSession(r)

	if userId, hasUserId := session.Values["userId"]; hasUserId {
		return userId.(string)
	}

	return ""
}

func (s *Server) requireAuth(handler errorHandlerFunc) errorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		userId := s.getCurrentUserId(r)

		if userId == "" {
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return nil
		}

		return handler(w, r)
	}
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.oauthConfig.AuthCodeURL("state"), http.StatusTemporaryRedirect)
}

func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) error {
	session := s.getSession(r)

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if state != "state" {
		return httpError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid state.",
		}
	}

	token, err := s.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		return fmt.Errorf("error exchanging token: %w", err)
	}

	discordClient, _ := discordgo.New(fmt.Sprintf("Bearer %s", token.AccessToken))

	currentUser, err := discordClient.User("@me")
	if err != nil {
		return fmt.Errorf("error getting current user: %w", err)
	}

	guilds, err := discordClient.UserGuilds(200, "", "", false)
	if err != nil {
		return fmt.Errorf("error getting guilds: %w", err)
	}

	isMember := false

	for _, guild := range guilds {
		if guild.ID == config.Instance.GuildId {
			isMember = true
			break
		}
	}

	if !isMember {
		return httpError{
			StatusCode: http.StatusForbidden,
			Message:    "You are not a member of the required guild.",
		}
	}

	session.Values["userId"] = currentUser.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
}
