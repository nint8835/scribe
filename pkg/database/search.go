package database

import (
	"fmt"

	"gorm.io/gorm/clause"
)

type SearchOptions struct {
	Author *string
	Query  *string
	Limit  int
	Page   int
}

func Search(opts SearchOptions) ([]Quote, int, error) {
	var quotes []Quote

	query := Instance.Model(&Quote{}).Preload(clause.Associations)

	if opts.Author != nil {
		query = query.
			Joins("INNER JOIN quote_authors ON quote_authors.quote_id = quotes.id").
			Where(map[string]interface{}{"quote_authors.author_id": opts.Author})
	}

	if opts.Query != nil {
		// Wrap the provided query in quotes to prevent any special characters from being interpreted as FTS operators
		queryString := fmt.Sprintf("\"%s\"", *opts.Query)

		filterQuery := Instance.Raw("SELECT ROWID FROM quotes_fts WHERE quotes_fts MATCH ?", queryString)
		query.Where("quotes.ROWID IN (?)", filterQuery)
	}

	var total int64
	result := query.Count(&total)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("error counting quotes: %w", result.Error)
	}

	result = query.Limit(opts.Limit).Offset(opts.Limit * (opts.Page - 1)).Find(&quotes)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("error getting quotes: %w", result.Error)
	}

	return quotes, int(total), nil
}
