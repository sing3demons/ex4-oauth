import { NavLink } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

export function Sidebar() {
  const { user } = useAuth();

  const navigation = [
    { name: 'Dashboard', href: '/dashboard', icon: '📊' },
    { name: 'Profile', href: '/profile', icon: '👤' },
  ];

  const adminNavigation = [
    { name: 'Admin Panel', href: '/admin', icon: '⚙️' },
  ];

  return (
    <div className="hidden lg:flex lg:flex-shrink-0">
      <div className="flex flex-col w-64 bg-white border-r border-gray-200 pt-16">
        <div className="flex-1 flex flex-col min-h-0">
          <nav className="flex-1 px-4 py-4 space-y-1">
            {navigation.map((item) => (
              <NavLink
                key={item.name}
                to={item.href}
                className={({ isActive }) =>
                  `group flex items-center px-2 py-2 text-sm font-medium rounded-md ${
                    isActive
                      ? 'bg-primary-100 text-primary-700'
                      : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                  }`
                }
              >
                <span className="mr-3 text-lg">{item.icon}</span>
                {item.name}
              </NavLink>
            ))}

            {user?.role === 'admin' && (
              <>
                <div className="border-t border-gray-200 my-4"></div>
                <div className="px-2 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wide">
                  Admin
                </div>
                {adminNavigation.map((item) => (
                  <NavLink
                    key={item.name}
                    to={item.href}
                    className={({ isActive }) =>
                      `group flex items-center px-2 py-2 text-sm font-medium rounded-md ${
                        isActive
                          ? 'bg-red-100 text-red-700'
                          : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                      }`
                    }
                  >
                    <span className="mr-3 text-lg">{item.icon}</span>
                    {item.name}
                  </NavLink>
                ))}
              </>
            )}
          </nav>
        </div>
      </div>
    </div>
  );
}