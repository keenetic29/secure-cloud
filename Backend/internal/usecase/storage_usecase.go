package usecase

import (
	"context"
	"errors"
	"time"
	
	"server/internal/entity"
	"server/internal/repository"
	"server/pkg/yandex_disk"
)

type storageUseCase struct {
	fileRepo    repository.FileMetadataRepository
	userRepo    repository.UserRepository
	yandexDisk  *yandex_disk.Client
}

func NewStorageUseCase(
	fileRepo repository.FileMetadataRepository,
	userRepo repository.UserRepository,
	yandexDisk *yandex_disk.Client,
) StorageUseCase {
	return &storageUseCase{
		fileRepo:   fileRepo,
		userRepo:   userRepo,
		yandexDisk: yandexDisk,
	}
}

func (uc *storageUseCase) GetAuthURL(ctx context.Context) string {
	return uc.yandexDisk.GetAuthURL()
}

func (uc *storageUseCase) HandleCallback(ctx context.Context, code string, userID uint) error {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}
	
	tokenResp, err := uc.yandexDisk.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return err
	}
	
	// Сохраняем токен в базе данных
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	user.YandexDiskToken = tokenResp.AccessToken
	user.YandexDiskExpiry = &expiry
	
	return uc.userRepo.UpdateUser(ctx, user)
}

func (uc *storageUseCase) GetYandexToken(ctx context.Context, userID uint) (string, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	if user.YandexDiskToken == "" {
		return "", errors.New("yandex disk not connected")
	}

	return user.YandexDiskToken, nil
}

func (uc *storageUseCase) GetFiles(ctx context.Context, userID uint, path string) ([]*entity.FileMetadata, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	
	if user.YandexDiskToken == "" {
		return nil, errors.New("yandex disk not connected")
	}
	
	// Получаем файлы из Яндекс.Диска
	diskResp, err := uc.yandexDisk.GetFilesList(ctx, user.YandexDiskToken, path)
	if err != nil {
		return nil, err
	}
	
	// Преобразуем в нашу структуру метаданных
	var files []*entity.FileMetadata
	for _, item := range diskResp.Embedded.Items {
		file := &entity.FileMetadata{
			UserID:       userID,
			Filename:     item.Name,
			EncryptedName: item.Name, // В реальной системе здесь было бы зашифрованное имя
			Path:         item.Path,
			Size:         item.Size,
			MimeType:     item.MimeType,
			IsEncrypted:  false, // Яндекс.Диск возвращает исходные файлы
		}
		files = append(files, file)
	}
	
	return files, nil
}

func (uc *storageUseCase) DeleteFile(ctx context.Context, userID uint, fileID uint) error {
	file, err := uc.fileRepo.GetFileMetadataByID(ctx, fileID)
	if err != nil {
		return errors.New("file not found")
	}
	
	if file.UserID != userID {
		return errors.New("access denied")
	}
	
	return uc.fileRepo.DeleteFileMetadata(ctx, fileID)
}

func (uc *storageUseCase) GetFileInfo(ctx context.Context, userID uint, fileID uint) (*entity.FileMetadata, error) {
	file, err := uc.fileRepo.GetFileMetadataByID(ctx, fileID)
	if err != nil {
		return nil, errors.New("file not found")
	}
	
	if file.UserID != userID {
		return nil, errors.New("access denied")
	}
	
	return file, nil
}