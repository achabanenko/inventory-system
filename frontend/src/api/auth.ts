import api from '../lib/api';

export interface GoogleOAuthRequest {
  code: string;
  tenant_slug?: string; // Make it truly optional
  redirect_uri: string;
}

export interface GoogleOAuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: {
    id: string;
    email: string;
    name: string;
    role: 'ADMIN' | 'MANAGER' | 'CLERK';
    tenant_id: string;
  };
  tenant?: {
    id: string;
    name: string;
    slug: string;
  };
  is_new_user: boolean;
  needs_tenant: boolean;
}

export const googleOAuth = async (data: GoogleOAuthRequest): Promise<GoogleOAuthResponse> => {
  console.log('googleOAuth API function called with:', data);
  
  // Only include tenant_slug if it's not empty
  const requestData: any = {
    code: data.code,
    redirect_uri: data.redirect_uri
  };
  
  if (data.tenant_slug && data.tenant_slug.trim() !== '') {
    requestData.tenant_slug = data.tenant_slug;
  }
  
  console.log('Final requestData being sent to backend:', requestData);
  console.log('Making POST request to /auth/google');
  
  const response = await api.post('/auth/google', requestData);
  return response.data;
};

export const login = async (email: string, password: string, tenantSlug?: string) => {
  const loginData: any = { email, password };
  if (tenantSlug) {
    loginData.tenant_slug = tenantSlug;
  }
  
  const response = await api.post('/auth/login', loginData);
  return response.data;
};

export const refreshToken = async (refreshToken: string) => {
  const response = await api.post('/auth/refresh', { refresh_token: refreshToken });
  return response.data;
};

export const logout = async () => {
  const response = await api.post('/auth/logout');
  return response.data;
};

export interface SelectTenantRequest {
  action: 'select' | 'create';
  tenant_slug?: string;
  tenant_name?: string;
  tenant_domain?: string;
}

export interface SelectTenantResponse {
  AccessToken: string;
  RefreshToken: string;
  ExpiresIn: number;
  User: {
    ID: string;
    Email: string;
    Name: string;
    Role: 'ADMIN' | 'MANAGER' | 'CLERK';
    TenantID: string;
  };
  Tenant: {
    ID: string;
    Name: string;
    Slug: string;
  };
  Message: string;
}

export const selectTenant = async (data: SelectTenantRequest): Promise<SelectTenantResponse> => {
  const response = await api.post('/auth/select-tenant', data);
  return response.data;
};
