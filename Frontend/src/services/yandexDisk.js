import api from './api';
import axios from 'axios';

class YandexDiskService {
  constructor() {
    this.baseURL = 'https://cloud-api.yandex.net/v1/disk';
  }

  async getAccessToken() {
    try {
      const response = await api.get('/storage/yandex/token');
      return response.data.access_token;
    } catch (error) {
      throw new Error('Failed to get Yandex.Disk access token');
    }
  }

  async getUploadUrl(path, accessToken) {
    const response = await axios.get(
      `${this.baseURL}/resources/upload`,
      {
        params: { 
          path: path,
          overwrite: true
        },
        headers: {
          'Authorization': `OAuth ${accessToken}`
        }
      }
    );
    return response.data.href;
  }

  async getDownloadUrl(path, accessToken) {
    const response = await axios.get(
      `${this.baseURL}/resources/download`,
      {
        params: { path: path },
        headers: {
          'Authorization': `OAuth ${accessToken}`
        }
      }
    );
    return response.data.href;
  }

  async uploadFile(file, encryptedFilename, onProgress = null) {
    try {
      const accessToken = await this.getAccessToken();
      
      // Получаем URL для загрузки
      const uploadUrl = await this.getUploadUrl(encryptedFilename, accessToken);
      
      // Загружаем файл напрямую в Яндекс.Диск
      await axios.put(uploadUrl, file, {
        headers: {
          'Content-Type': 'application/octet-stream',
        },
        onUploadProgress: (progressEvent) => {
          if (onProgress && progressEvent.total) {
            const percentCompleted = Math.round(
              (progressEvent.loaded * 100) / progressEvent.total
            );
            onProgress(percentCompleted);
          }
        },
      });

      return encryptedFilename;
    } catch (error) {
      console.error('Upload error:', error);
      throw new Error(`Upload failed: ${error.response?.data?.message || error.message}`);
    }
  }

  async downloadFile(filePath) {
    try {
      const accessToken = await this.getAccessToken();
      
      // Получаем URL для скачивания
      const downloadUrl = await this.getDownloadUrl(filePath, accessToken);
      
      // Скачиваем файл напрямую из Яндекс.Диска
      const response = await axios.get(downloadUrl, {
        responseType: 'blob',
      });

      return response.data;
    } catch (error) {
      console.error('Download error:', error);
      throw new Error(`Download failed: ${error.response?.data?.message || error.message}`);
    }
  }

  async getFilesList(path = '/') {
  try {
    const accessToken = await this.getAccessToken();
    
    // Нормализуем путь
    let normalizedPath = path;
    
    if (!path || path === '/' || path === '') {
      normalizedPath = '/';
    } else {
      normalizedPath = path.startsWith('/') ? path : '/' + path;
    }
    
    console.log('Requesting files from path:', normalizedPath);
    
    const response = await axios.get(
      `${this.baseURL}/resources`,
      {
        params: { 
          path: normalizedPath,
          limit: 100,
          sort: '-created'
        },
        headers: {
          'Authorization': `OAuth ${accessToken}`
        }
      }
    );

    console.log('Successfully got files from:', normalizedPath);
    return response.data;
  } catch (error) {
    console.error('Get files error for path:', path);
    console.error('Error response:', error.response?.data);
    
    // Более информативное сообщение об ошибке
    const errorMessage = error.response?.data?.message || error.message;
    throw new Error(`Failed to get files from "${path}": ${errorMessage}`);
  }
}

  async deleteFile(filePath) {
    try {
      const accessToken = await this.getAccessToken();
      
      await axios.delete(
        `${this.baseURL}/resources`,
        {
          params: { 
            path: filePath,
            permanently: true
          },
          headers: {
            'Authorization': `OAuth ${accessToken}`
          }
        }
      );
    } catch (error) {
      console.error('Delete error:', error);
      throw new Error(`Delete failed: ${error.response?.data?.message || error.message}`);
    }
  }
}

export const yandexDiskService = new YandexDiskService();