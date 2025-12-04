package postgres

import (
	"context"
	"fmt"
	
	"gorm.io/gorm"
	
	"server/internal/entity"
	"server/internal/repository"
)

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) repository.FileMetadataRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFileMetadata(ctx context.Context, file *entity.FileMetadata) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepository) GetFileMetadataByID(ctx context.Context, id uint) (*entity.FileMetadata, error) {
	var file entity.FileMetadata
	err := r.db.WithContext(ctx).Preload("User").First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) GetUserFiles(ctx context.Context, userID uint, path string) ([]*entity.FileMetadata, error) {
	var files []*entity.FileMetadata
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if path != "" {
		query = query.Where("path = ?", path)
	}
	err := query.Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *fileRepository) UpdateFileMetadata(ctx context.Context, file *entity.FileMetadata) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *fileRepository) DeleteFileMetadata(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.FileMetadata{}, id).Error
}

func (r *fileRepository) GetFileByPath(ctx context.Context, userID uint, path string) (*entity.FileMetadata, error) {
	var file entity.FileMetadata
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND path = ?", userID, path).
		First(&file).Error
	if err != nil {
		fmt.Printf("DEBUG: No existing metadata found for path %s (user %d): %v\n", path, userID, err)
		return nil, err
	}
	
	fmt.Printf("DEBUG: Found existing metadata for path %s: %s (ID: %d)\n", 
		path, file.Filename, file.ID)
	return &file, nil
}