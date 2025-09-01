import axios, { AxiosError } from 'axios';
import { getTenantIdFromToken } from './jwt';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

// Token provider function that will be set by the AuthContext
let getAccessToken: (() => string | null) | null = null;
let onTokenRefresh: (() => Promise<boolean>) | null = null;

// Function to set the token provider from AuthContext
export function setTokenProvider(provider: () => string | null): void {
  getAccessToken = provider;
}

// Function to set the token refresh callback
export function setTokenRefreshCallback(callback: () => Promise<boolean>): void {
  onTokenRefresh = callback;
}

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use(
  (config) => {
    // Get token from memory provider instead of localStorage
    const token = getAccessToken ? getAccessToken() : null;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;

      // Extract and set tenant ID from JWT token
      const tenantId = getTenantIdFromToken(token);
      if (tenantId) {
        config.headers['X-Tenant-ID'] = tenantId;
      }
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as any;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Use the token refresh callback from AuthContext
        if (onTokenRefresh) {
          const refreshSuccess = await onTokenRefresh();
          if (refreshSuccess) {
            // Get the new token and retry the request
            const newToken = getAccessToken ? getAccessToken() : null;
            if (newToken) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`;
              return api(originalRequest);
            }
          }
        }

        // If refresh failed or no callback, redirect to login
        throw new Error('Token refresh failed');
      } catch (refreshError) {
        // Clear any stored tokens and redirect to login
        if (getAccessToken) {
          // The AuthContext will handle clearing tokens
        }
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default api;