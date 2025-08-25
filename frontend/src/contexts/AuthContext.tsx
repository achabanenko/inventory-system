import React, { createContext, useContext, useState, useEffect } from 'react';
import api from '../lib/api';

interface User {
  id: string;
  email: string;
  name: string;
  role: 'ADMIN' | 'MANAGER' | 'CLERK';
  tenantId: string;
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
      // For now, using placeholder data
      setUser({
        id: '1',
        email: 'admin@example.com',
        name: 'Admin User',
        role: 'ADMIN',
        tenantId: 'default-tenant-id',
      });
      
      setTenant({
        id: 'default-tenant-id',
        name: 'Default Company',
        slug: 'default',
      });
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
      
      const response = await api.post('/auth/login', loginData);
      const { access_token, refresh_token, user, tenant } = response.data;
      
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
    } catch (error) {
      console.error('Login failed:', error);
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
    <AuthContext.Provider value={{ user, tenant, login, logout, isLoading }}>
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