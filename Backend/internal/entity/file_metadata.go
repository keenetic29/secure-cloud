package entity

import (
	"time"
	
	"gorm.io/gorm"
)

type FileMetadata struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       uint   `gorm:"not null;index"`
	Filename     string `gorm:"not null"`     // Исходное имя файла
	EncryptedName string `gorm:"not null"`    // Зашифрованное имя в облаке
	Path         string `gorm:"not null"`     // Путь в облачном хранилище
	Size         int64  `gorm:"not null"`     // Размер файла в байтах
	MimeType     string                       // MIME-тип
	IsEncrypted  bool   `gorm:"default:true"` // Флаг шифрования
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	
	User User `gorm:"foreignKey:UserID"`
}

func (FileMetadata) TableName() string {
	return "files_metadata"
}