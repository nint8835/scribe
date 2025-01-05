package web

import (
	"errors"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/gorilla/sessions"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/oauth2"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/web/static"
)

// TODO: Better error handling
type Server struct {
	serveMux     *http.ServeMux
	sessionStore *sessions.CookieStore
	oauthConfig  *oauth2.Config
	md           goldmark.Markdown
}

func (s *Server) Run() error {
	if err := http.ListenAndServe(":8000", s.serveMux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/rank", http.StatusTemporaryRedirect)
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
		// TODO:
		//   - Handle Discord mentions
		//   - Handle Discord emotes?
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.Strikethrough,
			),
		),
	}

	serverInst.serveMux.HandleFunc("GET /{$}", serverInst.requireAuth(serverInst.handleIndex))
	serverInst.serveMux.HandleFunc("GET /auth/login", serverInst.handleAuthLogin)
	serverInst.serveMux.HandleFunc("GET /auth/callback", serverInst.handleAuthCallback)

	serverInst.serveMux.HandleFunc("GET /rank", serverInst.requireAuth(serverInst.handleGetRank))
	serverInst.serveMux.HandleFunc("POST /rank", serverInst.requireAuth(serverInst.handlePostRank))

	serverInst.serveMux.Handle("GET /static/", http.StripPrefix("/static/", hashfs.FileServer(static.HashFS)))

	return serverInst, nil
}
