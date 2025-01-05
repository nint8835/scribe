package database

import "gorm.io/gorm"

type CompletedComparison struct {
	Meta gorm.Model `gorm:"embedded"`

	UserID   string
	QuoteAID uint
	QuoteA   Quote `gorm:"foreignKey:QuoteAID;references:ID"`
	QuoteBID uint
	QuoteB   Quote `gorm:"foreignKey:QuoteBID;references:ID"`
}
