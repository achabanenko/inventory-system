import React, { useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';

interface GoogleOAuthProps {
  onSuccess: (data: any) => void;
  onError: (error: string) => void;
  className?: string;
}

const GoogleOAuth: React.FC<GoogleOAuthProps> = ({ 
  onSuccess, 
  onError, 
  className = '' 
}) => {
  const { loginWithGoogle } = useAuth();

  const handleGoogleSignIn = useCallback(() => {
    try {
      // Get environment variables
      const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || 
                      import.meta.env.GOOGLE_CLIENT_ID ||
                      import.meta.env.REACT_APP_GOOGLE_CLIENT_ID;
      
      if (!clientId || clientId === 'your-google-client-id-here') {
        onError('Google OAuth is not configured. Please set VITE_GOOGLE_CLIENT_ID in your .env file.');
        return;
      }

      // Debug: Log all environment variables
      console.log('All environment variables:', import.meta.env);
      console.log('VITE_ prefix variables:', Object.keys(import.meta.env).filter(key => key.startsWith('VITE_')));
      
      // Use environment variable for redirect URI, fallback to current origin
      const redirectUri = import.meta.env.VITE_GOOGLE_REDIRECT_URI || 
                         `${window.location.origin}/google-oauth-callback`;

      // Debug: Log the redirect URI being sent
      console.log('Redirect URI being sent to Google:', redirectUri);
      console.log('Environment variable VITE_GOOGLE_REDIRECT_URI:', import.meta.env.VITE_GOOGLE_REDIRECT_URI);
      console.log('Window location origin:', window.location.origin);

      // Construct OAuth 2.0 authorization URL
      const authUrl = new URL('https://accounts.google.com/o/oauth2/v2/auth');
      authUrl.searchParams.set('client_id', clientId);
      authUrl.searchParams.set('redirect_uri', redirectUri);
      authUrl.searchParams.set('response_type', 'code');
      authUrl.searchParams.set('scope', 'openid email profile');
      authUrl.searchParams.set('access_type', 'offline');
      authUrl.searchParams.set('prompt', 'consent');
      
      // Add state parameter for security
      const state = Math.random().toString(36).substring(7);
      authUrl.searchParams.set('state', state);
      
      // Debug: Log the complete OAuth URL
      console.log('Complete OAuth URL:', authUrl.toString());
      console.log('All OAuth parameters:');
      authUrl.searchParams.forEach((value, key) => {
        console.log(`  ${key}: ${value}`);
      });
      
      // Store state in localStorage for verification
      console.log('Storing OAuth state:', state);
      localStorage.setItem('google_oauth_state', state);
      
      // Verify state was stored
      const storedState = localStorage.getItem('google_oauth_state');
      console.log('Stored state verification:', { original: state, stored: storedState, match: state === storedState });

      // Open OAuth popup
      window.open(
        authUrl.toString(),
        'google-oauth',
        'width=500,height=600,scrollbars=yes,resizable=yes'
      );

      // Listen for messages from popup
      const messageListener = (event: MessageEvent) => {
        if (event.origin !== window.location.origin) return;
        
        if (event.data.type === 'GOOGLE_OAUTH_SUCCESS') {
          const { code, state: returnedState } = event.data;
          
          console.log('OAuth success message received:', { code, returnedState });
          
          // Verify state parameter
          const storedState = localStorage.getItem('google_oauth_state');
          console.log('State verification:', { 
            returnedState, 
            storedState, 
            match: returnedState === storedState,
            returnedStateType: typeof returnedState,
            storedStateType: typeof storedState,
            returnedStateLength: returnedState?.length,
            storedStateLength: storedState?.length
          });
          
          if (returnedState !== storedState) {
            console.error('OAuth state verification failed:', { 
              returnedState, 
              storedState,
              returnedStateType: typeof returnedState,
              storedStateType: typeof storedState
            });
            onError('OAuth state verification failed');
            return;
          }
          
          // Call backend with authorization code
          console.log('Calling loginWithGoogle with:', { code, redirectUri });
          
          // Clean up AFTER successful backend call
          loginWithGoogle(code, redirectUri)
            .then((response) => {
              console.log('OAuth backend call successful, cleaning up state');
              localStorage.removeItem('google_oauth_state');
              // Pass the full OAuth response data to onSuccess
              onSuccess(response);
            })
            .catch((error: any) => {
              console.error('Google OAuth error:', error);
              onError(error.response?.data?.message || 'Google authentication failed');
            });
        } else if (event.data.type === 'GOOGLE_OAUTH_ERROR') {
          onError(event.data.error || 'Google OAuth failed');
        }
      };

      window.addEventListener('message', messageListener);

      // Set a timeout to clean up the message listener if no message is received
      // This prevents memory leaks if the popup is closed manually
      const cleanupTimeout = setTimeout(() => {
        console.log('Cleaning up message listener due to timeout');
        window.removeEventListener('message', messageListener);
      }, 60000); // 60 seconds timeout

      // Store the timeout ID so we can clear it if we get a message
      const messageListenerWithCleanup = (event: MessageEvent) => {
        // Clear the timeout since we got a message
        clearTimeout(cleanupTimeout);
        
        // Call the original message listener
        messageListener(event);
        
        // Clean up the listener after processing the message
        window.removeEventListener('message', messageListenerWithCleanup);
      };

      // Replace the original listener with the one that includes cleanup
      window.removeEventListener('message', messageListener);
      window.addEventListener('message', messageListenerWithCleanup);

    } catch (error: any) {
      console.error('Failed to initiate Google OAuth:', error);
      onError('Failed to initiate Google OAuth');
    }
  }, [onSuccess, onError, loginWithGoogle]);

  return (
    <div className={className}>
      <button
        type="button"
        onClick={handleGoogleSignIn}
        className="flex items-center justify-center w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        <svg className="w-5 h-5 mr-2" viewBox="0 0 24 24">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
        </svg>
        Sign in with Google
      </button>
    </div>
  );
};

export default GoogleOAuth;
