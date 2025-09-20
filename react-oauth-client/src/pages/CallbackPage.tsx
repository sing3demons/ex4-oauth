import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { authService } from '../services/authService';
import { toast } from 'react-hot-toast';

export default function CallbackPage() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // Check for error in URL params
        const errorParam = searchParams.get('error');
        if (errorParam) {
          const errorDescription = searchParams.get('error_description');
          throw new Error(`OAuth error: ${errorParam}${errorDescription ? ` - ${errorDescription}` : ''}`);
        }

        // Handle successful callback
        await authService.handleCallback();
        
        toast.success('Login successful!');
        navigate('/dashboard', { replace: true });
      } catch (error) {
        console.error('Callback handling failed:', error);
        const errorMessage = error instanceof Error ? error.message : 'Authentication failed';
        setError(errorMessage);
        toast.error(errorMessage);
        
        // Redirect to login after a delay
        setTimeout(() => {
          navigate('/login', { replace: true });
        }, 3000);
      } finally {
        setLoading(false);
      }
    };

    handleCallback();
  }, [navigate, searchParams]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-md w-full text-center">
          <div className="bg-white rounded-lg shadow-md p-8">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
            <h2 className="text-lg font-medium text-gray-900 mb-2">
              Processing Authentication...
            </h2>
            <p className="text-sm text-gray-500">
              Please wait while we complete your login.
            </p>
            <div className="mt-4 text-xs text-gray-400">
              <p>• Verifying authorization code</p>
              <p>• Exchanging for access tokens</p>
              <p>• Setting up your session</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-md w-full text-center">
          <div className="bg-white rounded-lg shadow-md p-8">
            <div className="w-16 h-16 mx-auto bg-red-100 rounded-full flex items-center justify-center mb-4">
              <span className="text-red-600 text-2xl">❌</span>
            </div>
            <h2 className="text-lg font-medium text-gray-900 mb-2">
              Authentication Failed
            </h2>
            <p className="text-sm text-red-600 mb-4">
              {error}
            </p>
            <p className="text-xs text-gray-500 mb-4">
              You will be redirected to the login page shortly.
            </p>
            <button
              onClick={() => navigate('/login', { replace: true })}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
            >
              Try Again
            </button>
          </div>
        </div>
      </div>
    );
  }

  return null;
}