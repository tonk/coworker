package models

type FavoriteUser struct {
	UserID         uint `gorm:"primaryKey" json:"user_id"`
	FavoriteUserID uint `gorm:"primaryKey" json:"favorite_user_id"`
}
