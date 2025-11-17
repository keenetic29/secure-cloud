import api from './api';
import { yandexDiskService } from './yandexDisk';

export const storageService = {
  getYandexAuthURL: () => 
    api.get('/storage/yandex/auth-url'),

  handleYandexCallback: (code) => 
    api.post('/storage/yandex/callback', { code }),

  getYandexToken: () =>
    api.get('/storage/yandex/token'),

  // Получаем файлы через прямое обращение к Яндекс.Диску
  // Получаем файлы через прямое обращение к Яндекс.Диску
  getFiles: async (path = '') => {
    const data = await yandexDiskService.getFilesList(path || '/');
    
    // Преобразуем ответ Яндекс.Диска в нашу структуру
    const files = data._embedded.items.map(item => ({
      id: item.resource_id,
      filename: item.name,
      path: item.path,
      size: item.size,
      mimeType: item.mime_type,
      type: item.type, // 'file' или 'dir'
      modified: item.modified,
      created: item.created,
      isEncrypted: item.name.endsWith('.encrypted') // Простая проверка
    }));

    return { data: { files } };
  },

  // Остальные методы теперь работают через прямой API
  uploadFile: (file, encryptedFilename) =>
    yandexDiskService.uploadFile(file, encryptedFilename),

  downloadFile: (filePath) =>
    yandexDiskService.downloadFile(filePath),

  deleteFile: (filePath) =>
    yandexDiskService.deleteFile(filePath),
};