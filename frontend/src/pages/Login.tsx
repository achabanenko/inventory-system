import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '../contexts/AuthContext';
import { Package } from 'lucide-react';
import GoogleOAuth from '../components/GoogleOAuth';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

type LoginForm = z.infer<typeof loginSchema>;

export default function Login() {
  const navigate = useNavigate();
  const { login, user } = useAuth();
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [rememberedEmail, setRememberedEmail] = useState<string>('');
  const [rememberMe, setRememberMe] = useState<boolean>(false);

  // Load remembered email on component mount
  useEffect(() => {
    const savedEmail = localStorage.getItem('remembered_email');
    if (savedEmail) {
      setRememberedEmail(savedEmail);
    }
  }, []);

  // Watch for user authentication and redirect if user has a tenant
  useEffect(() => {
    console.log('Login useEffect triggered - user state:', user);
    if (user && user.tenantId) {
      console.log('User authenticated with tenant, redirecting to dashboard');
      navigate('/dashboard');
    }
  }, [user, navigate]);

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: rememberedEmail || 'admin@example.com',
      password: 'admin123',
    },
  });

  // Update form when remembered email changes
  useEffect(() => {
    if (rememberedEmail) {
      setValue('email', rememberedEmail);
    }
  }, [rememberedEmail, setValue]);

  const onSubmit = async (data: LoginForm) => {
    setError('');
    setIsLoading(true);

    try {
      // Call login with rememberMe option for persistent authentication
      await login(data.email, data.password, undefined, rememberMe);

      // Remember the email for next login
      localStorage.setItem('remembered_email', data.email);
      setRememberedEmail(data.email);

      navigate('/dashboard');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Invalid email or password');
    } finally {
      setIsLoading(false);
    }
  };

  const handleGoogleOAuthSuccess = async (data: any) => {
    try {
      console.log('OAuth success response:', data);

      // The AuthContext will handle token storage
      // Refresh token is stored in HttpOnly cookie by backend
      console.log('OAuth authentication successful - AuthContext will handle tokens');

      // Remember the email for next login if available
      if (data.user?.email) {
        localStorage.setItem('remembered_email', data.user.email);
        setRememberedEmail(data.user.email);
      }
    } catch (err: any) {
      console.error('OAuth success handler error:', err);
      setError('Failed to complete Google authentication');
    }
  };

  const handleGoogleOAuthError = (error: string) => {
    setError(error);
  };



  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="flex justify-center">
            <Package className="h-12 w-12 text-blue-600" />
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Inventory Management System
          </p>
        </div>
        {/* Traditional Login Form */}
        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                Email address
              </label>
              <input
                {...register('email')}
                type="email"
                autoComplete="email"
                className="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                placeholder="your@email.com"
              />
              {errors.email && (
                <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
              )}
            </div>
            

            
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <input
                {...register('password')}
                type="password"
                autoComplete="current-password"
                className="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                placeholder="Password"
              />
              {errors.password && (
                <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
              )}
            </div>

            {/* Remember Me Checkbox */}
            <div className="flex items-center">
              <input
                id="remember-me"
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-700">
                Keep me signed in across browser sessions
              </label>
            </div>
          </div>

          {error && (
            <div className="rounded-md bg-red-50 p-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? 'Signing in...' : 'Sign in'}
            </button>
          </div>

          <div className="text-sm text-center text-gray-600">
            <p>Demo credentials:</p>
            <p>Email: admin@example.com</p>
            <p>Password: admin123</p>
          </div>

          <div className="text-center space-y-2">
            <p className="text-sm text-gray-600">
              Don't have an account?{' '}
              <Link
                to="/register"
                className="font-medium text-blue-600 hover:text-blue-500"
              >
                Create one here
              </Link>
            </p>
          </div>
        </form>

        {/* Google OAuth Section - Outside the form */}
        <div className="mt-6 space-y-4">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-gray-50 text-gray-500">Or continue with</span>
            </div>
          </div>
          
          <GoogleOAuth
            onSuccess={handleGoogleOAuthSuccess}
            onError={handleGoogleOAuthError}
            className="flex justify-center"
            rememberMe={rememberMe}
          />
        </div>
      </div>
    </div>
  );
}