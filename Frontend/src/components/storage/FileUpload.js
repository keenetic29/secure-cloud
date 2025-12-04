import React, { useState } from 'react';
import { storageService } from '../../services/storage';
import './FileUpload.css';

const FileUpload = ({ onUploadSuccess, currentPath = '/' }) => {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState('');

  const handleFileSelect = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    const masterPassword = prompt('Enter your master password to encrypt the file:');
    if (!masterPassword) return;

    // Сохраняем пароль для подсказок
    localStorage.setItem('lastMasterPassword', masterPassword);

    setUploading(true);
    setError('');
    setProgress(0);

    try {
      // Симулируем прогресс для UX
      setProgress(30);
      
      const response = await storageService.uploadFile(file, masterPassword, currentPath);
      
      setProgress(100);
      
      setTimeout(() => {
        setUploading(false);
        setProgress(0);
        onUploadSuccess();
        alert(`✅ File "${file.name}" encrypted and uploaded successfully!\nEncrypted as: ${response.data.encrypted_name}`);
      }, 500);

    } catch (error) {
      setError(error.response?.data?.error || 'Upload failed');
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
            <span>Encrypting and uploading... {Math.round(progress)}%</span>
          </div>
        ) : (
          <>
            <div className="upload-icon">📁</div>
            <div className="upload-text">
              <strong>Click to upload or drag and drop</strong>
              <small>Files are encrypted server-side before storage</small>
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
        <p>🔒 Files are encrypted with AES-256 on the server</p>
        <p>📝 Original filenames are stored securely in database</p>
        <p>🔑 Only you can decrypt files with your master password</p>
        <p>💾 Uploading to: {currentPath === '/' ? 'Root folder' : currentPath}</p>
      </div>
    </div>
  );
};

export default FileUpload;