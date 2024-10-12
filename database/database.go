package database

import (
	"errors"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func Initialize(connectionString string) error {
	newInstance, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error opening db connection: %w", err)
	}
	Instance = newInstance
	return nil
}

func Migrate() {
	if Instance == nil {
		panic(errors.New("Attempted to migrate uninstantiated database - ensure you call database.Initialize before making any database calls"))
	}

	Instance.AutoMigrate(&Quote{}, &Author{})
}
