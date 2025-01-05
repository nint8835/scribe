package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBPath       string `default:"quotes.sqlite" split_words:"true"`
	Token        string
	OwnerId      string `default:"106162668032802816" split_words:"true"`
	GuildId      string `default:"497544520695808000" split_words:"true"`
	AppId        string `default:"862525831552172045" split_words:"true"`
	CookieSecret string `split_words:"true"`
	ClientId     string `split_words:"true"`
	ClientSecret string `split_words:"true"`
	CallbackUrl  string `split_words:"true"`
	SyncCommands bool   `default:"true" split_words:"true"`
}

var Instance Config

func Load() error {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		// TODO: Proper logging
		fmt.Printf("Failed to load .env file: %s\n", err)
	}

	err = envconfig.Process("scribe", &Instance)
	if err != nil {
		return fmt.Errorf("error loading config: %s", err)
	}

	return nil
}
