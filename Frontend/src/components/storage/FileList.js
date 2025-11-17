import React, { useState, useEffect } from 'react';
import { storageService } from '../../services/storage';
import { cryptoService } from '../../utils/encryption';
import FileUpload from './FileUpload';
import './FileList.css';

const FileList = () => {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [currentPath, setCurrentPath] = useState('/');
  const [navigationHistory, setNavigationHistory] = useState(['/']);

  useEffect(() => {
    fetchFiles();
  }, [currentPath]);

  const fetchFiles = async () => {
    try {
      setLoading(true);
      const response = await storageService.getFiles(currentPath);
      setFiles(response.data.files || []);
      setError('');
    } catch (error) {
      setError('Failed to load files: ' + error.message);
      console.error('Error fetching files from path:', currentPath);
      console.error('Error details:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleFileUpload = () => {
    fetchFiles();
  };

  const handleFolderClick = (folderPath) => {
  console.log('Original folder path:', folderPath);
  
  // Очищаем путь от disk:
  let cleanPath = folderPath.replace(/^\/?disk:/, '');
  if (!cleanPath.startsWith('/')) {
    cleanPath = '/' + cleanPath;
  }
  
  console.log('Cleaned folder path:', cleanPath);
  setNavigationHistory(prev => [...prev, cleanPath]);
  setCurrentPath(cleanPath);
};

  const handleBackClick = () => {
    if (navigationHistory.length > 1) {
      const newHistory = navigationHistory.slice(0, -1);
      const previousPath = newHistory[newHistory.length - 1];
      
      console.log('Navigating back from:', currentPath, 'to:', previousPath);
      
      setNavigationHistory(newHistory);
      setCurrentPath(previousPath);
    }
  };

  const getBreadcrumbs = () => {
  if (currentPath === '/') return ['Root'];
  
  try {
    // Очищаем путь от disk: для breadcrumbs
    const cleanPath = currentPath.replace(/^\/?disk:/, '');
    const parts = cleanPath.split('/').filter(part => part !== '');
    return ['Root', ...parts];
  } catch (error) {
    return ['Root'];
  }
};

  const handleBreadcrumbClick = (index) => {
    const breadcrumbs = getBreadcrumbs();
    
    if (index === 0) {
      // Корень
      setNavigationHistory(['/']);
      setCurrentPath('/');
    } else {
      // Строим путь из breadcrumbs
      const pathParts = breadcrumbs.slice(1, index + 1);
      const newPath = '/' + pathParts.join('/');
      
      console.log('Breadcrumb navigation to:', newPath);
      setNavigationHistory(prev => [...prev.slice(0, -1), newPath]);
      setCurrentPath(newPath);
    }
  };

  const breadcrumbs = getBreadcrumbs();

  // УБИРАЕМ ФИЛЬТРАЦИЮ - показываем ВСЕ файлы и папки
  // const filteredFiles = files; // Просто используем files как есть

  const getMasterPassword = () => {
    return prompt('Enter your master password to decrypt the file:');
  };

  const handleDownload = async (file) => {
    try {
      const masterPassword = getMasterPassword();
      if (!masterPassword) return;

      const encryptedBlob = await storageService.downloadFile(file.path);
      const decryptedBlob = await cryptoService.decryptFile(encryptedBlob, masterPassword);
      
      let filename = file.filename;
      try {
        filename = await cryptoService.decryptFilename(file.filename, masterPassword);
      } catch (e) {
        console.warn('Could not decrypt filename, using original');
      }
      
      const url = URL.createObjectURL(decryptedBlob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      
      alert(`✅ File "${filename}" downloaded and decrypted successfully!`);
      
    } catch (error) {
      alert(`❌ Download failed: ${error.message}`);
    }
  };

  const handleDelete = async (file) => {
    if (!window.confirm(`Are you sure you want to delete "${file.filename}"?`)) {
      return;
    }

    try {
      await storageService.deleteFile(file.path);
      setFiles(files.filter(f => f.id !== file.id));
      alert('✅ File deleted successfully!');
    } catch (error) {
      alert(`❌ Delete failed: ${error.message}`);
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (loading) {
    return (
      <div className="file-list">
        <div className="loading">Loading files from Yandex.Disk...</div>
      </div>
    );
  }

  return (
    <div className="file-list">
      <div className="file-list-header">
        <div className="path-navigation">
          <button 
            onClick={handleBackClick} 
            disabled={navigationHistory.length <= 1}
            className="back-button"
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
                >
                  {crumb}
                </button>
              </span>
            ))}
          </div>
        </div>
        
        <h2>Your Yandex.Disk Files</h2>
        <FileUpload onUploadSuccess={handleFileUpload} currentPath={currentPath} />
      </div>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      {files.length === 0 ? (
        <div className="empty-state">
          <h3>No files found in {currentPath === '/' ? 'Yandex.Disk' : `"${currentPath}"`}</h3>
          <p>Upload your first encrypted file to get started.</p>
        </div>
      ) : (
        <div className="files-container">
          <div className="file-grid">
            {files.map((file) => (
              <div key={file.id} className="file-item">
                <div className="file-icon">
                  {file.mimeType && file.mimeType.startsWith('image/') ? (
                    <span>🖼️</span>
                  ) : file.type === 'dir' ? (
                    <span 
                      className="folder-icon"
                      onClick={() => handleFolderClick(file.path)}
                      style={{cursor: 'pointer'}}
                      title="Click to open folder"
                    >
                      📁
                    </span>
                  ) : (
                    <span>📄</span>
                  )}
                  {file.isEncrypted && <span className="encrypted-badge">🔒</span>}
                </div>
                
                <div className="file-info">
                  <h4 
                    className="file-name" 
                    title={file.filename}
                    style={file.type === 'dir' ? {cursor: 'pointer', color: '#007bff'} : {}}
                    onClick={file.type === 'dir' ? () => handleFolderClick(file.path) : undefined}
                  >
                    {file.filename}
                    {file.isEncrypted && <span className="encrypted-indicator"> (encrypted)</span>}
                  </h4>
                  <div className="file-meta">
                    {file.size && (
                      <span className="file-size">{formatFileSize(file.size)}</span>
                    )}
                    {file.modified && (
                      <span className="file-date">{formatDate(file.modified)}</span>
                    )}
                  </div>
                </div>

                <div className="file-actions">
                  {file.type === 'file' && (
                    <>
                      <button 
                        onClick={() => handleDownload(file)}
                        className="action-btn download-btn"
                        title="Download and decrypt file"
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
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default FileList;