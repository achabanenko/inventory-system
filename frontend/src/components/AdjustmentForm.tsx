import React, { useState, useEffect } from 'react';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { toast } from 'react-hot-toast';
import { listLocations } from '../api/locations';
import { getAdjustment, ADJUSTMENT_REASONS, type Adjustment, type CreateAdjustmentRequest } from '../api/adjustments';

const adjustmentSchema = z.object({
  location_id: z.string().min(1, 'Location is required'),
  reason: z.string().min(1, 'Reason is required'),
  notes: z.string().optional(),
  lines: z.array(z.object({
    item_identifier: z.string().min(1, 'Item is required'),
    qty_expected: z.number().min(0, 'Expected quantity must be 0 or greater'),
    qty_actual: z.number().min(0, 'Actual quantity must be 0 or greater'),
    notes: z.string().optional(),
  })).min(1, 'At least one item is required'),
});

type AdjustmentFormData = z.infer<typeof adjustmentSchema>;

interface AdjustmentFormProps {
  adjustment?: Adjustment | null;
  onClose: () => void;
  onSubmit: (data: CreateAdjustmentRequest) => Promise<void>;
  isSubmitting: boolean;
}

export function AdjustmentForm({ adjustment, onClose, onSubmit, isSubmitting }: AdjustmentFormProps) {
  const [locations, setLocations] = useState<any[]>([]);
  const [loadingLocations, setLoadingLocations] = useState(true);
  const [loadingAdjustment, setLoadingAdjustment] = useState(false);

  const {
    register,
    control,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
    reset,
  } = useForm<AdjustmentFormData>({
    resolver: zodResolver(adjustmentSchema),
    defaultValues: {
      location_id: '',
      reason: 'COUNT',
      notes: '',
      lines: [{ item_identifier: '', qty_expected: 0, qty_actual: 0, notes: '' }],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'lines',
  });

  // Load locations
  useEffect(() => {
    const loadLocations = async () => {
      try {
        const response = await listLocations();
        setLocations(response.data || []);
      } catch (error) {
        console.error('Failed to load locations:', error);
        toast.error('Failed to load locations');
      } finally {
        setLoadingLocations(false);
      }
    };

    loadLocations();
  }, []);

  // Load adjustment data for editing
  useEffect(() => {
    if (!adjustment?.id) return;

    const loadAdjustmentData = async () => {
      try {
        setLoadingAdjustment(true);
        const adjustmentData = await getAdjustment(adjustment.id);
        const initialValues = getInitialValues(adjustmentData);
        reset(initialValues);
      } catch (error) {
        console.error('Failed to load adjustment data:', error);
        toast.error('Failed to load adjustment data');
      } finally {
        setLoadingAdjustment(false);
      }
    };

    loadAdjustmentData();
  }, [adjustment?.id, reset]);

  const getInitialValues = (adjustmentData: Adjustment): AdjustmentFormData => {
    return {
      location_id: adjustmentData.location_id || '',
      reason: adjustmentData.reason || 'COUNT',
      notes: adjustmentData.notes || '',
      lines: adjustmentData.lines?.length ? adjustmentData.lines.map(line => ({
        item_identifier: line.item_identifier || line.item?.sku || line.item?.name || line.item_id || '',
        qty_expected: line.qty_expected || 0,
        qty_actual: line.qty_actual || 0,
        notes: line.notes || '',
      })) : [{ item_identifier: '', qty_expected: 0, qty_actual: 0, notes: '' }],
    };
  };

  const handleFormSubmit = async (data: AdjustmentFormData) => {
    const submitData: CreateAdjustmentRequest = {
      location_id: data.location_id,
      reason: data.reason,
      notes: data.notes || '',
      lines: data.lines.map(line => ({
        item_id: line.item_identifier,
        qty_expected: line.qty_expected,
        qty_actual: line.qty_actual,
        notes: line.notes || '',
      })),
    };

    await onSubmit(submitData);
  };

  const addLineItem = () => {
    append({ item_identifier: '', qty_expected: 0, qty_actual: 0, notes: '' });
  };

  const calculateDifference = (index: number) => {
    const lines = watch('lines');
    const line = lines[index];
    if (line) {
      return line.qty_actual - line.qty_expected;
    }
    return 0;
  };

  if (loadingLocations || loadingAdjustment) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg p-6 w-96">
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            <p className="mt-2 text-sm text-gray-600">Loading...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden">
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">
            {adjustment ? 'Edit Adjustment' : 'New Adjustment'}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            type="button"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit(handleFormSubmit)} className="flex flex-col h-full">
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {/* Basic Information */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Location *
                </label>
                <select
                  {...register('location_id')}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select Location</option>
                  {locations.map((location) => (
                    <option key={location.id} value={location.id}>
                      {location.name} ({location.code})
                    </option>
                  ))}
                </select>
                {errors.location_id && (
                  <p className="mt-1 text-xs text-red-600">{errors.location_id.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Reason *
                </label>
                <select
                  {...register('reason')}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  {Object.entries(ADJUSTMENT_REASONS).map(([key, value]) => (
                    <option key={key} value={value}>
                      {key === 'COUNT' ? 'Stock Count' :
                       key === 'DAMAGE' ? 'Damage' :
                       key === 'CORRECTION' ? 'Correction' :
                       key === 'EXPIRY' ? 'Expiry' :
                       key === 'THEFT' ? 'Theft' :
                       key === 'OTHER' ? 'Other' : value}
                    </option>
                  ))}
                </select>
                {errors.reason && (
                  <p className="mt-1 text-xs text-red-600">{errors.reason.message}</p>
                )}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Notes
              </label>
              <textarea
                {...register('notes')}
                rows={3}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Optional notes about this adjustment..."
              />
            </div>

            {/* Line Items */}
            <div>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">Line Items</h3>
                <button
                  type="button"
                  onClick={addLineItem}
                  className="bg-blue-600 text-white px-3 py-1 rounded-md text-sm hover:bg-blue-700 transition-colors"
                >
                  Add Item
                </button>
              </div>

              {errors.lines && (
                <p className="mb-4 text-sm text-red-600">{errors.lines.message}</p>
              )}

              <div className="overflow-x-auto">
                <table className="min-w-full border border-gray-200 rounded-lg">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Item Code *
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Expected Qty
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actual Qty
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Difference
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Notes
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {fields.map((field, index) => {
                      const actualIndex = index;
                      const difference = calculateDifference(actualIndex);
                      
                      return (
                        <tr key={field.id}>
                          <td className="px-4 py-3 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.item_identifier`)}
                              type="text"
                              placeholder="Enter item code, SKU, or name"
                              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            {errors.lines?.[actualIndex]?.item_identifier && (
                              <p className="mt-1 text-xs text-red-600">
                                {errors.lines[actualIndex]?.item_identifier?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.qty_expected`, { valueAsNumber: true })}
                              type="number"
                              min="0"
                              step="1"
                              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            {errors.lines?.[actualIndex]?.qty_expected && (
                              <p className="mt-1 text-xs text-red-600">
                                {errors.lines[actualIndex]?.qty_expected?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.qty_actual`, { valueAsNumber: true })}
                              type="number"
                              min="0"
                              step="1"
                              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            {errors.lines?.[actualIndex]?.qty_actual && (
                              <p className="mt-1 text-xs text-red-600">
                                {errors.lines[actualIndex]?.qty_actual?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap">
                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                              difference > 0 ? 'bg-green-100 text-green-800' :
                              difference < 0 ? 'bg-red-100 text-red-800' :
                              'bg-gray-100 text-gray-800'
                            }`}>
                              {difference > 0 ? '+' : ''}{difference}
                            </span>
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.notes`)}
                              type="text"
                              placeholder="Optional notes"
                              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap">
                            {fields.length > 1 && (
                              <button
                                type="button"
                                onClick={() => remove(actualIndex)}
                                className="text-red-600 hover:text-red-800"
                                title="Remove item"
                              >
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                              </button>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          {/* Form Actions */}
          <div className="flex items-center justify-end space-x-3 p-6 border-t border-gray-200 bg-gray-50">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Saving...' : (adjustment ? 'Update Adjustment' : 'Create Adjustment')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
