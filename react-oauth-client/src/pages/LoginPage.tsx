import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { LoginForm } from '../components/auth/LoginForm';
import { RegisterForm } from '../components/auth/RegisterForm';

export default function LoginPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const from = location.state?.from?.pathname || '/dashboard';

  useEffect(() => {
    if (isAuthenticated) {
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, from]);

  const handleRegisterSuccess = () => {
    setMode('login');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h1 className="text-3xl font-extrabold text-gray-900">
            {mode === 'login' ? 'Welcome Back' : 'Create Account'}
          </h1>
          <p className="mt-2 text-sm text-gray-600">
            {mode === 'login' 
              ? 'Sign in to your account or authenticate with OAuth2' 
              : 'Join us today and get started'}
          </p>
        </div>

        {mode === 'login' ? (
          <LoginForm 
            onSwitchToRegister={() => setMode('register')}
            showOAuthOption={true}
          />
        ) : (
          <RegisterForm 
            onSwitchToLogin={() => setMode('login')}
            onRegisterSuccess={handleRegisterSuccess}
          />
        )}

        <div className="text-center">
          <div className="bg-white rounded-lg shadow-sm p-4 text-xs text-gray-500 space-y-2">
            <p><strong>Authentication Options:</strong></p>
            <ul className="list-disc list-inside space-y-1">
              <li><strong>Local Login:</strong> Email/password authentication with JWT tokens</li>
              <li><strong>OAuth2 PKCE:</strong> Secure authorization code flow with PKCE</li>
              <li><strong>Registration:</strong> Create new account with email verification</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}