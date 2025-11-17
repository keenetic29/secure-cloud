import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import './Header.css';

const Header = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/auth/login');
  };

  return (
    <header className="header">
      <div className="header-content">
        <Link to="/" className="logo">
          Secure Cloud Storage
        </Link>
        
        <nav className="nav">
          <Link to="/" className="nav-link">
            Files
          </Link>
          <Link to="/connect-yandex" className="nav-link">
            Connect Yandex.Disk
          </Link>
        </nav>

        <div className="user-section">
          <span className="user-email">{user?.email}</span>
          <button onClick={handleLogout} className="logout-btn">
            Logout
          </button>
        </div>
      </div>
    </header>
  );
};

export default Header;