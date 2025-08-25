import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Package, Building, ArrowLeft, ExternalLink } from 'lucide-react';
import api from '../lib/api';

const lookupSchema = z.object({
  email: z.string().email('Invalid email address'),
});

type LookupForm = z.infer<typeof lookupSchema>;

interface TenantResult {
  id: string;
  name: string;
  slug: string;
  role: string;
}

export default function TenantLookup() {
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [tenants, setTenants] = useState<TenantResult[]>([]);
  const [searchPerformed, setSearchPerformed] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LookupForm>();

  const onSubmit = async (data: LookupForm) => {
    setError('');
    setIsLoading(true);
    setSearchPerformed(false);
    
    try {
      const response = await api.get(`/auth/tenant-lookup?email=${encodeURIComponent(data.email)}`);
      setTenants(response.data.tenants || []);
      setSearchPerformed(true);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to lookup tenants');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="flex justify-center">
            <Package className="h-12 w-12 text-blue-600" />
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Find Your Companies
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Enter your email to see which companies you're registered with
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700">
              Email Address
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
              {isLoading ? 'Searching...' : 'Find My Companies'}
            </button>
          </div>
        </form>

        {/* Results */}
        {searchPerformed && (
          <div className="mt-6">
            {tenants.length > 0 ? (
              <>
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  Found {tenants.length} company{tenants.length === 1 ? '' : 'ies'}:
                </h3>
                <div className="space-y-3">
                  {tenants.map((tenant) => (
                    <div
                      key={tenant.id}
                      className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center">
                          <Building className="h-8 w-8 text-gray-400 mr-3" />
                          <div>
                            <h4 className="text-sm font-medium text-gray-900">
                              {tenant.name}
                            </h4>
                            <p className="text-xs text-gray-500">
                              Role: {tenant.role} • ID: {tenant.slug}
                            </p>
                          </div>
                        </div>
                        <Link
                          to={`/login?tenant=${tenant.slug}`}
                          className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded text-blue-700 bg-blue-100 hover:bg-blue-200"
                        >
                          Login
                          <ExternalLink className="ml-1 h-3 w-3" />
                        </Link>
                      </div>
                    </div>
                  ))}
                </div>
              </>
            ) : (
              <div className="text-center py-8">
                <Building className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">No companies found</h3>
                <p className="mt-1 text-sm text-gray-500">
                  No companies are associated with this email address.
                </p>
                <div className="mt-6">
                  <Link
                    to="/register"
                    className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
                  >
                    Create New Company
                  </Link>
                </div>
              </div>
            )}
          </div>
        )}

        <div className="flex items-center justify-center">
          <Link
            to="/login"
            className="flex items-center text-sm text-blue-600 hover:text-blue-500"
          >
            <ArrowLeft className="h-4 w-4 mr-1" />
            Back to login
          </Link>
        </div>
      </div>
    </div>
  );
}
