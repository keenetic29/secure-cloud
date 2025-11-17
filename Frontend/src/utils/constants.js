// Константы приложения
export const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

// Коды ошибок
export const ERROR_CODES = {
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  SERVER_ERROR: 500
};

// Сообщения для пользователя
export const MESSAGES = {
  UPLOAD_SUCCESS: 'File uploaded and encrypted successfully',
  UPLOAD_FAILED: 'Failed to upload file',
  DELETE_SUCCESS: 'File deleted successfully',
  DELETE_FAILED: 'Failed to delete file',
  CONNECTION_SUCCESS: 'Cloud storage connected successfully',
  CONNECTION_FAILED: 'Failed to connect cloud storage'
};

// Настройки шифрования
export const CRYPTO_CONFIG = {
  ALGORITHM: 'AES-GCM',
  KEY_LENGTH: 256,
  PBKDF2_ITERATIONS: 100000,
  SALT_LENGTH: 16,
  IV_LENGTH: 12
};