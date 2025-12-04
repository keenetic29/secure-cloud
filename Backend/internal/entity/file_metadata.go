package entity

import (
	"time"
	
	"gorm.io/gorm"
)

type FileMetadata struct {
    ID           uint   `gorm:"primaryKey" json:"id"`
    UserID       uint   `gorm:"not null;index" json:"user_id"`
    Filename     string `gorm:"not null" json:"filename"`     // Исходное имя файла
    EncryptedName string `gorm:"not null" json:"encrypted_name"`    // Зашифрованное имя в облаке
    Path         string `gorm:"not null" json:"path"`     // Путь в облачном хранилище
    Size         int64  `gorm:"not null" json:"size"`     // Размер файла в байтах
    MimeType     string `json:"mime_type"`                // MIME-тип
    IsEncrypted  bool   `gorm:"default:true" json:"is_encrypted"` // Флаг шифрования
    Type         string `gorm:"default:'file'" json:"type"` // 'file' или 'dir'
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
    
    User User `gorm:"foreignKey:UserID" json:"-"`
}

func (FileMetadata) TableName() string {
	return "files_metadata"
}