package entity

import (
	"time"
	
	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"` // Хэш мастер-пароля
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// OAuth токены для Яндекс.Диска
	YandexDiskToken  string
	YandexDiskExpiry *time.Time
}

func (User) TableName() string {
	return "users"
}