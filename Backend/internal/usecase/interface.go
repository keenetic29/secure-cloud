package usecase

import (
	"context"
	"mime/multipart"
	
	"server/internal/entity"
)

// AuthUseCase определяет контракт для аутентификации
type AuthUseCase interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (uint, error)
}

// UserUseCase определяет контракт для работы с пользователями
type UserUseCase interface {
	GetUser(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
}

// StorageUseCase определяет контракт для работы с облачным хранилищем
type StorageUseCase interface {
	GetAuthURL(ctx context.Context) string
	HandleCallback(ctx context.Context, code string, userID uint) error
	GetFiles(ctx context.Context, userID uint, path string) ([]*entity.FileMetadata, error)
	GetFileInfo(ctx context.Context, userID uint, fileID uint) (*entity.FileMetadata, error)
	GetDecryptedFilename(ctx context.Context, userID uint, fileID uint, masterPassword string) (string, error)
	UploadFile(ctx context.Context, userID uint, fileHeader *multipart.FileHeader, masterPassword, path string) (*entity.FileMetadata, error)
	DownloadFile(ctx context.Context, userID uint, fileID uint, masterPassword string) ([]byte, string, error)
	DeleteFile(ctx context.Context, userID uint, fileID uint) error
	GetYandexToken(ctx context.Context, userID uint) (string, error)
}