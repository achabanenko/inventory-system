import React, { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';

const GoogleOAuthCallback: React.FC = () => {
  const [searchParams] = useSearchParams();

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const error = searchParams.get('error');

    console.log('OAuth callback received:', { code, state, error });

    if (error) {
      console.error('OAuth error:', error);
      // Send error message to parent window
      window.opener?.postMessage({
        type: 'GOOGLE_OAUTH_ERROR',
        error: error
      }, window.location.origin);
      window.close();
      return;
    }

    if (code && state) {
      console.log('OAuth success, sending message to parent:', { code, state });
      // Send success message to parent window
      window.opener?.postMessage({
        type: 'GOOGLE_OAUTH_SUCCESS',
        code: code,
        state: state
      }, window.location.origin);
      window.close();
    } else {
      console.error('Missing OAuth parameters:', { code, state });
      // Missing parameters
      window.opener?.postMessage({
        type: 'GOOGLE_OAUTH_ERROR',
        error: 'Missing authorization code or state parameter'
      }, window.location.origin);
      window.close();
    }
  }, [searchParams]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p className="text-gray-600">Completing Google authentication...</p>
      </div>
    </div>
  );
};

export default GoogleOAuthCallback;
