import api from './api';

export const storageService = {
  // Получение списка файлов
  getFiles: (path = '/') => 
    api.get('/storage/files', { params: { path } }),

  // Получение информации о файле
  getFileInfo: (fileId) => 
    api.get(`/storage/files/${fileId}`),

  // Получение расшифрованного имени файла
  getDecryptedFilename: (fileId, masterPassword) => 
    api.post(`/storage/files/${fileId}/decrypt-name`, { master_password: masterPassword }),

  // Загрузка файла
  uploadFile: (file, masterPassword, path = '/') => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('master_password', masterPassword);
    
    return api.post(`/storage/upload?path=${encodeURIComponent(path)}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
  },

  // Скачивание файла
  downloadFile: (fileId, masterPassword) => 
    api.post(`/storage/files/${fileId}/download`, 
      { master_password: masterPassword },
      { 
        responseType: 'blob',
        // Добавляем обработку заголовков
        transformResponse: [(data, headers) => {
          return {
            data: data,
            headers: headers
          };
        }]
      }
    ).then(response => {
      // Нормализуем ответ для единообразной обработки
      return {
        data: response.data.data,
        headers: response.data.headers || response.headers
      };
    }),

  // Удаление файла
  deleteFile: (fileId) => 
    api.delete(`/storage/files/${fileId}`),

  // Яндекс.Диск OAuth
  getYandexAuthURL: () => 
    api.get('/storage/yandex/auth-url'),

  handleYandexCallback: (code) => 
    api.post('/storage/yandex/callback', { code }),

  getYandexToken: () =>
    api.get('/storage/yandex/token'),
};