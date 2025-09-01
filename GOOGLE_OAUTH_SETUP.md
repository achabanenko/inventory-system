# Google OAuth Setup Guide

This document explains how to configure Google OAuth for all three components of the inventory system.

## Current Status ✅

All components now support:
- **Email/Password authentication** with JWT tokens
- **Google OAuth authentication** 
- **Automatic token persistence** and refresh
- **Secure storage** (mobile app uses Flutter Secure Storage)

## Google OAuth Configuration

### 1. Google Cloud Console Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable the Google+ API (if not already enabled)
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client IDs"

### 2. OAuth Client Configuration

Create **3 separate OAuth clients** for each platform:

#### Web Client (for frontend)
- Application type: **Web application**
- Authorized JavaScript origins:
  - `http://localhost:3000`
  - `http://localhost:5173` (Vite dev server)
  - Your production domain
- Authorized redirect URIs:
  - `http://localhost:3000/auth/google/callback`
  - `http://localhost:5173/auth/google/callback`
  - Your production OAuth callback URL

#### Android Client (for mobile app)
- Application type: **Android**
- Package name: `com.inventory.app` (or your app's package)
- SHA-1 certificate fingerprint (get with `keytool -keystore ~/.android/debug.keystore -list -v`)

#### Server Client (for backend)
- Application type: **Web application**
- This is used by the mobile app for server-side authentication
- No redirect URIs needed

### 3. Environment Variables

Update your `.env` files with the OAuth credentials:

#### Backend (.env)
```bash
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-web-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:3000/auth/google/callback
```

#### Frontend (.env.local)
```bash
VITE_GOOGLE_CLIENT_ID=your-web-client-id.apps.googleusercontent.com
```

#### Mobile App
Update `lib/services/secure_auth_service.dart`:
```dart
final GoogleSignIn _googleSignIn = GoogleSignIn(
  scopes: ['email', 'profile'],
  serverClientId: 'your-server-client-id.apps.googleusercontent.com', // Server client ID
);
```

### 4. Mobile App Platform Configuration

#### Android Configuration
Add to `android/app/build.gradle`:
```gradle
android {
    defaultConfig {
        // ... other config
        manifestPlaceholders = [
            'googleAuthRedirectScheme': 'com.inventory.app'
        ]
    }
}
```

#### iOS Configuration (if needed)
Add URL scheme to `ios/Runner/Info.plist`:
```xml
<key>CFBundleURLTypes</key>
<array>
    <dict>
        <key>CFBundleURLName</key>
        <string>google-auth</string>
        <key>CFBundleURLSchemes</key>
        <array>
            <string>com.inventory.app</string>
        </array>
    </dict>
</array>
```

## Features Implemented

### Backend ✅
- Email/password login with bcrypt hashing
- Google OAuth flow with authorization code exchange
- JWT access tokens (15 min expiry) + refresh tokens (7 days)
- Multi-tenant support
- Automatic token refresh endpoint
- Role-based access control (ADMIN/MANAGER/CLERK)

### Frontend ✅  
- Email/password login form with validation
- Google OAuth button with popup flow
- Automatic token persistence in sessionStorage
- Automatic token refresh (60 seconds before expiry)
- Protected routes with authentication checks
- User context with authentication state

### Mobile App ✅
- Email/password login with secure storage
- **NEW**: Google OAuth with Google Sign-In SDK
- **NEW**: Enhanced automatic token refresh scheduling
- **NEW**: Token expiry checking on app startup
- Biometric authentication support (optional)
- Secure token storage with Flutter Secure Storage
- Automatic API retry with token refresh on 401 errors

## Authentication Flow

### Email/Password Flow
1. User enters email/password
2. Backend validates credentials & generates JWT tokens
3. Frontend/Mobile stores access token & user info
4. Refresh token stored securely (HttpOnly cookie/secure storage)
5. Automatic refresh before token expiry

### Google OAuth Flow
1. User clicks "Continue with Google"
2. OAuth popup/SDK opens Google authentication 
3. User grants permissions
4. Google returns authorization code
5. Frontend/Mobile sends code to backend
6. Backend exchanges code for user info with Google
7. Backend creates/updates user account
8. Backend returns JWT tokens
9. Tokens stored securely with automatic refresh

## Security Features

- **JWT tokens**: Short-lived access tokens (15 min)
- **Refresh tokens**: Long-lived (7 days) for seamless re-authentication
- **Secure storage**: HttpOnly cookies (web) + Secure Storage (mobile)
- **Auto-refresh**: Transparent token renewal before expiry
- **Role-based access**: ADMIN/MANAGER/CLERK permissions
- **Biometric auth**: Optional fingerprint/face unlock (mobile)
- **CORS protection**: Configured for allowed origins
- **Input validation**: Email format, password strength checks

## Testing

### Test Accounts
- **Admin**: admin@example.com / admin123
- **Google OAuth**: Any Google account (will create new user)

### Test Flow
1. Start backend: `cd backend && go run cmd/api/main.go`
2. Start frontend: `cd frontend && npm run dev`
3. Open mobile app in emulator/device
4. Test both authentication methods
5. Verify token persistence after app restart
6. Test automatic token refresh

## Troubleshooting

### Common Issues
- **"Invalid redirect URI"**: Check OAuth client redirect URLs match exactly
- **"Invalid client ID"**: Verify client IDs in environment variables
- **Mobile OAuth fails**: Ensure server client ID is used (not Android/iOS client)
- **Token refresh fails**: Check refresh token expiry and endpoint configuration

### Debug Logs
- Backend: Check console for OAuth exchange logs
- Frontend: Open browser DevTools → Network tab
- Mobile: Check Flutter console logs for authentication flow

The authentication system now provides enterprise-grade security with seamless user experience across all platforms! 🔐✨