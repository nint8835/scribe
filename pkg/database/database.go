package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func initNonGormResources() error {
	var quoteFtsTableExists int

	err := Instance.
		Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='quotes_fts'").
		Scan(&quoteFtsTableExists).
		Error
	if err != nil {
		return fmt.Errorf("error checking for quotes_fts table: %w", err)
	}

	if quoteFtsTableExists == 0 {
		err = Instance.Exec("CREATE VIRTUAL TABLE quotes_fts USING fts5(text, content=quotes)").Error
		if err != nil {
			return fmt.Errorf("error creating quotes_fts table: %w", err)
		}

		err = Instance.Exec(`INSERT INTO quotes_fts (rowid, text) SELECT ROWID, "text" FROM quotes`).Error
		if err != nil {
			return fmt.Errorf("error populating quotes_fts table: %w", err)
		}

		err = Instance.Exec(`
		CREATE TRIGGER quotes_fts_insert AFTER INSERT ON quotes BEGIN
			INSERT INTO quotes_fts (rowid, text) VALUES (new.rowid, new."text");
		END
		`).Error
		if err != nil {
			return fmt.Errorf("error creating quotes_fts_insert trigger: %w", err)
		}

		err = Instance.Exec(`
		CREATE TRIGGER quotes_fts_update AFTER UPDATE ON quotes BEGIN
			INSERT INTO quotes_fts (quotes_fts, rowid, text) VALUES ('delete', old.rowid, old."text");
			INSERT INTO quotes_fts (rowid, text) VALUES (new.rowid, new."text");
		END
		`).Error
		if err != nil {
			return fmt.Errorf("error creating quotes_fts_update trigger: %w", err)
		}

		err = Instance.Exec(`
		CREATE TRIGGER quotes_fts_delete AFTER DELETE ON quotes BEGIN
			INSERT INTO quotes_fts (quotes_fts, rowid, text) VALUES ('delete', old.rowid, old."text");
		END
		`).Error
		if err != nil {
			return fmt.Errorf("error creating quotes_fts_delete trigger: %w", err)
		}
	}

	return nil
}

func Initialize(connectionString string) error {
	newInstance, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error opening db connection: %w", err)
	}
	Instance = newInstance

	Instance.AutoMigrate(&Quote{}, &Author{}, &CompletedComparison{})

	err = initNonGormResources()
	if err != nil {
		return fmt.Errorf("error initializing non-GORM resources: %w", err)
	}

	return nil
}
