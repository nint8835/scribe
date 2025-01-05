package web

import (
	"errors"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/web/static"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
)

// TODO: Better error handling
type Server struct {
	serveMux     *http.ServeMux
	sessionStore *sessions.CookieStore
	oauthConfig  *oauth2.Config
}

func (s *Server) Run() error {
	if err := http.ListenAndServe(":8000", s.serveMux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	pages.Index().Render(r.Context(), w)
}

func New() (*Server, error) {
	serverInst := &Server{
		serveMux:     http.NewServeMux(),
		sessionStore: sessions.NewCookieStore([]byte(config.Instance.CookieSecret)),
		oauthConfig: &oauth2.Config{
			ClientID:     config.Instance.ClientId,
			ClientSecret: config.Instance.ClientSecret,
			RedirectURL:  config.Instance.CallbackUrl,
			Scopes:       []string{"identify", "guilds"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		},
	}

	serverInst.serveMux.HandleFunc("GET /{$}", serverInst.requireAuth(serverInst.handleIndex))
	serverInst.serveMux.HandleFunc("GET /auth/login", serverInst.handleAuthLogin)
	serverInst.serveMux.HandleFunc("GET /auth/callback", serverInst.handleAuthCallback)

	serverInst.serveMux.Handle("GET /static/", http.StripPrefix("/static/", hashfs.FileServer(static.HashFS)))

	return serverInst, nil
}
