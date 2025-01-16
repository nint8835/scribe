package web

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/gorilla/sessions"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/oauth2"
	goldmark_discord_mentions "pkg.nit.so/goldmark-discord-mentions"

	"github.com/nint8835/scribe/pkg/bot"
	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/web/static"
	"github.com/nint8835/scribe/pkg/web/ui/pages"
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

func (s *Server) handleGetHome(w http.ResponseWriter, r *http.Request) {
	pages.Home().Render(r.Context(), w)
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
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.Strikethrough,
				goldmark_discord_mentions.New(bot.Instance.Session, config.Instance.MentionCachePath),
			),
		),
	}

	serverInst.sessionStore.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   30 * 24 * 60 * 60,
	}

	serverInst.serveMux.HandleFunc("GET /auth/login", serverInst.handleAuthLogin)
	serverInst.serveMux.HandleFunc("GET /auth/callback", serverInst.handleAuthCallback)

	serverInst.serveMux.Handle("GET /static/", http.StripPrefix("/static/", hashfs.FileServer(static.HashFS)))

	// All routes below this point require authentication

	serverInst.serveMux.HandleFunc("GET /{$}", serverInst.requireAuth(serverInst.handleGetHome))
	serverInst.serveMux.HandleFunc("GET /leaderboard", serverInst.requireAuth(serverInst.handleGetLeaderboard))

	serverInst.serveMux.HandleFunc("GET /rank", serverInst.requireAuth(serverInst.handleGetRank))
	serverInst.serveMux.HandleFunc("POST /rank", serverInst.requireAuth(serverInst.handlePostRank))
	serverInst.serveMux.HandleFunc("GET /rank/stats", serverInst.requireAuth(serverInst.handleRankStats))

	slog.Info("Web server listening on port 8000")

	return serverInst, nil
}
