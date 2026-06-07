package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/lmittmann/tint"
)

type Config struct {
	DBPath             string `default:"quotes.sqlite" split_words:"true"`
	Token              string
	OwnerId            string `default:"106162668032802816" split_words:"true"`
	GuildId            string `default:"497544520695808000" split_words:"true"`
	AppId              string `default:"862525831552172045" split_words:"true"`
	CookieSecret       string `split_words:"true"`
	ClientId           string `split_words:"true"`
	ClientSecret       string `split_words:"true"`
	BaseUrl            string `default:"http://localhost:8000" split_words:"true"`
	RunBot             bool   `default:"true" split_words:"true"`
	MentionCachePath   string `default:"mentions.json" split_words:"true"`
	LogLevel           string `default:"info" split_words:"true"`
	TimeZone           string `default:"America/St_Johns" split_words:"true"`
	OnThisDayChannelId string `split_words:"true"`
	OnThisDayCron      string `default:"0 9 * * *" split_words:"true"`
}

var Instance Config
var BaseUrl *url.URL
var Location *time.Location

func Load() error {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		// TODO: Proper logging
		slog.Warn("Failed to load .env file", "error", err)
	}

	err = envconfig.Process("scribe", &Instance)
	if err != nil {
		return fmt.Errorf("error loading config: %s", err)
	}

	level, validLevel := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}[strings.ToLower(Instance.LogLevel)]
	if !validLevel {
		return fmt.Errorf("invalid log level: %s", Instance.LogLevel)
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(
			os.Stderr,
			&tint.Options{
				TimeFormat: time.Kitchen,
				Level:      level,
			},
		),
	))

	Location, err = time.LoadLocation(Instance.TimeZone)
	if err != nil {
		return fmt.Errorf("error loading time zone %q: %w", Instance.TimeZone, err)
	}

	BaseUrl, err = url.Parse(Instance.BaseUrl)
	if err != nil {
		return fmt.Errorf("error parsing base URL: %w", err)
	}

	return nil
}
