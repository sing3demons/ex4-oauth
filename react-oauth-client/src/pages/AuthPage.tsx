import React, { useState } from 'react';
import { LoginForm } from '../components/auth/LoginForm';
import { RegisterForm } from '../components/auth/RegisterForm';

export const AuthPage: React.FC = () => {
  const [isLogin, setIsLogin] = useState(true);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="mx-auto h-12 w-12 bg-blue-600 rounded-full flex items-center justify-center">
            <svg className="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <h1 className="mt-4 text-3xl font-bold text-gray-900">OAuth2 Demo</h1>
          <p className="mt-2 text-gray-600">
            Secure authentication with PKCE flow
          </p>
        </div>

        {/* Toggle Buttons */}
        <div className="bg-white rounded-lg p-1 mb-6 shadow-sm">
          <div className="grid grid-cols-2 gap-1">
            <button
              onClick={() => setIsLogin(true)}
              className={`py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                isLogin
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Sign In
            </button>
            <button
              onClick={() => setIsLogin(false)}
              className={`py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                !isLogin
                  ? 'bg-green-600 text-white shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Sign Up
            </button>
          </div>
        </div>

        {/* Forms */}
        {isLogin ? (
          <LoginForm 
            onSwitchToRegister={() => setIsLogin(false)}
            showOAuthOption={true}
          />
        ) : (
          <RegisterForm 
            onSwitchToLogin={() => setIsLogin(true)}
            onRegisterSuccess={() => {
              setIsLogin(true);
            }}
          />
        )}

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-sm text-gray-500">
            OAuth2 Server with MongoDB • PKCE Flow • JWT Tokens
          </p>
          <div className="mt-2 flex justify-center space-x-4 text-xs text-gray-400">
            <span>✓ Email Verification</span>
            <span>✓ Token Refresh</span>
            <span>✓ Secure Storage</span>
          </div>
        </div>
      </div>
    </div>
  );
};