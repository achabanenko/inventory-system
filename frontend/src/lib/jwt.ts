/**
 * JWT Utility Functions for production-ready token management
 */

// JWT Payload interface
export interface JWTPayload {
  user_id: string;
  tenant_id: string;
  email: string;
  role: 'ADMIN' | 'MANAGER' | 'CLERK';
  exp: number;
  iat: number;
  iss?: string;
  aud?: string;
}

/**
 * Decode JWT token without external library (base64 decode)
 */
export function decodeJWT(token: string): JWTPayload | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid JWT token format');
    }

    // Decode payload (second part)
    const payload = parts[1];
    const decodedPayload = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));

    return decodedPayload as JWTPayload;
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return null;
  }
}

/**
 * Check if JWT token is expired
 */
export function isTokenExpired(token: string): boolean {
  const payload = decodeJWT(token);
  if (!payload) return true;

  const now = Math.floor(Date.now() / 1000);
  return payload.exp < now;
}

/**
 * Check if JWT token will expire within the next N seconds
 */
export function isTokenExpiringSoon(token: string, seconds: number = 60): boolean {
  const payload = decodeJWT(token);
  if (!payload) return true;

  const now = Math.floor(Date.now() / 1000);
  const expiryTime = payload.exp;
  const threshold = now + seconds;

  return expiryTime < threshold;
}

/**
 * Get remaining seconds until token expires
 */
export function getTokenExpirySeconds(token: string): number {
  const payload = decodeJWT(token);
  if (!payload) return 0;

  const now = Math.floor(Date.now() / 1000);
  const remaining = payload.exp - now;

  return Math.max(0, remaining);
}

/**
 * Validate JWT token format and structure
 */
export function isValidJWT(token: string): boolean {
  if (!token || typeof token !== 'string') {
    return false;
  }

  const parts = token.split('.');
  if (parts.length !== 3) {
    return false;
  }

  try {
    // Try to decode payload
    const payload = decodeJWT(token);
    return payload !== null && typeof payload.exp === 'number';
  } catch {
    return false;
  }
}

/**
 * Extract user information from JWT token
 */
export function getUserFromToken(token: string): {
  id: string;
  email: string;
  role: 'ADMIN' | 'MANAGER' | 'CLERK';
  tenantId: string | null;
} | null {
  const payload = decodeJWT(token);
  if (!payload) return null;

  return {
    id: payload.user_id,
    email: payload.email,
    role: payload.role,
    tenantId: payload.tenant_id || null,
  };
}

/**
 * Check if user has required role
 */
export function hasRole(token: string, requiredRole: 'ADMIN' | 'MANAGER' | 'CLERK'): boolean {
  const user = getUserFromToken(token);
  if (!user) return false;

  const roleHierarchy = {
    ADMIN: 3,
    MANAGER: 2,
    CLERK: 1,
  };

  const userLevel = roleHierarchy[user.role];
  const requiredLevel = roleHierarchy[requiredRole];

  return userLevel >= requiredLevel;
}

/**
 * Get tenant ID from token
 */
export function getTenantIdFromToken(token: string): string | null {
  const payload = decodeJWT(token);
  return payload?.tenant_id || null;
}
