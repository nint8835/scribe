package database

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nint8835/scribe/pkg/embedding"
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
		err = initializeFts5()
		if err != nil {
			return fmt.Errorf("error initializing FTS5: %w", err)
		}
	}

	var quoteEmbeddingsTableExists int
	err = Instance.
		Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='quote_embeddings'").
		Scan(&quoteEmbeddingsTableExists).
		Error
	if err != nil {
		return fmt.Errorf("error checking for quote_embeddings table: %w", err)
	}

	if quoteEmbeddingsTableExists == 0 {
		err = initializeQuoteEmbeddings()
		if err != nil {
			return fmt.Errorf("error initializing quote embeddings: %w", err)
		}
	}

	err = Instance.Exec("CREATE INDEX IF NOT EXISTS idx_comparisons_user_b_a ON completed_comparisons(user_id, quote_b_id, quote_a_id)").Error
	if err != nil {
		return fmt.Errorf("error creating idx_comparisons_user_b_a index: %w", err)
	}

	err = Instance.Exec("CREATE INDEX IF NOT EXISTS idx_comparisons_user_a_b ON completed_comparisons(user_id, quote_a_id, quote_b_id)").Error
	if err != nil {
		return fmt.Errorf("error creating idx_comparisons_user_a_b index: %w", err)
	}

	return nil
}

func initializeFts5() error {
	slog.Debug("Initializing FTS5")
	err := Instance.Exec("CREATE VIRTUAL TABLE quotes_fts USING fts5(text, content=quotes)").Error
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
	return nil
}

func initializeQuoteEmbeddings() error {
	slog.Debug("Initializing quote embeddings")
	err := Instance.Exec(`CREATE VIRTUAL TABLE quote_embeddings USING vec0(embedding float[384] distance_metric=cosine)`).Error
	if err != nil {
		return fmt.Errorf("error creating quote_embeddings table: %w", err)
	}

	var quotes []Quote
	err = Instance.Find(&quotes).Error
	if err != nil {
		return fmt.Errorf("error querying quotes for backfill: %w", err)
	}

	slog.Info("Backfilling quote embeddings - this may take a while", "count", len(quotes))

	for _, quote := range quotes {
		slog.Info("Backfilling embedding for quote", "id", quote.Meta.ID)
		encodedEmbedding, err := embedding.EmbedQuote(quote.Text)
		if err != nil {
			log.Printf("error embedding quote ID %d for backfill: %v", quote.Meta.ID, err)
			continue
		}

		err = Instance.Exec(
			"INSERT INTO quote_embeddings(rowid, embedding) VALUES(?, ?)",
			quote.Meta.ID,
			encodedEmbedding,
		).Error
		if err != nil {
			log.Printf("error inserting embedding for quote ID %d for backfill: %v", quote.Meta.ID, err)
			continue
		}
	}

	slog.Info("Finished backfilling quote embeddings")

	return nil
}

func Initialize(connectionString string) error {
	sqlite_vec.Auto()

	newInstance, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             1 * time.Second,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
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
