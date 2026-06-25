package database

import "time"

type BannedUser struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func IsUserBanned(userID string) bool {
	if Instance == nil {
		return false
	}
	var count int64
	Instance.Model(&BannedUser{}).Where("id = ?", userID).Count(&count)
	return count > 0
}