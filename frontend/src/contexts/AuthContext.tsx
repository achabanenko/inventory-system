import React, { createContext, useContext, useState, useEffect } from 'react';
import { login as loginApi, googleOAuth } from '../api/auth';

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
  login: (email: string, password: string, tenantSlug?: string) => Promise<void>;
  loginWithGoogle: (code: string, redirectUri: string) => Promise<any>;
  logout: () => void;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      // Fetch current user and tenant information
      fetchCurrentUserAndTenant();
    } else {
      setIsLoading(false);
    }
  }, []);

  const fetchCurrentUserAndTenant = async () => {
    try {
      // TODO: Replace with actual API calls to get user and tenant info
      // For now, don't set any default values - let the user authenticate properly
      console.log('fetchCurrentUserAndTenant called - no default values set');
    } catch (error) {
      console.error('Failed to fetch user/tenant info:', error);
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (email: string, password: string, tenantSlug?: string) => {
    try {
      const loginData: any = { email, password };
      if (tenantSlug) {
        loginData.tenant_slug = tenantSlug;
      }
      
      const response = await loginApi(email, password, tenantSlug);
      const { access_token, refresh_token, user, tenant } = response;
      
      localStorage.setItem('access_token', access_token);
      localStorage.setItem('refresh_token', refresh_token);
      
      // Set user and tenant from login response
      if (user && tenant) {
        setUser({
          id: user.id,
          email: user.email,
          name: user.name,
          role: user.role,
          tenantId: user.tenant_id,
        });
        
        setTenant({
          id: tenant.id,
          name: tenant.name,
          slug: tenant.slug,
        });
      } else {
        // Fallback to fetching user and tenant information
        await fetchCurrentUserAndTenant();
      }
    } catch (error: any) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const loginWithGoogle = async (code: string, redirectUri: string): Promise<any> => {
    try {
      console.log('loginWithGoogle called with:', { code, redirectUri });
      
      // Call the backend OAuth endpoint
      const oauthData = { code, redirect_uri: redirectUri };
      console.log('OAuth data being sent:', oauthData);
      const response = await googleOAuth(oauthData);
      console.log('OAuth response received:', response);
      
      localStorage.setItem('access_token', response.access_token);
      localStorage.setItem('refresh_token', response.refresh_token);
      
      // Set user and tenant from OAuth response (backend returns lowercase property names)
      if (response.user) {
        console.log('Setting user state:', response.user);
        const newUser = {
          id: response.user.id,
          email: response.user.email,
          name: response.user.name,
          role: response.user.role,
          tenantId: response.user.tenant_id || null, // May be null if user needs tenant selection
        };
        
        console.log('About to set user state to:', newUser);
        setUser(newUser);
        
        // Only set tenant if it exists (user already has a tenant)
        if (response.tenant) {
          console.log('Setting tenant state:', response.tenant);
          const newTenant = {
            id: response.tenant.id,
            name: response.tenant.name,
            slug: response.tenant.slug,
          };
          
          console.log('About to set tenant state to:', newTenant);
          setTenant(newTenant);
        } else {
          console.log('No tenant in response - user needs tenant selection');
          setTenant(null);
        }
        
        // Verify state was set
        setTimeout(() => {
          console.log('User state after setting:', user);
          console.log('Tenant state after setting:', tenant);
        }, 100);
      } else {
        console.log('No user in response, response structure:', response);
        console.log('Response keys:', Object.keys(response));
      }
      
      // Return the response so the calling component can access it
      return response;
    } catch (error: any) {
      console.error('Google OAuth login failed:', error);
      throw error;
    }
  };

  const logout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    setUser(null);
    setTenant(null);
  };

  return (
    <AuthContext.Provider value={{
      user,
      tenant,
      login,
      loginWithGoogle,
      logout,
      isLoading,
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