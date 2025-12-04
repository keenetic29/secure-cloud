package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

type EncryptionService struct {
	salt []byte
}

func NewEncryptionService() *EncryptionService {
	// Фиксированная соль для упрощения (в проде должен быть уникальным на файл)
	salt := []byte("secure-cloud-salt-2024")
	return &EncryptionService{salt: salt}
}

// deriveKey создает ключ из мастер-пароля
func (s *EncryptionService) deriveKey(masterPassword string) []byte {
	return pbkdf2.Key([]byte(masterPassword), s.salt, 100000, 32, sha256.New)
}

// EncryptFile шифрует файл
func (s *EncryptionService) EncryptFile(data []byte, masterPassword string) ([]byte, error) {
	key := s.deriveKey(masterPassword)
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	
	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return encrypted, nil
}

// DecryptFile дешифрует файл
func (s *EncryptionService) DecryptFile(encryptedData []byte, masterPassword string) ([]byte, error) {
	key := s.deriveKey(masterPassword)
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed - check master password")
	}
	
	return decrypted, nil
}

// EncryptFilename шифрует имя файла
func (s *EncryptionService) EncryptFilename(filename, masterPassword string) (string, error) {
	key := s.deriveKey(masterPassword)
	
	block, err := aes.NewCipher(key[:16]) // Используем 16 байт для AES-128 для имен
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	encrypted := gcm.Seal(nonce, nonce, []byte(filename), nil)
	
	// Конвертируем в base64 для использования в имени файла
	encoded := base64.URLEncoding.EncodeToString(encrypted)
	return encoded + ".encrypted", nil
}

// DecryptFilename дешифрует имя файла
func (s *EncryptionService) DecryptFilename(encryptedFilename, masterPassword string) (string, error) {
	if !strings.HasSuffix(encryptedFilename, ".encrypted") {
		return encryptedFilename, nil
	}
	
	// Убираем расширение
	base64Str := strings.TrimSuffix(encryptedFilename, ".encrypted")
	
	// Декодируем base64
	encryptedData, err := base64.URLEncoding.DecodeString(base64Str)
	if err != nil {
		return "", err
	}
	
	key := s.deriveKey(masterPassword)
	
	block, err := aes.NewCipher(key[:16])
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", errors.New("invalid encrypted filename")
	}
	
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt filename: %w", err)
	}
	
	return string(decrypted), nil
}