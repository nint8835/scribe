package database

import "gorm.io/gorm"

type CompletedComparison struct {
	Meta gorm.Model `gorm:"embedded"`

	UserID   string `gorm:"index"`
	QuoteAID uint   `gorm:"index"`
	QuoteA   Quote  `gorm:"foreignKey:QuoteAID;references:ID"`
	QuoteBID uint   `gorm:"index"`
	QuoteB   Quote  `gorm:"foreignKey:QuoteBID;references:ID"`
	WinnerID uint
	Winner   Quote `gorm:"foreignKey:WinnerID;references:ID"`
}
