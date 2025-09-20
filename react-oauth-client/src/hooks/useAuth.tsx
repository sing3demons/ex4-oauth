import { useState, useEffect, createContext, useContext, ReactNode } from 'react';
import { User, LoginCredentials, RegisterData } from '../types';
import { authService } from '../services/authService';
import { StorageUtils } from '../utils/oauth';
import { toast } from 'react-hot-toast';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  login: () => Promise<void>;
  loginWithCredentials: (email: string, password: string) => Promise<User>;
  register: (userData: RegisterData) => Promise<User>;
  startOAuthFlow: () => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const isAuthenticated = !!user;

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const accessToken = StorageUtils.getAccessToken();
      if (!accessToken) {
        setLoading(false);
        return;
      }

      // Verify token and get user info
      const userInfo = await authService.getUserInfo();
      setUser(userInfo);
    } catch (error) {
      console.error('Auth check failed:', error);
      // Clear invalid tokens
      StorageUtils.clearTokens();
    } finally {
      setLoading(false);
    }
  };

  const login = async () => {
    try {
      setLoading(true);
      await authService.login();
      // If we reach here without redirect, something went wrong
      // But don't show error if OAuth flow started (redirect happened)
      if (!sessionStorage.getItem('oauth_flow_started')) {
        setLoading(false);
      }
    } catch (error) {
      // Check if this is a redirect scenario (not a real error)
      if (sessionStorage.getItem('oauth_flow_started')) {
        // OAuth flow started successfully, don't show error
        return;
      }
      
      console.error('Login failed:', error);
      toast.error('Login failed. Please try again.');
      setLoading(false);
    }
  };

  const loginWithCredentials = async (email: string, password: string): Promise<User> => {
    try {
      setLoading(true);
      const user = await authService.loginWithCredentials(email, password);
      setUser(user);
      return user;
    } catch (error) {
      console.error('Login with credentials failed:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const register = async (userData: RegisterData): Promise<User> => {
    try {
      setLoading(true);
      const user = await authService.register(userData);
      setUser(user);
      return user;
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const startOAuthFlow = async (): Promise<void> => {
    try {
      await authService.startOAuthFlow();
    } catch (error) {
      console.error('OAuth flow failed:', error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      setLoading(true);
      await authService.logout();
      setUser(null);
      toast.success('Logged out successfully');
    } catch (error) {
      console.error('Logout failed:', error);
      toast.error('Logout failed');
    } finally {
      setLoading(false);
    }
  };

  const refreshUser = async () => {
    try {
      const userInfo = await authService.getUserInfo();
      setUser(userInfo);
    } catch (error) {
      console.error('Failed to refresh user info:', error);
      // If user info fails, user might need to re-authenticate
      setUser(null);
      StorageUtils.clearTokens();
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated,
    loading,
    login,
    loginWithCredentials,
    register,
    startOAuthFlow,
    logout,
    refreshUser,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}