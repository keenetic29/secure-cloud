import { useState, useEffect } from 'react';
import { storageService } from '../services/storage';

export const useStorage = (path = '/') => {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchFiles = async () => {
    try {
      setLoading(true);
      console.log(`Fetching files from path: ${path}`);
      
      const response = await storageService.getFiles(path);
      console.log('API Response:', response);
      console.log('Response data:', response.data);
      
      if (!response.data) {
        console.warn('No data in response');
        setFiles([]);
        setError('');
        return;
      }
      
      // Обрабатываем ответ с заглавными полями
      const normalizeFileData = (file) => {
        // Если поля уже в нижнем регистре, оставляем как есть
        if (file.id !== undefined) {
          return file;
        }
        
        // Конвертируем поля из заглавных в строчные
        return {
          id: file.ID || file.id || 0,
          user_id: file.UserID || file.user_id || 0,
          filename: file.Filename || file.filename || '',
          encrypted_name: file.EncryptedName || file.encrypted_name || '',
          path: file.Path || file.path || '',
          size: file.Size || file.size || 0,
          mime_type: file.MimeType || file.mime_type || '',
          is_encrypted: file.IsEncrypted || file.is_encrypted || false,
          type: file.Type || file.type || 'file',
          created_at: file.CreatedAt || file.created_at,
          updated_at: file.UpdatedAt || file.updated_at,
          deleted_at: file.DeletedAt || file.deleted_at,
          // Сохраняем оригинальные данные
          ...file
        };
      };
      
      const filesArray = response.data.files || [];
      console.log('Files array (raw):', filesArray);
      
      // Нормализуем данные файлов
      const processedFiles = filesArray.map((file, index) => {
        const normalizedFile = normalizeFileData(file);
        console.log(`Processing file ${index}:`, normalizedFile);
        
        // Определяем тип (файл или папка)
        let type = normalizedFile.type;
        if (!type) {
          type = normalizedFile.mime_type === 'directory' ? 'dir' : 'file';
        }
        
        // Определяем имя для отображения
        let displayName = normalizedFile.filename;
        if (!displayName || displayName === 'unknown' || displayName === 'encrypted_file') {
          displayName = normalizedFile.encrypted_name || 
                       normalizedFile.path?.split('/').pop() || 
                       `item_${index}`;
        }
        
        // Обрабатываем путь
        let processedPath = normalizedFile.path;
        if (processedPath && processedPath.startsWith('disk:')) {
          processedPath = processedPath.substring(5); // Убираем "disk:"
        }
        
        // Определяем, зашифрован ли файл
        const isEncrypted = normalizedFile.is_encrypted || 
                           (normalizedFile.encrypted_name && 
                            normalizedFile.encrypted_name.endsWith('.encrypted'));
        
        return {
          id: normalizedFile.id || index,
          type: type.toLowerCase(),
          filename: displayName,
          path: processedPath,
          original_path: normalizedFile.path,
          is_encrypted: isEncrypted,
          size: normalizedFile.size || 0,
          mime_type: normalizedFile.mime_type || (type === 'dir' ? 'directory' : 'application/octet-stream'),
          encrypted_name: normalizedFile.encrypted_name,
          created_at: normalizedFile.created_at,
          updated_at: normalizedFile.updated_at,
          // Сохраняем все нормализованные данные
          ...normalizedFile
        };
      });
      
      console.log(`Processed ${processedFiles.length} items:`, processedFiles);
      setFiles(processedFiles);
      setError('');
      
    } catch (error) {
      console.error('Error fetching files:', error);
      console.error('Error details:', error.response?.data || error.message);
      
      // Показываем более конкретное сообщение об ошибке
      if (error.response?.status === 401) {
        setError('Authentication failed. Please login again.');
      } else if (error.response?.status === 403) {
        setError('Access denied. Check your permissions.');
      } else if (error.response?.data?.error) {
        setError(`Server error: ${error.response.data.error}`);
      } else if (error.message.includes('Network Error')) {
        setError('Network error. Check your connection and server.');
      } else {
        setError(`Failed to load files: ${error.message}`);
      }
      
      setFiles([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFiles();
  }, [path]);

  const refresh = () => {
    fetchFiles();
  };

  const uploadFile = async (file, masterPassword) => {
    try {
      const response = await storageService.uploadFile(file, masterPassword, path);
      await fetchFiles(); // Обновляем список
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Upload error:', error);
      return { 
        success: false, 
        error: error.response?.data?.error || 'Upload failed' 
      };
    }
  };

  const downloadFile = async (fileId, masterPassword, originalFilename = null) => {
  try {
    const response = await storageService.downloadFile(fileId, masterPassword);
    
    // Получаем имя файла из заголовков
    const contentDisposition = response.headers['content-disposition'];
    let filename = 'downloaded_file';
    
    if (contentDisposition) {
      // Извлекаем имя файла из заголовка Content-Disposition
      const filenameMatch = contentDisposition.match(/filename="([^"]+)"/);
      if (filenameMatch && filenameMatch.length === 2) {
        filename = decodeURIComponent(filenameMatch[1]);
        console.log('Filename from headers:', filename);
      }
    }
    
    // Если имя файла не извлечено из заголовков, используем переданное оригинальное имя
    if ((filename === 'downloaded_file' || !filename.includes('.')) && originalFilename) {
      filename = originalFilename;
      console.log('Using original filename:', filename);
    }
    
    // Убедимся, что у файла есть расширение
    if (!filename.includes('.')) {
      console.log('No extension in filename, trying to determine...');
      
      // 1. Проверим MIME-тип из заголовков ответа
      const contentType = response.headers['content-type'];
      if (contentType) {
        const ext = getExtensionFromMimeType(contentType);
        if (ext) {
          filename += '.' + ext;
          console.log('Added extension from MIME type:', ext);
        }
      }
      
      // 2. Если все еще нет расширения, проверим исходные данные файла
      if (!filename.includes('.') && originalFilename) {
        // Попробуем извлечь расширение из оригинального имени
        const extMatch = originalFilename.match(/\.([a-zA-Z0-9]+)$/);
        if (extMatch && extMatch[1]) {
          filename += '.' + extMatch[1];
          console.log('Added extension from original name:', extMatch[1]);
        }
      }
      
      // 3. Дефолтное расширение
      if (!filename.includes('.')) {
        filename += '.bin';
        console.log('Added default extension: .bin');
      }
    }
    
    console.log('Final filename for download:', filename);
    
    // Создаем blob с правильным MIME-типом
    const mimeType = response.headers['content-type'] || 'application/octet-stream';
    const blob = new Blob([response.data], { type: mimeType });
    
    // Создаем ссылку для скачивания
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
    
    return { success: true, filename: filename };
  } catch (error) {
    console.error('Download error:', error);
    return { 
      success: false, 
      error: error.response?.data?.error || 'Download failed' 
    };
  }
};

// Улучшенная функция для определения расширения
const getExtensionFromMimeType = (mimeType) => {
  const mimeToExt = {
    // Images
    'image/jpeg': 'jpg',
    'image/jpg': 'jpg',
    'image/png': 'png',
    'image/gif': 'gif',
    'image/webp': 'webp',
    'image/svg+xml': 'svg',
    'image/bmp': 'bmp',
    'image/tiff': 'tiff',
    
    // Documents
    'application/pdf': 'pdf',
    'application/msword': 'doc',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'docx',
    'application/vnd.ms-excel': 'xls',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'xlsx',
    'application/vnd.ms-powerpoint': 'ppt',
    'application/vnd.openxmlformats-officedocument.presentationml.presentation': 'pptx',
    
    // Text
    'text/plain': 'txt',
    'text/html': 'html',
    'text/css': 'css',
    'text/javascript': 'js',
    'application/json': 'json',
    'application/xml': 'xml',
    
    // Archives
    'application/zip': 'zip',
    'application/x-rar-compressed': 'rar',
    'application/x-7z-compressed': '7z',
    'application/gzip': 'gz',
    'application/x-tar': 'tar',
    
    // Audio/Video
    'audio/mpeg': 'mp3',
    'audio/wav': 'wav',
    'video/mp4': 'mp4',
    'video/mpeg': 'mpeg',
    'video/avi': 'avi',
    
    // Other
    'application/octet-stream': 'bin'
  };
  
  return mimeToExt[mimeType?.toLowerCase()];
};


  const deleteFile = async (fileId) => {
    try {
      await storageService.deleteFile(fileId);
      await fetchFiles();
      return { success: true };
    } catch (error) {
      console.error('Delete error:', error);
      return { 
        success: false, 
        error: error.response?.data?.error || 'Delete failed' 
      };
    }
  };

  const getDecryptedFilename = async (fileId, masterPassword) => {
    try {
      const response = await storageService.getDecryptedFilename(fileId, masterPassword);
      return { success: true, filename: response.data.filename };
    } catch (error) {
      console.error('Decrypt filename error:', error);
      return { 
        success: false, 
        error: error.response?.data?.error || 'Failed to get filename' 
      };
    }
  };

  return {
    files,
    loading,
    error,
    refresh,
    uploadFile,
    downloadFile,
    deleteFile,
    getDecryptedFilename
  };
};