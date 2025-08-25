import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Building, Plus, Users, ArrowRight } from 'lucide-react';
import { selectTenant } from '../api/auth';

const tenantSchema = z.object({
  action: z.enum(['select', 'create']),
  tenant_slug: z.string().min(1, 'Company identifier is required'),
  tenant_name: z.string().optional(),
  tenant_domain: z.string().optional(),
}).refine((data) => {
  if (data.action === 'create') {
    return data.tenant_name && data.tenant_name.length > 0;
  }
  return true;
}, {
  message: 'Company name is required when creating a new company',
  path: ['tenant_name'],
});

type TenantForm = z.infer<typeof tenantSchema>;

interface TenantSelectionProps {
  onSuccess: (data: any) => void;
  onError: (error: string) => void;
}

const TenantSelection: React.FC<TenantSelectionProps> = ({ onSuccess, onError }) => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [action, setAction] = useState<'select' | 'create'>('select');

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TenantForm>({
    resolver: zodResolver(tenantSchema),
    defaultValues: {
      action: 'select',
      tenant_slug: '',
      tenant_name: '',
      tenant_domain: '',
    },
  });



  const onSubmit = async (data: TenantForm) => {
    setIsLoading(true);
    try {
      const response = await selectTenant(data);
      
      // Store new tokens (backend returns capitalized property names)
      localStorage.setItem('access_token', response.AccessToken);
      localStorage.setItem('refresh_token', response.RefreshToken);
      
      onSuccess(response);
      navigate('/dashboard');
    } catch (error: any) {
      onError(error.response?.data?.message || 'Failed to select tenant');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="flex justify-center">
            <Building className="h-12 w-12 text-blue-600" />
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Welcome! Let's get you set up
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Choose how you'd like to proceed with your account
          </p>
        </div>

        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          <div className="space-y-6">
            {/* Action Selection */}
            <div className="space-y-4">
              <div className="flex space-x-4">
                <button
                  type="button"
                  onClick={() => setAction('select')}
                  className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                    action === 'select'
                      ? 'bg-blue-100 text-blue-700 border-2 border-blue-300'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  <Users className="inline-block w-4 h-4 mr-2" />
                  Join Existing Company
                </button>
                <button
                  type="button"
                  onClick={() => setAction('create')}
                  className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
                    action === 'create'
                      ? 'bg-blue-100 text-blue-700 border-2 border-blue-300'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  <Plus className="inline-block w-4 h-4 mr-2" />
                  Create New Company
                </button>
              </div>
            </div>

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              <input type="hidden" {...register('action')} value={action} />

              {/* Company Identifier */}
              <div>
                <label htmlFor="tenant_slug" className="block text-sm font-medium text-gray-700">
                  Company Identifier
                </label>
                <div className="mt-1">
                  <input
                    {...register('tenant_slug')}
                    type="text"
                    className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    placeholder={action === 'select' ? 'Enter company identifier' : 'Choose a unique identifier'}
                  />
                </div>
                {errors.tenant_slug && (
                  <p className="mt-1 text-sm text-red-600">{errors.tenant_slug.message}</p>
                )}
                <p className="mt-1 text-xs text-gray-500">
                  {action === 'select' 
                    ? 'Enter the identifier provided by your company'
                    : 'This will be used in your company URL (e.g., company-name.yourdomain.com)'
                  }
                </p>
              </div>

              {/* Company Name (for create action) */}
              {action === 'create' && (
                <div>
                  <label htmlFor="tenant_name" className="block text-sm font-medium text-gray-700">
                    Company Name
                  </label>
                  <div className="mt-1">
                    <input
                      {...register('tenant_name')}
                      type="text"
                      className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="Enter your company name"
                    />
                  </div>
                  {errors.tenant_name && (
                    <p className="mt-1 text-sm text-red-600">{errors.tenant_name.message}</p>
                  )}
                </div>
              )}

              {/* Company Domain (optional for create action) */}
              {action === 'create' && (
                <div>
                  <label htmlFor="tenant_domain" className="block text-sm font-medium text-gray-700">
                    Company Domain (Optional)
                  </label>
                  <div className="mt-1">
                    <input
                      {...register('tenant_domain')}
                      type="text"
                      className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="yourcompany.com"
                    />
                  </div>
                  <p className="mt-1 text-xs text-gray-500">
                    Your company's website domain
                  </p>
                </div>
              )}

              <div>
                <button
                  type="submit"
                  disabled={isLoading}
                  className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isLoading ? (
                    'Setting up...'
                  ) : (
                    <>
                      {action === 'select' ? 'Join Company' : 'Create Company'}
                      <ArrowRight className="ml-2 h-4 w-4" />
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TenantSelection;
