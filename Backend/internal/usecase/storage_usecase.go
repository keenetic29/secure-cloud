package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"server/internal/entity"
	"server/internal/repository"
	"server/pkg/encryption"
	"server/pkg/yandex_disk"
)

type storageUseCase struct {
	fileRepo     repository.FileMetadataRepository
	userRepo     repository.UserRepository
	yandexDisk   *yandex_disk.Client
	encryption   *encryption.EncryptionService
}

func NewStorageUseCase(
	fileRepo repository.FileMetadataRepository,
	userRepo repository.UserRepository,
	yandexDisk *yandex_disk.Client,
) StorageUseCase {
	return &storageUseCase{
		fileRepo:     fileRepo,
		userRepo:     userRepo,
		yandexDisk:   yandexDisk,
		encryption:   encryption.NewEncryptionService(),
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
		fmt.Printf("DEBUG: User %d not found: %v\n", userID, err)
		return nil, errors.New("user not found")
	}

	if user.YandexDiskToken == "" {
		fmt.Printf("DEBUG: Yandex.Disk not connected for user %d\n", userID)
		return nil, errors.New("yandex disk not connected")
	}

	// Нормализуем путь
	if path == "" {
		path = "/"
	}
	
	fmt.Printf("DEBUG: Getting files from Yandex.Disk for user %d, path: '%s'\n", userID, path)

	// Получаем файлы из Яндекс.Диска
	diskResp, err := uc.yandexDisk.GetFilesList(ctx, user.YandexDiskToken, path)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get files from Yandex.Disk: %v\n", err)
		return nil, fmt.Errorf("failed to get files from yandex disk: %w", err)
	}

	fmt.Printf("DEBUG: Got %d items from Yandex.Disk path '%s'\n", len(diskResp.Embedded.Items), path)

	// Для каждого элемента создаем метаданные
	var filesMetadata []*entity.FileMetadata
	for _, item := range diskResp.Embedded.Items {
		fmt.Printf("DEBUG: Processing item: %s (type: %s, path: %s, size: %d)\n", 
			item.Name, item.Type, item.Path, item.Size)
		
		// Определяем, является ли файл зашифрованным
		isEncrypted := strings.HasSuffix(item.Name, ".encrypted")
		
		// Определяем оригинальное имя
		originalName := item.Name
		if isEncrypted {
			// Пытаемся извлечь оригинальное имя из зашифрованного
			originalName = strings.TrimSuffix(item.Name, ".encrypted")
			// Если это base64 строка, оставляем как есть или преобразуем
			if len(originalName) > 50 {
				originalName = "encrypted_file"
			}
		}
		
		// Определяем тип (dir или file)
		itemType := "file"
		if item.Type == "dir" {
			itemType = "dir"
		}
		
		// Определяем MIME тип
		mimeType := item.MimeType
		if mimeType == "" {
			if itemType == "dir" {
				mimeType = "directory"
			} else if strings.HasSuffix(strings.ToLower(item.Name), ".jpg") || 
			   strings.HasSuffix(strings.ToLower(item.Name), ".jpeg") {
				mimeType = "image/jpeg"
			} else if strings.HasSuffix(strings.ToLower(item.Name), ".png") {
				mimeType = "image/png"
			} else if strings.HasSuffix(strings.ToLower(item.Name), ".pdf") {
				mimeType = "application/pdf"
			} else if strings.HasSuffix(strings.ToLower(item.Name), ".txt") {
				mimeType = "text/plain"
			} else {
				mimeType = "application/octet-stream"
			}
		}
		
		// Проверяем, есть ли уже метаданные в БД
		fileMetadata, err := uc.fileRepo.GetFileByPath(ctx, userID, item.Path)
		if err != nil {
			// Создаем новую запись
			fileMetadata = &entity.FileMetadata{
				UserID:        userID,
				Filename:      originalName,
				EncryptedName: item.Name,
				Path:          item.Path,
				Size:          item.Size,
				MimeType:      mimeType,
				IsEncrypted:   isEncrypted,
				Type:          itemType,
			}
			
			// Сохраняем в БД только файлы (не папки)
			if itemType == "file" {
				err = uc.fileRepo.CreateFileMetadata(ctx, fileMetadata)
				if err != nil {
					fmt.Printf("DEBUG: Could not save metadata for %s: %v\n", item.Name, err)
					// Продолжаем даже если не удалось сохранить
				}
			}
			
			fmt.Printf("DEBUG: Created metadata for %s: %s (encrypted: %v, type: %s, size: %d)\n", 
				item.Type, originalName, isEncrypted, itemType, item.Size)
		} else {
			// Обновляем существующую запись
			fileMetadata.Filename = originalName
			fileMetadata.EncryptedName = item.Name
			fileMetadata.Size = item.Size
			fileMetadata.MimeType = mimeType
			fileMetadata.IsEncrypted = isEncrypted
			fileMetadata.Type = itemType
			
			// Обновляем в БД
			if err := uc.fileRepo.UpdateFileMetadata(ctx, fileMetadata); err != nil {
				fmt.Printf("DEBUG: Could not update metadata for %s: %v\n", item.Name, err)
			}
			
			fmt.Printf("DEBUG: Updated metadata for %s: %s (ID: %d, type: %s, size: %d)\n", 
				item.Type, fileMetadata.Filename, fileMetadata.ID, itemType, item.Size)
		}
		
		// Добавляем в результат
		filesMetadata = append(filesMetadata, fileMetadata)
	}

	fmt.Printf("DEBUG: Returning %d items (files + folders)\n", len(filesMetadata))
	return filesMetadata, nil
}

func (uc *storageUseCase) UploadFile(ctx context.Context, userID uint, fileHeader *multipart.FileHeader, masterPassword, path string) (*entity.FileMetadata, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.YandexDiskToken == "" {
		return nil, errors.New("yandex disk not connected")
	}

	// Открываем файл
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Читаем содержимое файла
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Шифруем файл
	encryptedContent, err := uc.encryption.EncryptFile(fileContent, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt file: %w", err)
	}

	// Шифруем имя файла
	encryptedFilename, err := uc.encryption.EncryptFilename(fileHeader.Filename, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt filename: %w", err)
	}

	// Формируем полный путь
	fullPath := path
	if path != "/" && !strings.HasSuffix(path, "/") {
		fullPath += "/"
	}
	fullPath += encryptedFilename

	// Загружаем зашифрованный файл в Яндекс.Диск
	reader := bytes.NewReader(encryptedContent)
	err = uc.yandexDisk.UploadFile(ctx, user.YandexDiskToken, fullPath, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to yandex disk: %w", err)
	}

	// Сохраняем метаданные в БД
	fileMetadata := &entity.FileMetadata{
		UserID:        userID,
		Filename:      fileHeader.Filename,
		EncryptedName: encryptedFilename,
		Path:          fullPath,
		Size:          int64(len(encryptedContent)),
		MimeType:      fileHeader.Header.Get("Content-Type"),
		IsEncrypted:   true,
		Type:          "file",
	}

	err = uc.fileRepo.CreateFileMetadata(ctx, fileMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return fileMetadata, nil
}

func (uc *storageUseCase) DownloadFile(ctx context.Context, userID uint, fileID uint, masterPassword string) ([]byte, string, error) {
	// Получаем метаданные файла
	fileMetadata, err := uc.fileRepo.GetFileMetadataByID(ctx, fileID)
	if err != nil {
		return nil, "", errors.New("file not found")
	}

	if fileMetadata.UserID != userID {
		return nil, "", errors.New("access denied")
	}

	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", errors.New("user not found")
	}

	// Скачиваем зашифрованный файл из Яндекс.Диска
	reader, err := uc.yandexDisk.DownloadFile(ctx, user.YandexDiskToken, fileMetadata.Path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}
	defer reader.Close()

	// Читаем зашифрованное содержимое
	encryptedContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Дешифруем файл
	decryptedContent, err := uc.encryption.DecryptFile(encryptedContent, masterPassword)
	if err != nil {
		return nil, "", fmt.Errorf("decryption failed: %w", err)
	}

	return decryptedContent, fileMetadata.Filename, nil
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

func (uc *storageUseCase) DeleteFile(ctx context.Context, userID uint, fileID uint) error {
	file, err := uc.fileRepo.GetFileMetadataByID(ctx, fileID)
	if err != nil {
		return errors.New("file not found")
	}

	if file.UserID != userID {
		return errors.New("access denied")
	}

	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Удаляем файл из Яндекс.Диска
	err = uc.yandexDisk.DeleteFile(ctx, user.YandexDiskToken, file.Path)
	if err != nil {
		return err
	}

	// Удаляем запись из БД
	return uc.fileRepo.DeleteFileMetadata(ctx, fileID)
}

func (uc *storageUseCase) GetDecryptedFilename(ctx context.Context, userID uint, fileID uint, masterPassword string) (string, error) {
	file, err := uc.fileRepo.GetFileMetadataByID(ctx, fileID)
	if err != nil {
		return "", errors.New("file not found")
	}

	if file.UserID != userID {
		return "", errors.New("access denied")
	}

	// Дешифруем имя файла
	decryptedName, err := uc.encryption.DecryptFilename(file.EncryptedName, masterPassword)
	if err != nil {
		return file.Filename, nil // Возвращаем сохраненное имя если не удалось расшифровать
	}

	return decryptedName, nil
}