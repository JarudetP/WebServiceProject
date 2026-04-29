import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { Dashboard } from './pages/Dashboard';
import { Games } from './pages/Games';
import { GameDetails } from './pages/GameDetails';
import { Profile } from './pages/Profile';
import { Analytics } from './pages/Analytics';
import { getAuthToken } from './services/api';

const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = getAuthToken();
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
};

const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = getAuthToken();
  if (token) {
    return <Navigate to="/dashboard" replace />;
  }
  return <>{children}</>;
};

function App() {
  return (
    <Router>
      <Toaster 
        position="top-right"
        toastOptions={{
          className: 'text-sm font-medium border border-border shadow-sm rounded-xl',
          duration: 4000,
          style: {
            background: '#ffffff',
            color: '#111827',
          },
        }}
      />
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        
        <Route path="/login" element={
          <PublicRoute>
            <Login />
          </PublicRoute>
        } />
        
        <Route path="/register" element={
          <PublicRoute>
            <Register />
          </PublicRoute>
        } />
        
        <Route path="/dashboard" element={
          <PrivateRoute>
            <Dashboard />
          </PrivateRoute>
        } />
        
        <Route path="/games" element={
          <PrivateRoute>
            <Games />
          </PrivateRoute>
        } />
        
        <Route path="/games/:id" element={
          <PrivateRoute>
            <GameDetails />
          </PrivateRoute>
        } />
        
        <Route path="/profile" element={
          <PrivateRoute>
            <Profile />
          </PrivateRoute>
        } />

        <Route path="/analytics" element={
          <PrivateRoute>
            <Analytics />
          </PrivateRoute>
        } />
        
        {/* Placeholder for settings */}
        <Route path="/settings" element={
          <PrivateRoute>
            <div className="flex items-center justify-center min-h-screen bg-background">
              <p className="text-accent">Settings coming soon...</p>
            </div>
          </PrivateRoute>
        } />
      </Routes>
    </Router>
  );
}

export default App;
