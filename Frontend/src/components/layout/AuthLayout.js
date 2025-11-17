import React from 'react';
import { Outlet } from 'react-router-dom';
import './AuthLayout.css';

const AuthLayout = () => {
  return (
    <div className="auth-layout">
      <div className="auth-container">
        <div className="auth-card">
          <h1 className="auth-title">Secure Cloud Storage</h1>
          <p className="auth-subtitle">Protect your files with client-side encryption</p>
          <Outlet />
        </div>
      </div>
    </div>
  );
};

export default AuthLayout;