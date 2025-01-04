package database

type CompletedComparison struct {
	UserID   string `gorm:"primaryKey"`
	QuoteAID uint
	QuoteA   Quote `gorm:"primaryKey;foreignKey:QuoteAID;references:ID"`
	QuoteBID uint
	QuoteB   Quote `gorm:"primaryKey;foreignKey:QuoteBID;references:ID"`
}
