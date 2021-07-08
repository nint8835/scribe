package database

import "gorm.io/gorm"

type Quote struct {
	Meta    gorm.Model `gorm:"embedded"`
	Text    string
	Authors []*Author `gorm:"many2many:quote_authors;"`
	Source  *string
}
