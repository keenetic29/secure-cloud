package repository

import (
	"context"
	
	"server/internal/entity"
)

// UserRepository определяет контракт для работы с пользователями
type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
}

// FileMetadataRepository определяет контракт для работы с метаданными файлов
type FileMetadataRepository interface {
	CreateFileMetadata(ctx context.Context, file *entity.FileMetadata) error
	GetFileMetadataByID(ctx context.Context, id uint) (*entity.FileMetadata, error)
	GetUserFiles(ctx context.Context, userID uint, path string) ([]*entity.FileMetadata, error)
	UpdateFileMetadata(ctx context.Context, file *entity.FileMetadata) error
	DeleteFileMetadata(ctx context.Context, id uint) error
	GetFileByPath(ctx context.Context, userID uint, path string) (*entity.FileMetadata, error)
}