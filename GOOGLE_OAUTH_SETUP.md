# Google OAuth Integration Setup Guide

This guide explains how to set up Google OAuth authentication for the Inventory Management System.

## Overview

The system now supports Google OAuth authentication alongside traditional email/password login. Users can sign in with their Google accounts, and the system will automatically create accounts for new users or authenticate existing ones.

## Features

- **Google Sign-In Button**: Integrated into the login page
- **Automatic Account Creation**: New users get accounts with default CLERK role
- **Tenant Isolation**: OAuth users are properly scoped to their tenant
- **Profile Information**: Automatically imports name and profile picture from Google
- **Fallback Support**: Traditional email/password login still available

## Backend Setup

### 1. Environment Variables

Create a `.env` file in the `backend/` directory with the following variables:

```bash
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id-here
GOOGLE_CLIENT_SECRET=your-google-client-secret-here
GOOGLE_REDIRECT_URL=http://localhost:5173/auth/google/callback
```

### 2. Database Migration

Run the OAuth migration to add required fields:

```bash
cd backend
go run cmd/migrate/main.go
```

Or manually execute the SQL script:

```bash
psql -d inventory -f scripts/migrate-oauth.sql
```

### 3. New API Endpoint

The system now includes a new OAuth endpoint:

- **POST** `/api/v1/auth/google` - Handle Google OAuth authentication

## Frontend Setup

### 1. Environment Variables

Create a `.env` file in the `frontend/` directory:

```bash
VITE_GOOGLE_CLIENT_ID=your-google-client-id-here
```

### 2. Google OAuth Component

The `GoogleOAuth` component automatically:
- Loads Google Identity Services
- Renders the Google Sign-In button
- Handles authentication flow
- Integrates with the existing auth system

## Google Cloud Console Setup

### 1. Create OAuth 2.0 Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the Google+ API
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client IDs"
5. Choose "Web application"
6. Add authorized redirect URIs:
   - `http://localhost:5173/auth/google/callback` (development)
   - `https://yourdomain.com/auth/google/callback` (production)

### 2. Get Client ID and Secret

Copy the generated Client ID and Client Secret to your environment files.

## How It Works

### 1. User Flow

1. User clicks "Sign in with Google" button
2. Google OAuth popup appears
3. User authenticates with Google
4. Google returns authorization code
5. Frontend sends code to backend `/auth/google` endpoint
6. Backend exchanges code for Google access token
7. Backend fetches user info from Google
8. Backend creates/updates user account
9. Backend generates JWT tokens
10. User is authenticated and redirected to dashboard

### 2. Backend Process

```go
// 1. Exchange authorization code for access token
googleToken := exchangeCodeForToken(code, redirectURI)

// 2. Get user info from Google
googleUser := getGoogleUserInfo(googleToken)

// 3. Check if user exists
user := findUserByEmail(googleUser.Email, tenantID)

// 4. Create new user if doesn't exist
if user == nil {
    user = createUser(googleUser, tenantID)
}

// 5. Generate JWT tokens
tokens := generateTokens(user)

// 6. Return authentication response
return OAuthResponse{tokens, user, tenant}
```

### 3. Database Changes

The `users` table now includes:

- `oauth_provider`: OAuth provider (e.g., "google")
- `oauth_id`: Provider's user ID
- `avatar_url`: Profile picture URL
- `password_hash`: Now optional (NULL for OAuth users)

## Security Considerations

### 1. Token Validation

- Google OAuth tokens are validated server-side
- JWT tokens are generated using the same secret as password-based auth
- All OAuth requests go through the same middleware

### 2. Tenant Isolation

- OAuth users are properly scoped to their tenant
- Users cannot access data from other tenants
- Tenant slug is required for OAuth authentication

### 3. Account Creation

- New OAuth users get the default CLERK role
- Admins can promote users to higher roles
- Email verification is handled by Google

## Testing

### 1. Development Testing

1. Set up environment variables
2. Run database migration
3. Start backend and frontend servers
4. Navigate to login page
5. Click "Sign in with Google"
6. Complete Google authentication
7. Verify user account creation/authentication

### 2. Test Scenarios

- **New User**: First-time Google sign-in should create account
- **Existing User**: Returning Google sign-in should authenticate
- **Mixed Auth**: Users can switch between Google and password auth
- **Tenant Switching**: OAuth works with different tenant slugs

## Troubleshooting

### Common Issues

1. **"Google OAuth error"**: Check client ID/secret and redirect URI
2. **"Tenant not found"**: Verify tenant slug exists and is active
3. **"Failed to create user"**: Check database permissions and schema
4. **Button not rendering**: Verify Google Identity Services script loading

### Debug Steps

1. Check browser console for JavaScript errors
2. Verify environment variables are loaded
3. Check backend logs for OAuth errors
4. Verify Google Cloud Console configuration
5. Test with different Google accounts

## Production Deployment

### 1. Environment Variables

Update environment files with production values:

```bash
# Backend
GOOGLE_REDIRECT_URL=https://yourdomain.com/auth/google/callback

# Frontend  
VITE_GOOGLE_CLIENT_ID=your-production-client-id
```

### 2. Google Cloud Console

Add production redirect URIs to OAuth credentials.

### 3. SSL Requirements

Google OAuth requires HTTPS in production.

## API Reference

### Google OAuth Endpoint

**POST** `/api/v1/auth/google`

**Request Body:**
```json
{
  "code": "google-authorization-code",
  "tenant_slug": "company-identifier",
  "redirect_uri": "http://localhost:5173/auth/google/callback"
}
```

**Response:**
```json
{
  "access_token": "jwt-access-token",
  "refresh_token": "jwt-refresh-token",
  "expires_in": 900,
  "user": {
    "id": "user-uuid",
    "email": "user@example.com",
    "name": "User Name",
    "role": "CLERK",
    "tenant_id": "tenant-uuid"
  },
  "tenant": {
    "id": "tenant-uuid",
    "name": "Company Name",
    "slug": "company-identifier"
  },
  "is_new_user": true
}
```

## Support

For issues or questions about Google OAuth integration:

1. Check the troubleshooting section above
2. Review backend logs for detailed error messages
3. Verify Google Cloud Console configuration
4. Test with a fresh Google account

## Future Enhancements

Potential improvements for the OAuth system:

- Support for additional OAuth providers (GitHub, Microsoft, etc.)
- OAuth account linking (connect multiple providers to one account)
- Enhanced profile management
- OAuth-based password reset
- Social login analytics
