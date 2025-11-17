import React, { useState } from 'react';
import { cryptoService } from '../../utils/encryption';
import { storageService } from '../../services/storage';
import './FileUpload.css';

const FileUpload = ({ onUploadSuccess, currentPath = '/' }) => {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState('');

  const getMasterPassword = () => {
    return prompt('Enter your master password to encrypt the file:');
  };

  const handleFileSelect = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    const masterPassword = getMasterPassword();
    if (!masterPassword) return;

    setUploading(true);
    setError('');
    setProgress(0);

    try {
      // Шифруем файл на клиенте
      setProgress(30);
      const encryptedFile = await cryptoService.encryptFile(file, masterPassword);
      
      // Шифруем имя файла
      setProgress(60);
      const encryptedFilename = await cryptoService.encryptFilename(file.name, masterPassword);
      
      // Создаем полный путь с учетом текущей папки
      const fullPath = currentPath === '/' ? 
        encryptedFilename : 
        `${currentPath}/${encryptedFilename}`;
      
      // Загружаем напрямую в Яндекс.Диск с отслеживанием прогресса
      setProgress(80);
      await storageService.uploadFile(encryptedFile, fullPath, (uploadProgress) => {
        setProgress(80 + Math.floor(uploadProgress * 0.2)); // 80-100%
      });
      
      setProgress(100);
      
      // Обновляем список файлов
      setTimeout(() => {
        setUploading(false);
        setProgress(0);
        onUploadSuccess();
        alert(`✅ File "${file.name}" successfully encrypted and uploaded to Yandex.Disk!`);
      }, 500);

    } catch (error) {
      setError(error.message);
      setUploading(false);
      setProgress(0);
    }
  };


  const handleDragOver = (event) => {
    event.preventDefault();
    event.currentTarget.classList.add('drag-over');
  };

  const handleDragLeave = (event) => {
    event.preventDefault();
    event.currentTarget.classList.remove('drag-over');
  };

  const handleDrop = (event) => {
    event.preventDefault();
    event.currentTarget.classList.remove('drag-over');
    
    const files = event.dataTransfer.files;
    if (files.length > 0) {
      const input = document.getElementById('file-input');
      input.files = files;
      handleFileSelect(event);
    }
  };

  return (
    <div className="file-upload">
      <input
        type="file"
        id="file-input"
        onChange={handleFileSelect}
        style={{ display: 'none' }}
        disabled={uploading}
      />
      
      <label 
        htmlFor="file-input"
        className={`upload-area ${uploading ? 'uploading' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        {uploading ? (
          <div className="upload-progress">
            <div className="progress-bar">
              <div 
                className="progress-fill" 
                style={{ width: `${progress}%` }}
              ></div>
            </div>
            <span>Encrypting and uploading... {progress}%</span>
          </div>
        ) : (
          <>
            <div className="upload-icon">📁</div>
            <div className="upload-text">
              <strong>Click to upload or drag and drop</strong>
              <small>Files are encrypted before uploading to Yandex.Disk</small>
            </div>
          </>
        )}
      </label>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      <div className="upload-info">
        <p>🔒 Files are encrypted with AES-256 on your device before uploading</p>
        <p>🚀 Uploaded directly to your Yandex.Disk</p>
        <p>🔑 Only you can decrypt files with your master password</p>
      </div>
    </div>
  );
};

export default FileUpload;