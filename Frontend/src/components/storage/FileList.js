import React, { useState, useEffect } from 'react';
import { useStorage } from '../../hooks/useStorage';
import FileUpload from './FileUpload';
import './FileList.css';

const FileList = () => {
  const [currentPath, setCurrentPath] = useState('/');
  const [hoveredFile, setHoveredFile] = useState(null);
  const [originalName, setOriginalName] = useState(null);
  const [navigationHistory, setNavigationHistory] = useState(['/']);
  
  const {
    files,
    loading,
    error,
    refresh,
    downloadFile,
    deleteFile,
    getDecryptedFilename
  } = useStorage(currentPath);

  const handleFolderClick = (folder) => {
    console.log('Folder clicked (full object):', folder);
    
    // Если передана строка вместо объекта, преобразуем
    let folderObj = folder;
    if (typeof folder === 'string') {
      // Находим объект папки по пути
      folderObj = files.find(f => f.path === folder || f.original_path === folder);
      if (!folderObj) {
        // Пытаемся создать минимальный объект из строки
        const pathParts = folder.split('/').filter(part => part !== '' && part !== 'disk:');
        const folderName = pathParts[pathParts.length - 1] || 'folder';
        folderObj = {
          path: folder,
          filename: folderName,
          type: 'dir'
        };
      }
    }
    
    // Используем нормализованные поля
    const folderPath = folderObj?.path || folderObj?.original_path || folder;
    
    if (!folderPath) {
      console.error('No path found for folder:', folderObj);
      return;
    }
    
    console.log('Folder path for navigation:', folderPath);
    
    // Извлекаем имя папки для навигации
    const pathParts = folderPath.split('/').filter(part => part !== '' && part !== 'disk:');
    const folderName = pathParts[pathParts.length - 1] || folderObj.filename || 'folder';
    
    console.log('Folder name extracted:', folderName);
    
    // Формируем новый путь для API запроса
    let newPath;
    if (currentPath === '/') {
      newPath = '/' + folderName;
    } else {
      // Убедимся, что текущий путь не заканчивается слэшом
      const cleanCurrent = currentPath.endsWith('/') 
        ? currentPath.slice(0, -1) 
        : currentPath;
      newPath = cleanCurrent + '/' + folderName;
    }
    
    console.log(`Navigating from "${currentPath}" to "${newPath}"`);
    setNavigationHistory(prev => [...prev, newPath]);
    setCurrentPath(newPath);
  };

  const handleBackClick = () => {
    if (navigationHistory.length > 1) {
      const newHistory = navigationHistory.slice(0, -1);
      const previousPath = newHistory[newHistory.length - 1];
      console.log('Going back from', currentPath, 'to', previousPath);
      setNavigationHistory(newHistory);
      setCurrentPath(previousPath);
    }
  };

  const getBreadcrumbs = () => {
    if (currentPath === '/') return ['Root'];
    
    try {
      const cleanPath = currentPath.replace(/^\/+|\/+$/g, '');
      const parts = cleanPath.split('/').filter(part => part !== '');
      return ['Root', ...parts];
    } catch (error) {
      console.error('Error parsing breadcrumbs:', error);
      return ['Root'];
    }
  };

  const handleBreadcrumbClick = (index) => {
    const breadcrumbs = getBreadcrumbs();
    
    if (index === 0) {
      setNavigationHistory(['/']);
      setCurrentPath('/');
    } else {
      const pathParts = breadcrumbs.slice(1, index + 1);
      const newPath = '/' + pathParts.join('/');
      console.log('Breadcrumb navigation to:', newPath);
      setNavigationHistory(prev => [...prev.slice(0, -1), newPath]);
      setCurrentPath(newPath);
    }
  };

  const handleDownload = async (file) => {
    if (file.type === 'dir') return;
    
    const masterPassword = prompt('Enter your master password to download the file:');
    if (!masterPassword) return;

    const result = await downloadFile(file.id, masterPassword);
    if (!result.success) {
      alert(`❌ Download failed: ${result.error}`);
    }
  };

  const handleDelete = async (file) => {
    if (!window.confirm(`Are you sure you want to delete "${file.filename}"?`)) {
      return;
    }

    const result = await deleteFile(file.id);
    if (result.success) {
      alert('✅ File deleted successfully!');
      refresh();
    } else {
      alert(`❌ Delete failed: ${result.error}`);
    }
  };

  const handleMouseEnter = async (file) => {
    setHoveredFile(file);
    
    // Если файл зашифрован, пытаемся получить оригинальное имя
    if (file.is_encrypted && file.type === 'file') {
      const masterPassword = localStorage.getItem('lastMasterPassword');
      if (masterPassword) {
        const result = await getDecryptedFilename(file.id, masterPassword);
        if (result.success) {
          setOriginalName(result.filename);
        } else {
          // Если не получилось расшифровать, показываем что можем
          setOriginalName('encrypted_file');
        }
      }
    }
  };

  const formatFileSize = (bytes) => {
    if (!bytes || bytes === 0) {
      if (bytes === 0) return '0 Bytes';
      return 'N/A';
    }
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString) => {
    if (!dateString) return '';
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch (e) {
      return '';
    }
  };

  const getFileTooltip = (file) => {
    if (file.type === 'dir') {
      return `Folder: ${file.filename}\nPath: ${file.path}\nClick to open`;
    }
    
    if (file.is_encrypted && originalName && hoveredFile?.id === file.id) {
      return `Original: ${originalName}\nEncrypted: ${file.encrypted_name || file.filename}\nSize: ${formatFileSize(file.size)}\nType: ${file.mime_type || 'unknown'}`;
    }
    
    return `${file.filename}\nSize: ${formatFileSize(file.size)}\nType: ${file.mime_type || 'unknown'}\nPath: ${file.path}`;
  };

  const getDisplayName = (file) => {
    if (!file) return 'Unknown';
    
    // Используем нормализованные поля
    const filename = file.filename || '';
    const encryptedName = file.encrypted_name || '';
    const path = file.path || file.original_path || '';
    
    // Для файлов с именем "unknown" или "encrypted_file" показываем зашифрованное имя или путь
    if (!filename || filename === 'unknown' || filename === 'encrypted_file') {
      return encryptedName || 
             path.split('/').pop() || 
             'file';
    }
    return filename;
  };

  if (loading) {
    return (
      <div className="file-list">
        <div className="loading">
          <div className="loading-spinner"></div>
          <p>Loading files from Yandex.Disk...</p>
          <p>Path: {currentPath}</p>
        </div>
      </div>
    );
  }

  const breadcrumbs = getBreadcrumbs();
  
  // Разделяем файлы и папки для лучшего отображения
  const folders = files.filter(f => f.type === 'dir');
  const fileItems = files.filter(f => f.type === 'file');

  return (
    <div className="file-list">
      <div className="file-list-header">
        <div className="path-navigation">
          <button 
            onClick={handleBackClick} 
            disabled={navigationHistory.length <= 1}
            className="back-button"
            title="Go back"
          >
            ← Back
          </button>
          
          <div className="breadcrumbs">
            {breadcrumbs.map((crumb, index) => (
              <span key={index} className="breadcrumb">
                {index > 0 && <span className="breadcrumb-separator"> / </span>}
                <button
                  onClick={() => handleBreadcrumbClick(index)}
                  className={`breadcrumb-link ${index === breadcrumbs.length - 1 ? 'current' : ''}`}
                  disabled={index === breadcrumbs.length - 1}
                  title={`Go to ${crumb}`}
                >
                  {crumb}
                </button>
              </span>
            ))}
          </div>
          
          <span className="current-path" title="Current directory">
            📁 {currentPath === '/' ? 'Root' : currentPath}
          </span>
        </div>
        
        <div className="header-main">
          <h2>Secure Cloud Storage</h2>
          <div className="stats">
            <span className="stat-item">
              📁 {folders.length} folder{folders.length !== 1 ? 's' : ''}
            </span>
            <span className="stat-item">
              📄 {fileItems.length} file{fileItems.length !== 1 ? 's' : ''}
            </span>
          </div>
        </div>
        
        <FileUpload 
          onUploadSuccess={refresh} 
          currentPath={currentPath}
        />
      </div>

      {error && (
        <div className="error-message">
          ⚠️ {error}
        </div>
      )}

      {files.length === 0 ? (
        <div className="empty-state">
          <h3>📭 Empty folder</h3>
          <p>No files or folders found in {currentPath === '/' ? 'root' : `"${currentPath}"`}</p>
          <p>Upload files or connect Yandex.Disk to see content.</p>
        </div>
      ) : (
        <div className="files-container">
          {/* Папки */}
          {folders.length > 0 && (
            <>
              <h3 className="section-title">📁 Folders ({folders.length})</h3>
              <div className="file-grid">
                {folders.map((folder) => (
                  <div 
                    key={folder.id || folder.path} 
                    className="file-item folder-item"
                    onMouseEnter={() => handleMouseEnter(folder)}
                    onMouseLeave={() => {
                      setHoveredFile(null);
                      setOriginalName(null);
                    }}
                    title={getFileTooltip(folder)}
                    onClick={() => handleFolderClick(folder)} // Передаем объект папки
                  >
                    <div className="file-icon">
                      <span className="folder-icon-large">📁</span>
                    </div>
                    
                    <div className="file-info">
                      <h4 className="file-name">
                        {getDisplayName(folder)}
                        <span className="folder-indicator"> (folder)</span>
                      </h4>
                      <div className="file-meta">
                        <span className="file-type">Directory</span>
                        {folder.updated_at && (
                          <span className="file-date">{formatDate(folder.updated_at)}</span>
                        )}
                      </div>
                    </div>
                    
                    <div className="file-actions">
                      <button 
                        className="action-btn open-btn"
                        title="Open folder"
                        onClick={(e) => {
                          e.stopPropagation(); // Останавливаем всплытие
                          handleFolderClick(folder); // Передаем объект папки
                        }}
                      >
                        🔍
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}

          {/* Файлы */}
          {fileItems.length > 0 && (
            <>
              <h3 className="section-title">📄 Files ({fileItems.length})</h3>
              <div className="file-grid">
                {fileItems.map((file) => (
                  <div 
                    key={file.id} 
                    className="file-item"
                    onMouseEnter={() => handleMouseEnter(file)}
                    onMouseLeave={() => {
                      setHoveredFile(null);
                      setOriginalName(null);
                    }}
                    title={getFileTooltip(file)}
                  >
                    <div className="file-icon">
                      {file.mime_type?.startsWith('image/') ? (
                        <span>🖼️</span>
                      ) : file.mime_type?.includes('pdf') ? (
                        <span>📕</span>
                      ) : file.mime_type?.includes('text') ? (
                        <span>📝</span>
                      ) : (
                        <span>📄</span>
                      )}
                      {file.is_encrypted && <span className="encrypted-badge">🔒</span>}
                    </div>
                    
                    <div className="file-info">
                      <h4 className="file-name">
                        {file.is_encrypted && hoveredFile?.id === file.id && originalName ? (
                          <>
                            <span className="original-name">{originalName}</span>
                            <span className="encrypted-hint"> (encrypted)</span>
                          </>
                        ) : (
                          <>
                            {getDisplayName(file)}
                            {file.is_encrypted && <span className="encrypted-indicator"> (encrypted)</span>}
                          </>
                        )}
                      </h4>
                      <div className="file-meta">
                        <span className="file-size">{formatFileSize(file.size)}</span>
                        {file.mime_type && file.mime_type !== 'application/octet-stream' && (
                          <span className="file-type">{file.mime_type.split('/')[1] || file.mime_type}</span>
                        )}
                        {file.updated_at && (
                          <span className="file-date">{formatDate(file.updated_at)}</span>
                        )}
                      </div>
                    </div>

                    <div className="file-actions">
                      <button 
                        onClick={() => handleDownload(file)}
                        className="action-btn download-btn"
                        title="Download file"
                        disabled={file.type === 'dir'}
                      >
                        ⬇️
                      </button>
                      <button 
                        onClick={() => handleDelete(file)}
                        className="action-btn delete-btn"
                        title="Delete file"
                      >
                        🗑️
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default FileList;