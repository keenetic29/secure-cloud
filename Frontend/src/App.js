import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import ProtectedRoute from './components/common/ProtectedRoute';
import AuthLayout from './components/layout/AuthLayout';
import MainLayout from './components/layout/MainLayout';
import Login from './components/auth/Login';
import Register from './components/auth/Register';
import FileList from './components/storage/FileList';
import YandexConnect from './components/storage/YandexConnect';

// App.js
function App() {
  return (
    <AuthProvider>
      <Routes>
        {/* Public routes */}
        <Route path="/auth" element={<AuthLayout />}>
          <Route path="login" element={<Login />} />
          <Route path="register" element={<Register />} />
        </Route>

        {/* Protected routes */}
        <Route path="/" element={
          <ProtectedRoute>
            <MainLayout />
          </ProtectedRoute>
        }>
          <Route index element={<FileList />} />
          <Route path="connect-yandex" element={<YandexConnect />} />
        </Route>
        
        {/* Callback route - должен быть доступен без защиты */}
        <Route path="/api/v1/storage/yandex/callback" element={
          <ProtectedRoute>
            <YandexConnect />
          </ProtectedRoute>
        } />
      </Routes>
    </AuthProvider>
  );
}

export default App;