import { Link } from 'react-router-dom';

export default function NotFoundPage() {
  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="text-center">
          <div className="text-6xl mb-4">🔍</div>
          <h1 className="text-4xl font-bold text-gray-900 mb-2">404</h1>
          <h2 className="text-2xl font-semibold text-gray-700 mb-4">Page Not Found</h2>
          <p className="text-gray-500 mb-8">
            Sorry, the page you're looking for doesn't exist or has been moved.
          </p>
          
          <div className="space-y-4">
            <Link
              to="/"
              className="inline-flex items-center px-4 py-2 border border-transparent text-base font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
            >
              Go Home
            </Link>
            
            <div className="text-sm text-gray-500">
              <Link
                to="/dashboard"
                className="text-primary-600 hover:text-primary-500"
              >
                Go to Dashboard
              </Link>
              {' · '}
              <Link
                to="/login"
                className="text-primary-600 hover:text-primary-500"
              >
                Sign In
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}