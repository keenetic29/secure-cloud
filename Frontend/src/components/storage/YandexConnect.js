// YandexConnect.js - обновленная версия
import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { storageService } from '../../services/storage';
import './YandexConnect.css';

const YandexConnect = () => {
  const [authUrl, setAuthUrl] = useState('');
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState(false);
  const [message, setMessage] = useState('');
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  // Проверяем, есть ли код в URL (callback от Яндекс)
  const code = searchParams.get('code');

  useEffect(() => {
    if (code) {
      // Обрабатываем callback
      handleCallback(code);
    } else {
      fetchAuthUrl();
    }
  }, [code]);

  const fetchAuthUrl = async () => {
    try {
      const response = await storageService.getYandexAuthURL();
      setAuthUrl(response.data.auth_url);
    } catch (error) {
      setMessage('Failed to get authorization URL');
      console.error('Error fetching auth URL:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCallback = async (authCode) => {
    setConnecting(true);
    setMessage('Connecting Yandex.Disk...');

    try {
      await storageService.handleYandexCallback(authCode);
      setMessage('Yandex.Disk successfully connected!');
      
      // Редирект через 2 секунды
      setTimeout(() => {
        navigate('/');
      }, 2000);
    } catch (error) {
      setMessage('Failed to connect Yandex.Disk: ' + (error.response?.data?.error || 'Unknown error'));
    } finally {
      setConnecting(false);
    }
  };

  const handleConnect = () => {
    if (!authUrl) return;
    window.location.href = authUrl;
  };

  if (loading) {
    return (
      <div className="yandex-connect">
        <h2>Connect Yandex.Disk</h2>
        <div className="loading">Loading...</div>
      </div>
    );
  }

  // Если это callback страница
  if (code) {
    return (
      <div className="yandex-connect">
        <h2>Connecting Yandex.Disk</h2>
        <div className={`message ${connecting ? 'info' : message.includes('successfully') ? 'success' : 'error'}`}>
          {message}
        </div>
        {message.includes('successfully') && (
          <p>Redirecting to files page...</p>
        )}
      </div>
    );
  }

  return (
    <div className="yandex-connect">
      <h2>Connect Yandex.Disk</h2>
      
      <div className="connection-info">
        <p>
          Connect your Yandex.Disk account to securely store and manage your encrypted files.
        </p>
        <ul>
          <li>Files are encrypted on your device before uploading</li>
          <li>Only you can decrypt your files with your master password</li>
          <li>Your data remains private and secure</li>
        </ul>
      </div>

      {message && (
        <div className={`message ${message.includes('successfully') ? 'success' : 'error'}`}>
          {message}
        </div>
      )}

      <button 
        onClick={handleConnect} 
        disabled={!authUrl}
        className="connect-button"
      >
        Connect Yandex.Disk
      </button>

      <div className="oauth-steps">
        <h3>How it works:</h3>
        <ol>
          <li>Click "Connect Yandex.Disk"</li>
          <li>Authorize the application in the Yandex window</li>
          <li>You will be redirected back to complete the connection</li>
        </ol>
      </div>
    </div>
  );
};

export default YandexConnect;