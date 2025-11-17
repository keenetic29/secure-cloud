import { useState, useEffect } from 'react';
import { storageService } from '../services/storage';

export const useStorage = (path = '') => {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchFiles();
  }, [path]);

  const fetchFiles = async () => {
    try {
      setLoading(true);
      const response = await storageService.getFiles(path);
      setFiles(response.data.files || []);
      setError('');
    } catch (error) {
      setError(error.response?.data?.error || 'Failed to load files');
    } finally {
      setLoading(false);
    }
  };

  const refreshFiles = () => {
    fetchFiles();
  };

  return {
    files,
    loading,
    error,
    refreshFiles
  };
};