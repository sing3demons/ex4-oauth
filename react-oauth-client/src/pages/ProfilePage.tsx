import { useAuth } from '../hooks/useAuth';

export default function ProfilePage() {
  const { user } = useAuth();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Profile</h1>
        <p className="text-gray-600">Manage your account information.</p>
      </div>

      {/* Profile Card */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-5 border-b border-gray-200">
          <div className="flex items-center space-x-4">
            <div className="w-16 h-16 bg-primary-500 rounded-full flex items-center justify-center">
              <span className="text-white font-medium text-xl">
                {user?.first_name?.charAt(0)?.toUpperCase() || user?.username?.charAt(0)?.toUpperCase() || user?.email?.charAt(0)?.toUpperCase() || 'U'}
              </span>
            </div>
            <div>
              <h3 className="text-lg font-medium text-gray-900">
                {user?.first_name && user?.last_name ? `${user.first_name} ${user.last_name}` : user?.username || 'User'}
              </h3>
              <p className="text-sm text-gray-500">{user?.email}</p>
            </div>
          </div>
        </div>

        <div className="px-6 py-5">
          <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-gray-500">User ID</dt>
              <dd className="mt-1 text-sm text-gray-900">{user?.id}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Email</dt>
              <dd className="mt-1 text-sm text-gray-900 flex items-center">
                {user?.email}
                {user?.email_verified && (
                  <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                    Verified
                  </span>
                )}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Username</dt>
              <dd className="mt-1 text-sm text-gray-900">{user?.username}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Role</dt>
              <dd className="mt-1 text-sm text-gray-900">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  user?.role === 'admin' ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'
                }`}>
                  {user?.role || 'user'}
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">First Name</dt>
              <dd className="mt-1 text-sm text-gray-900">{user?.first_name || 'Not provided'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Last Name</dt>
              <dd className="mt-1 text-sm text-gray-900">{user?.last_name || 'Not provided'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Account Status</dt>
              <dd className="mt-1 text-sm text-gray-900">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  user?.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  {user?.is_active ? 'Active' : 'Inactive'}
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Avatar</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {user?.avatar ? (
                  <img 
                    src={user.avatar} 
                    alt="Avatar" 
                    className="w-8 h-8 rounded-full"
                  />
                ) : (
                  'No avatar set'
                )}
              </dd>
            </div>
          </dl>
        </div>
      </div>

      {/* Security Information */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-5 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Security & Authentication</h3>
        </div>
        <div className="px-6 py-5">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-sm font-medium text-gray-900">OAuth2 Authentication</h4>
                <p className="text-sm text-gray-500">Authenticated via OAuth2 Authorization Code Flow with PKCE</p>
              </div>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                Active
              </span>
            </div>
            
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-sm font-medium text-gray-900">Email Verification</h4>
                <p className="text-sm text-gray-500">Your email address verification status</p>
              </div>
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                user?.email_verified ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
              }`}>
                {user?.email_verified ? 'Verified' : 'Pending'}
              </span>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-sm font-medium text-gray-900">Session Security</h4>
                <p className="text-sm text-gray-500">CSRF protection and secure token storage</p>
              </div>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                Protected
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}