import React from 'react';
import { useAuth } from '../hooks/useAuth';
import { TokenInfo } from '../components/dashboard/TokenInfo';
import { toast } from 'react-hot-toast';

export default function DashboardPage() {
  const { user, logout, refreshUser } = useAuth();

  const handleLogout = async () => {
    if (window.confirm('Are you sure you want to logout?')) {
      await logout();
    }
  };

  const handleRefreshUser = async () => {
    try {
      await refreshUser();
      toast.success('User information refreshed');
    } catch (error) {
      toast.error('Failed to refresh user information');
    }
  };

  const stats = [
    { 
      name: 'Authentication Status', 
      value: 'Authenticated', 
      icon: '🔐',
      color: 'bg-green-100 text-green-800'
    },
    { 
      name: 'Account Status', 
      value: user?.is_active ? 'Active' : 'Inactive', 
      icon: user?.is_active ? '✅' : '❌',
      color: user?.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
    },
    { 
      name: 'Email Verified', 
      value: user?.email_verified ? 'Yes' : 'No', 
      icon: user?.email_verified ? '📧' : '📭',
      color: user?.email_verified ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
    },
    { 
      name: 'User Role', 
      value: user?.role || 'User', 
      icon: user?.role === 'admin' ? '�' : '�👤',
      color: user?.role === 'admin' ? 'bg-purple-100 text-purple-800' : 'bg-blue-100 text-blue-800'
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
            <p className="text-gray-600 mt-1">
              Welcome back, {user?.first_name && user?.last_name 
                ? `${user.first_name} ${user.last_name}` 
                : user?.username || user?.email}!
            </p>
          </div>
          <div className="flex space-x-3">
            <button
              onClick={handleRefreshUser}
              className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
            >
              🔄 Refresh
            </button>
            <button
              onClick={handleLogout}
              className="bg-red-600 text-white px-4 py-2 rounded-md hover:bg-red-700 transition-colors"
            >
              🚪 Logout
            </button>
          </div>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <div key={stat.name} className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <span className="text-2xl">{stat.icon}</span>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">
                      {stat.name}
                    </dt>
                    <dd className="text-lg font-medium">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-sm font-medium ${stat.color}`}>
                        {stat.value}
                      </span>
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* User Information Card */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-5 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">User Information</h3>
          </div>
          <div className="px-6 py-5">
            <dl className="grid grid-cols-1 gap-x-4 gap-y-6">
              <div>
                <dt className="text-sm font-medium text-gray-500">Email</dt>
                <dd className="mt-1 text-sm text-gray-900 flex items-center">
                  {user?.email}
                  {user?.email_verified && (
                    <span className="ml-2 inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                      ✓ Verified
                    </span>
                  )}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Username</dt>
                <dd className="mt-1 text-sm text-gray-900">{user?.username || 'Not set'}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Full Name</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {user?.first_name && user?.last_name 
                    ? `${user.first_name} ${user.last_name}` 
                    : 'Not provided'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">User ID</dt>
                <dd className="mt-1 text-sm text-gray-900">{user?.id}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Role</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    user?.role === 'admin' ? 'bg-red-100 text-red-800' : 
                    user?.role === 'moderator' ? 'bg-yellow-100 text-yellow-800' :
                    'bg-green-100 text-green-800'
                  }`}>
                    {user?.role || 'user'}
                  </span>
                </dd>
              </div>
              {user?.created_at && (
                <div>
                  <dt className="text-sm font-medium text-gray-500">Member Since</dt>
                  <dd className="mt-1 text-sm text-gray-900">
                    {new Date(user.created_at).toLocaleDateString()}
                  </dd>
                </div>
              )}
            </dl>
          </div>
        </div>

        {/* OAuth2 Features */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-5 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Authentication Features</h3>
          </div>
          <div className="px-6 py-5">
            <div className="space-y-4">
              <div className="flex items-center">
                <span className="text-green-500 text-lg mr-3">✅</span>
                <div>
                  <div className="text-sm font-medium text-gray-900">OAuth2 PKCE</div>
                  <div className="text-xs text-gray-500">Proof Key for Code Exchange security</div>
                </div>
              </div>
              <div className="flex items-center">
                <span className="text-green-500 text-lg mr-3">✅</span>
                <div>
                  <div className="text-sm font-medium text-gray-900">CSRF Protection</div>
                  <div className="text-xs text-gray-500">State parameter validation</div>
                </div>
              </div>
              <div className="flex items-center">
                <span className="text-green-500 text-lg mr-3">✅</span>
                <div>
                  <div className="text-sm font-medium text-gray-900">Secure Token Storage</div>
                  <div className="text-xs text-gray-500">Safe client-side token management</div>
                </div>
              </div>
              <div className="flex items-center">
                <span className="text-green-500 text-lg mr-3">✅</span>
                <div>
                  <div className="text-sm font-medium text-gray-900">OpenID Connect</div>
                  <div className="text-xs text-gray-500">Identity layer with ID tokens</div>
                </div>
              </div>
              <div className="flex items-center">
                <span className="text-green-500 text-lg mr-3">✅</span>
                <div>
                  <div className="text-sm font-medium text-gray-900">Auto Token Refresh</div>
                  <div className="text-xs text-gray-500">Seamless session management</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Token Information */}
      <TokenInfo />

      {/* Quick Actions */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-5 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Quick Actions</h3>
        </div>
        <div className="px-6 py-5">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <button className="bg-gray-50 hover:bg-gray-100 p-4 rounded-lg text-left transition-colors border border-gray-200">
              <div className="text-2xl mb-2">📊</div>
              <h4 className="font-medium text-gray-900">View Profile</h4>
              <p className="text-sm text-gray-600 mt-1">Manage account details</p>
            </button>
            <button className="bg-gray-50 hover:bg-gray-100 p-4 rounded-lg text-left transition-colors border border-gray-200">
              <div className="text-2xl mb-2">🔑</div>
              <h4 className="font-medium text-gray-900">Change Password</h4>
              <p className="text-sm text-gray-600 mt-1">Update your password</p>
            </button>
            <button className="bg-gray-50 hover:bg-gray-100 p-4 rounded-lg text-left transition-colors border border-gray-200">
              <div className="text-2xl mb-2">📧</div>
              <h4 className="font-medium text-gray-900">Email Settings</h4>
              <p className="text-sm text-gray-600 mt-1">Manage notifications</p>
            </button>
            <button className="bg-gray-50 hover:bg-gray-100 p-4 rounded-lg text-left transition-colors border border-gray-200">
              <div className="text-2xl mb-2">🔒</div>
              <h4 className="font-medium text-gray-900">Security</h4>
              <p className="text-sm text-gray-600 mt-1">Review security settings</p>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}