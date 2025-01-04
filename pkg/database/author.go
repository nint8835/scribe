package database

import (
	"time"

	"gorm.io/gorm"
)

type Author struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Quotes    []*Quote       `gorm:"many2many:quote_authors;"`
}
