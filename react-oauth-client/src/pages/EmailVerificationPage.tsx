import React, { useEffect, useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { toast } from 'react-hot-toast';
import { authService } from '../services/authService';

export const EmailVerificationPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('');
  const [userInfo, setUserInfo] = useState<any>(null);

  useEffect(() => {
    const verifyEmail = async () => {
      const token = searchParams.get('token');
      
      if (!token) {
        setStatus('error');
        setMessage('Invalid verification link. No token provided.');
        return;
      }

      try {
        const response = await authService.verifyEmail(token);
        setStatus('success');
        setMessage(response.message);
        setUserInfo(response.user);
        toast.success('Email verified successfully!');
      } catch (error) {
        setStatus('error');
        setMessage(error instanceof Error ? error.message : 'Email verification failed');
        toast.error('Email verification failed');
      }
    };

    verifyEmail();
  }, [searchParams]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <div className="bg-white rounded-lg shadow-md p-8 text-center">
          {status === 'loading' && (
            <div>
              <div className="mx-auto h-12 w-12 mb-4">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
              </div>
              <h2 className="text-xl font-bold text-gray-900 mb-2">Verifying Email</h2>
              <p className="text-gray-600">Please wait while we verify your email address...</p>
            </div>
          )}

          {status === 'success' && (
            <div>
              <div className="mx-auto h-12 w-12 bg-green-100 rounded-full flex items-center justify-center mb-4">
                <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h2 className="text-xl font-bold text-green-900 mb-2">Email Verified!</h2>
              <p className="text-gray-600 mb-4">{message}</p>
              
              {userInfo && (
                <div className="bg-green-50 border border-green-200 rounded-md p-4 mb-4">
                  <h3 className="text-sm font-medium text-green-800 mb-2">Account Information</h3>
                  <div className="text-sm text-green-700 space-y-1">
                    <p><strong>Email:</strong> {userInfo.email}</p>
                    <p><strong>Username:</strong> {userInfo.username}</p>
                    <p><strong>Status:</strong> {userInfo.email_verified ? 'Verified ✓' : 'Pending'}</p>
                  </div>
                </div>
              )}

              <Link
                to="/login"
                className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                Continue to Login
              </Link>
            </div>
          )}

          {status === 'error' && (
            <div>
              <div className="mx-auto h-12 w-12 bg-red-100 rounded-full flex items-center justify-center mb-4">
                <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </div>
              <h2 className="text-xl font-bold text-red-900 mb-2">Verification Failed</h2>
              <p className="text-gray-600 mb-6">{message}</p>
              
              <div className="space-y-3">
                <Link
                  to="/login"
                  className="block w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                >
                  Back to Login
                </Link>
                <Link
                  to="/register"
                  className="block w-full px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition-colors"
                >
                  Create New Account
                </Link>
              </div>
            </div>
          )}
        </div>

        {/* Info Box */}
        <div className="mt-6 bg-blue-50 border border-blue-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">
                Need Help?
              </h3>
              <div className="mt-2 text-sm text-blue-700">
                <p>If you're having trouble with email verification:</p>
                <ul className="mt-1 list-disc list-inside space-y-1">
                  <li>Check your spam/junk folder</li>
                  <li>Make sure the link hasn't expired</li>
                  <li>Try registering again if needed</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};