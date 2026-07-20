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

// UnbanUser removes the user with the given ID from the ban list. It returns
// true when a row was actually deleted, and false if the user was not banned.
func UnbanUser(userID string) (bool, error) {
	if Instance == nil {
		return false, nil
	}
	result := Instance.Delete(&BannedUser{}, "id = ?", userID)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
