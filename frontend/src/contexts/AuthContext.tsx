import React, { createContext, useContext, useState, useEffect, useRef, useCallback } from 'react';
import { login as loginApi, googleOAuth, refreshToken } from '../api/auth';
import { setTokenProvider, setTokenRefreshCallback } from '../lib/api';
import {
  decodeJWT,
  isValidJWT,
  isTokenExpired,
  getUserFromToken
} from '../lib/jwt';

interface User {
  id: string;
  email: string;
  name: string;
  role: 'ADMIN' | 'MANAGER' | 'CLERK';
  tenantId: string | null;
}

interface Tenant {
  id: string;
  name: string;
  slug: string;
  domain?: string;
}

interface AuthContextType {
  user: User | null;
  tenant: Tenant | null;
  login: (email: string, password: string, tenantSlug?: string, rememberMe?: boolean) => Promise<void>;
  loginWithGoogle: (code: string, redirectUri: string, rememberMe?: boolean) => Promise<any>;
  logout: () => void;
  isLoading: boolean;
  isAuthenticated: boolean;
  accessToken: string | null;
  isPersistent: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Token storage (in memory, with configurable persistence)
class TokenManager {
  private accessToken: string | null = null;
  private refreshTimer: ReturnType<typeof setTimeout> | null = null;
  private persistentMode: boolean = false; // Whether to use localStorage or sessionStorage

  constructor() {
    // First try to load from localStorage (persistent across browser sessions)
    const persistentStored = localStorage.getItem('auth_token');
    if (persistentStored) {
      try {
        const data = JSON.parse(persistentStored);
        if (data.token && isValidJWT(data.token) && !isTokenExpired(data.token)) {
          this.accessToken = data.token;
          this.persistentMode = true;
          console.log('Restored persistent authentication from localStorage');
          return;
        } else {
          // Invalid or expired token in localStorage, remove it
          localStorage.removeItem('auth_token');
          localStorage.removeItem('auth_user');
        }
      } catch (error) {
        console.error('Failed to restore token from localStorage:', error);
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user');
      }
    }

    // Fallback to sessionStorage (current session only)
    const sessionStored = sessionStorage.getItem('auth_token');
    if (sessionStored) {
      try {
        const data = JSON.parse(sessionStored);
        if (data.token && isValidJWT(data.token) && !isTokenExpired(data.token)) {
          this.accessToken = data.token;
          this.persistentMode = false;
          console.log('Restored session authentication from sessionStorage');
        }
      } catch (error) {
        console.error('Failed to restore token from sessionStorage:', error);
        sessionStorage.removeItem('auth_token');
        sessionStorage.removeItem('auth_user');
      }
    }
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  setAccessToken(token: string | null, persistent: boolean = false): void {
    this.accessToken = token;
    this.persistentMode = persistent;

    if (token) {
      const tokenData = JSON.stringify({
        token,
        timestamp: Date.now(),
      });

      if (persistent) {
        // Store in localStorage for persistence across browser sessions
        localStorage.setItem('auth_token', tokenData);
        // Also clear any session storage to avoid conflicts
        sessionStorage.removeItem('auth_token');
        console.log('Token stored persistently in localStorage');
      } else {
        // Store in sessionStorage for current session only
        sessionStorage.setItem('auth_token', tokenData);
        // Clear localStorage to avoid conflicts
        localStorage.removeItem('auth_token');
        console.log('Token stored in sessionStorage for current session');
      }
    } else {
      // Clear from both storages
      sessionStorage.removeItem('auth_token');
      localStorage.removeItem('auth_token');
    }
  }

  clearTokens(): void {
    this.accessToken = null;
    sessionStorage.removeItem('auth_token');
    localStorage.removeItem('auth_token');
  }

  // Store complete user information for persistence
  setUserInfo(user: User | null, tenant: Tenant | null): void {
    if (user) {
      const userData = JSON.stringify({
        user,
        tenant,
        timestamp: Date.now(),
      });

      if (this.persistentMode) {
        localStorage.setItem('auth_user', userData);
        sessionStorage.removeItem('auth_user');
      } else {
        sessionStorage.setItem('auth_user', userData);
        localStorage.removeItem('auth_user');
      }
    } else {
      sessionStorage.removeItem('auth_user');
      localStorage.removeItem('auth_user');
    }
  }

  // Get stored user information
  getUserInfo(): { user: User; tenant: Tenant | null } | null {
    // First try localStorage (persistent)
    let stored = localStorage.getItem('auth_user');
    if (stored) {
      try {
        const data = JSON.parse(stored);
        return {
          user: data.user,
          tenant: data.tenant,
        };
      } catch (error) {
        console.error('Failed to restore user info from localStorage:', error);
        localStorage.removeItem('auth_user');
      }
    }

    // Fallback to sessionStorage
    stored = sessionStorage.getItem('auth_user');
    if (stored) {
      try {
        const data = JSON.parse(stored);
        return {
          user: data.user,
          tenant: data.tenant,
        };
      } catch (error) {
        console.error('Failed to restore user info from sessionStorage:', error);
        sessionStorage.removeItem('auth_user');
      }
    }

    return null;
  }

  clearUserInfo(): void {
    sessionStorage.removeItem('auth_user');
    localStorage.removeItem('auth_user');
  }

  // Check if currently in persistent mode
  isPersistent(): boolean {
    return this.persistentMode;
  }

  setupAutoRefresh(onRefresh: () => Promise<boolean | void>): void {
    this.clearAutoRefresh();

    if (!this.accessToken) return;

    const payload = decodeJWT(this.accessToken);
    if (!payload) return;

    // Calculate time until we should refresh (60 seconds before expiry)
    const now = Math.floor(Date.now() / 1000);
    const refreshAt = payload.exp - 60; // 60 seconds before expiry
    const delay = Math.max(0, (refreshAt - now) * 1000);

    console.log(`Setting up token refresh in ${delay / 1000} seconds`);

    this.refreshTimer = setTimeout(async () => {
      try {
        await onRefresh();
      } catch (error) {
        console.error('Auto refresh failed:', error);
      }
    }, delay);
  }

  clearAutoRefresh(): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = null;
    }
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const tokenManager = useRef(new TokenManager()).current;

  // Auto refresh function
  const performTokenRefresh = useCallback(async (): Promise<boolean> => {
    try {
      console.log('Attempting token refresh...');

      // Call refresh endpoint (refresh token comes from HttpOnly cookie)
      const response = await refreshToken('');

      if (response.access_token) {
        tokenManager.setAccessToken(response.access_token);

        // Try to restore complete user info from sessionStorage
        const storedUserInfo = tokenManager.getUserInfo();

        if (storedUserInfo) {
          // Use stored user information to maintain consistency
          console.log('Restoring user info after token refresh:', storedUserInfo.user);
          setUser(storedUserInfo.user);
          if (storedUserInfo.tenant) {
            setTenant(storedUserInfo.tenant);
          }
        } else {
          // Fallback: extract basic info from new JWT token
          const newUser = getUserFromToken(response.access_token);
          if (newUser) {
            const basicUser = {
              id: newUser.id,
              email: newUser.email,
              name: '', // Will be populated on next login
              role: newUser.role,
              tenantId: newUser.tenantId,
            };
            setUser(basicUser);

            if (newUser.tenantId) {
              setTenant({
                id: newUser.tenantId,
                name: 'Current Tenant',
                slug: 'current-tenant',
              });
            }
          }
        }

        // Setup next auto refresh
        tokenManager.setupAutoRefresh(performTokenRefresh);
        console.log('Token refreshed successfully');
        return true;
      }
    } catch (error) {
      console.error('Token refresh failed:', error);
      // Clear tokens and logout
      tokenManager.clearTokens();
      tokenManager.clearUserInfo();
      setUser(null);
      setTenant(null);
    }
    return false;
  }, [tokenManager]);

  // Set up token provider and refresh callback for API client
  useEffect(() => {
    setTokenProvider(tokenManager.getAccessToken.bind(tokenManager));
    setTokenRefreshCallback(performTokenRefresh);
  }, [tokenManager, performTokenRefresh]);

  useEffect(() => {
    const initializeAuth = async () => {
      const token = tokenManager.getAccessToken();

      if (token && isValidJWT(token)) {
        if (isTokenExpired(token)) {
          console.log('Token expired, attempting refresh...');
          const refreshSuccess = await performTokenRefresh();
          if (!refreshSuccess) {
            // Refresh failed, clear everything
            tokenManager.clearTokens();
            tokenManager.clearUserInfo();
            setIsLoading(false);
            return;
          }
        }

        // Try to restore complete user info from sessionStorage first
        const storedUserInfo = tokenManager.getUserInfo();

        if (storedUserInfo) {
          // Use stored user information (includes name, etc.)
          console.log('Restoring user from sessionStorage:', storedUserInfo.user);
          setUser(storedUserInfo.user);
          if (storedUserInfo.tenant) {
            setTenant(storedUserInfo.tenant);
          }

          // Setup auto refresh
          tokenManager.setupAutoRefresh(performTokenRefresh);
        } else {
          // Fallback: extract basic info from JWT token
          console.log('No stored user info, extracting from JWT token');
          const userInfo = getUserFromToken(token);
          if (userInfo) {
            const basicUser = {
              id: userInfo.id,
              email: userInfo.email,
              name: '', // Will be populated on next login
              role: userInfo.role,
              tenantId: userInfo.tenantId,
            };
            setUser(basicUser);

            if (userInfo.tenantId) {
              setTenant({
                id: userInfo.tenantId,
                name: 'Current Tenant',
                slug: 'current-tenant',
              });
            }

            // Setup auto refresh
            tokenManager.setupAutoRefresh(performTokenRefresh);
          }
        }
      }

      setIsLoading(false);
    };

    initializeAuth();

    // Cleanup timer on unmount
    return () => {
      tokenManager.clearAutoRefresh();
    };
  }, [tokenManager, performTokenRefresh]);

  const login = async (email: string, password: string, tenantSlug?: string, rememberMe: boolean = false) => {
    try {
      const response = await loginApi(email, password, tenantSlug);
      const { access_token, user, tenant } = response;

      // Store access token in memory with persistence option
      tokenManager.setAccessToken(access_token, rememberMe);

      // Set user and tenant from login response
      let userData: User | null = null;
      let tenantData: Tenant | null = null;

      if (user) {
        userData = {
          id: user.id,
          email: user.email,
          name: user.name || '',
          role: user.role,
          tenantId: user.tenant_id,
        };
        setUser(userData);
      }

      if (tenant) {
        tenantData = {
          id: tenant.id,
          name: tenant.name,
          slug: tenant.slug,
        };
        setTenant(tenantData);
      }

      // Store complete user information for persistence across sessions
      tokenManager.setUserInfo(userData, tenantData);

      // Setup auto refresh
      tokenManager.setupAutoRefresh(performTokenRefresh);

    } catch (error: any) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const loginWithGoogle = async (code: string, redirectUri: string, rememberMe: boolean = false): Promise<any> => {
    try {
      console.log('loginWithGoogle called with:', { code, redirectUri });

      // Call the backend OAuth endpoint
      const oauthData = { code, redirect_uri: redirectUri };
      console.log('OAuth data being sent:', oauthData);
      const response = await googleOAuth(oauthData);
      console.log('OAuth response received:', response);

      // Store access token in memory with persistence option
      tokenManager.setAccessToken(response.access_token, rememberMe);

      // Set user and tenant from OAuth response (backend returns lowercase property names)
      let userData: User | null = null;
      let tenantData: Tenant | null = null;

      if (response.user) {
        console.log('Setting user state:', response.user);
        userData = {
          id: response.user.id,
          email: response.user.email,
          name: response.user.name || '',
          role: response.user.role,
          tenantId: response.user.tenant_id || null, // May be null if user needs tenant selection
        };

        console.log('About to set user state to:', userData);
        setUser(userData);

        // Only set tenant if it exists (user already has a tenant)
        if (response.tenant) {
          console.log('Setting tenant state:', response.tenant);
          tenantData = {
            id: response.tenant.id,
            name: response.tenant.name,
            slug: response.tenant.slug,
          };

          console.log('About to set tenant state to:', tenantData);
          setTenant(tenantData);
        } else {
          console.log('No tenant in response - user needs tenant selection');
          setTenant(null);
        }

        // Store complete user information for persistence
        tokenManager.setUserInfo(userData, tenantData);
      } else {
        console.log('No user in response, response structure:', response);
        console.log('Response keys:', Object.keys(response));
      }

      // Setup auto refresh
      tokenManager.setupAutoRefresh(performTokenRefresh);

      // Return the response so the calling component can access it
      return response;
    } catch (error: any) {
      console.error('Google OAuth login failed:', error);
      throw error;
    }
  };

  const logout = () => {
    // Clear tokens from memory and sessionStorage
    tokenManager.clearTokens();
    tokenManager.clearAutoRefresh();

    // Clear user information from sessionStorage
    tokenManager.clearUserInfo();

    // Clear user and tenant state
    setUser(null);
    setTenant(null);

    // Clear any legacy tokens from localStorage
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('tenant_id');
  };

  return (
    <AuthContext.Provider value={{
      user,
      tenant,
      login,
      loginWithGoogle,
      logout,
      isLoading,
      isAuthenticated: !!user,
      accessToken: tokenManager.getAccessToken(),
      isPersistent: tokenManager.isPersistent(),
    }}>
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