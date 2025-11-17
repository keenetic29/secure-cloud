// Утилиты для шифрования с использованием Web Crypto API

export class CryptoService {
  constructor() {
    this.algorithm = { name: 'AES-GCM', length: 256 };
  }

  // Генерация ключа из мастер-пароля
  async deriveKey(password, salt) {
    const encoder = new TextEncoder();
    const keyMaterial = await window.crypto.subtle.importKey(
      'raw',
      encoder.encode(password),
      'PBKDF2',
      false,
      ['deriveKey']
    );

    return window.crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt: salt,
        iterations: 100000,
        hash: 'SHA-256'
      },
      keyMaterial,
      this.algorithm,
      false,
      ['encrypt', 'decrypt']
    );
  }

  // Шифрование файла
  async encryptFile(file, password) {
    try {
      const salt = window.crypto.getRandomValues(new Uint8Array(16));
      const iv = window.crypto.getRandomValues(new Uint8Array(12));
      
      const key = await this.deriveKey(password, salt);
      const fileBuffer = await file.arrayBuffer();
      
      const encryptedContent = await window.crypto.subtle.encrypt(
        { ...this.algorithm, iv },
        key,
        fileBuffer
      );

      // Объединяем salt, iv и зашифрованные данные
      const result = new Uint8Array(salt.length + iv.length + encryptedContent.byteLength);
      result.set(salt, 0);
      result.set(iv, salt.length);
      result.set(new Uint8Array(encryptedContent), salt.length + iv.length);

      return new Blob([result], { type: 'application/octet-stream' });
    } catch (error) {
      console.error('Encryption error:', error);
      throw new Error('File encryption failed');
    }
  }

  // Дешифрование файла
  async decryptFile(encryptedBlob, password) {
    try {
      const encryptedBuffer = await encryptedBlob.arrayBuffer();
      const encryptedArray = new Uint8Array(encryptedBuffer);
      
      const salt = encryptedArray.slice(0, 16);
      const iv = encryptedArray.slice(16, 28);
      const encryptedContent = encryptedArray.slice(28);
      
      const key = await this.deriveKey(password, salt);
      
      const decryptedContent = await window.crypto.subtle.decrypt(
        { ...this.algorithm, iv },
        key,
        encryptedContent
      );

      return new Blob([decryptedContent]);
    } catch (error) {
      console.error('Decryption error:', error);
      throw new Error('File decryption failed - check your master password');
    }
  }

  // Генерация зашифрованного имени файла
  async encryptFilename(filename, password) {
    try {
      const encoder = new TextEncoder();
      const data = encoder.encode(filename);
      
      const salt = window.crypto.getRandomValues(new Uint8Array(16));
      const key = await this.deriveKey(password, salt);
      
      // Используем короткий IV для имен файлов
      const iv = window.crypto.getRandomValues(new Uint8Array(8));
      
      const encrypted = await window.crypto.subtle.encrypt(
        { ...this.algorithm, iv },
        key,
        data
      );

      // Кодируем в base64 для использования в имени файла
      const combined = new Uint8Array(salt.length + iv.length + encrypted.byteLength);
      combined.set(salt, 0);
      combined.set(iv, salt.length);
      combined.set(new Uint8Array(encrypted), salt.length + iv.length);
      
      const base64Name = btoa(String.fromCharCode(...combined))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');
      
      return base64Name + '.encrypted';
    } catch (error) {
      console.error('Filename encryption error:', error);
      throw new Error('Filename encryption failed');
    }
  }

  // Дешифрование имени файла
  async decryptFilename(encryptedFilename, password) {
    try {
      if (!encryptedFilename.endsWith('.encrypted')) {
        return encryptedFilename; // Возвращаем как есть, если не зашифровано
      }

      const base64Name = encryptedFilename.replace('.encrypted', '');
      
      // Декодируем из base64
      const binaryString = atob(base64Name.replace(/-/g, '+').replace(/_/g, '/'));
      const encryptedArray = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        encryptedArray[i] = binaryString.charCodeAt(i);
      }
      
      const salt = encryptedArray.slice(0, 16);
      const iv = encryptedArray.slice(16, 24);
      const encryptedContent = encryptedArray.slice(24);
      
      const key = await this.deriveKey(password, salt);
      
      const decrypted = await window.crypto.subtle.decrypt(
        { ...this.algorithm, iv: new Uint8Array([...iv, 0, 0, 0, 0]) }, // Дополняем IV до 12 байт
        key,
        encryptedContent
      );
      
      const decoder = new TextDecoder();
      return decoder.decode(decrypted);
    } catch (error) {
      console.error('Filename decryption error:', error);
      return encryptedFilename; // Возвращаем исходное имя если дешифрование не удалось
    }
  }
}

export const cryptoService = new CryptoService();