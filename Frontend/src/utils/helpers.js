// Вспомогательные функции

/**
 * Форматирование размера файла
 */
export const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

/**
 * Форматирование даты
 */
export const formatDate = (dateString, options = {}) => {
  const defaultOptions = {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  };
  
  return new Date(dateString).toLocaleDateString('en-US', {
    ...defaultOptions,
    ...options
  });
};

/**
 * Проверка email
 */
export const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

/**
 * Генерация случайной строки
 */
export const generateRandomString = (length = 16) => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
};

/**
 * Обработчик ошибок API
 */
export const handleApiError = (error) => {
  if (error.response) {
    // Сервер ответил с ошибкой
    return error.response.data.error || 'Server error occurred';
  } else if (error.request) {
    // Запрос был сделан, но ответ не получен
    return 'Network error - please check your connection';
  } else {
    // Что-то пошло не так при настройке запроса
    return error.message || 'An unexpected error occurred';
  }
};

/**
 * Задержка выполнения
 */
export const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));