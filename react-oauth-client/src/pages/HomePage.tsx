import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export default function HomePage() {
  const { isAuthenticated } = useAuth();

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 sm:text-5xl md:text-6xl">
            OAuth2 Client
          </h1>
          <p className="mt-3 max-w-md mx-auto text-base text-gray-500 sm:text-lg md:mt-5 md:text-xl md:max-w-3xl">
            A secure OAuth2 client implementation built with React and TypeScript. 
            Experience modern authentication with PKCE flow support.
          </p>

          <div className="mt-8 flex justify-center space-x-4">
            {isAuthenticated ? (
              <Link
                to="/dashboard"
                className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
              >
                Go to Dashboard
              </Link>
            ) : (
              <Link
                to="/login"
                className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
              >
                Sign In
              </Link>
            )}
          </div>

          <div className="mt-12 grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="px-6 py-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <span className="text-3xl">🔒</span>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900">Secure Authentication</h3>
                    <p className="text-sm text-gray-500">OAuth2 with PKCE for enhanced security</p>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="px-6 py-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <span className="text-3xl">⚡</span>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900">Fast & Modern</h3>
                    <p className="text-sm text-gray-500">Built with React, TypeScript, and Vite</p>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="px-6 py-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <span className="text-3xl">🛠️</span>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900">Custom Implementation</h3>
                    <p className="text-sm text-gray-500">No external OAuth libraries - pure custom code</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}