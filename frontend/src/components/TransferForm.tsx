import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { X, Plus } from 'lucide-react';
import { listLocations } from '../api/locations';
import { getTransfer, type Transfer, type CreateTransferRequest } from '../api/transfers';

const ITEMS_PER_PAGE = 10;

const transferSchema = z.object({
  from_location_id: z.string().min(1, 'From location is required'),
  to_location_id: z.string().min(1, 'To location is required'),
  notes: z.string().optional(),
  lines: z.array(z.object({
    item_identifier: z.string().min(1, 'Item is required'),
    description: z.string().optional(),
    qty: z.number().min(1, 'Quantity must be at least 1'),
  })).min(1, 'At least one item is required'),
}).refine((data) => data.from_location_id !== data.to_location_id, {
  message: "From and to locations cannot be the same",
  path: ["to_location_id"],
});

type TransferFormData = z.infer<typeof transferSchema>;

interface TransferFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateTransferRequest) => void;
  initialData?: Transfer | null;
  isSubmitting?: boolean;
}

export default function TransferForm({
  isOpen,
  onClose,
  onSubmit,
  initialData,
  isSubmitting = false
}: TransferFormProps) {

  const [currentPage, setCurrentPage] = useState(1);

  // Fetch locations
  const { data: locationsData } = useQuery({
    queryKey: ['locations'],
    queryFn: () => listLocations({ page_size: 100 }),
  });



  // Fetch transfer data when editing
  const { data: transferData, refetch } = useQuery({
    queryKey: ['transfer', initialData?.id],
    queryFn: () => getTransfer(initialData!.id),
    enabled: !!initialData?.id && isOpen,
    staleTime: 0,
    refetchOnMount: true,
    refetchOnWindowFocus: false,
  });

  // Calculate initial values based on props
  const getInitialValues = (): TransferFormData => {
    if (initialData && transferData?.lines) {
      return {
        from_location_id: initialData.from_location_id || '',
        to_location_id: initialData.to_location_id || '',
        notes: initialData.notes || '',
        lines: transferData.lines.map(line => ({
          item_identifier: line.item_identifier || line.item?.sku || line.item?.name || line.item_id || '',
          description: line.description || '',
          qty: line.qty || 1,
        })),
      };
    }
    if (initialData) {
      return {
        from_location_id: initialData.from_location_id || '',
        to_location_id: initialData.to_location_id || '',
        notes: initialData.notes || '',
        lines: [{ item_identifier: '', description: '', qty: 1 }],
      };
    }
    return {
      from_location_id: '',
      to_location_id: '',
      notes: '',
      lines: [{ item_identifier: '', description: '', qty: 1 }],
    };
  };

  const {
    register,
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<TransferFormData>({
    resolver: zodResolver(transferSchema),
    defaultValues: getInitialValues(),
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'lines',
  });

  // Reset form when transfer data changes (for editing)
  useEffect(() => {
    if (transferData?.lines && initialData) {
      const newValues = getInitialValues();
      setValue('lines', newValues.lines);
      setValue('from_location_id', initialData.from_location_id);
      setValue('to_location_id', initialData.to_location_id);
      setValue('notes', initialData.notes);
    }
  }, [transferData?.lines, initialData, setValue]);

  // Force refetch when form opens for editing
  useEffect(() => {
    if (isOpen && initialData?.id) {
      refetch();
    }
  }, [isOpen, initialData?.id, refetch]);

  const watchedLines = watch('lines');

  // Pagination logic
  const totalPages = Math.ceil(fields.length / ITEMS_PER_PAGE);
  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const endIndex = startIndex + ITEMS_PER_PAGE;
  const paginatedFields = fields.slice(startIndex, endIndex);

  const handleFormSubmit = (data: TransferFormData) => {
    const formattedData: CreateTransferRequest = {
      from_location_id: data.from_location_id,
      to_location_id: data.to_location_id,
      notes: data.notes || '',
      lines: data.lines.map(line => ({
        item_id: line.item_identifier, // This can be SKU or item ID
        description: line.description || '',
        qty: line.qty,
      })),
    };
    onSubmit(formattedData);
  };

  const addLineItem = () => {
    append({ item_identifier: '', description: '', qty: 1 });
    // Navigate to the page containing the new item
    const newItemIndex = fields.length;
    const newItemPage = Math.ceil((newItemIndex + 1) / ITEMS_PER_PAGE);
    setCurrentPage(newItemPage);
  };

  const removeLineItem = (index: number) => {
    if (fields.length > 1) {
      const actualIndex = startIndex + index;
      if (window.confirm(`Remove item ${actualIndex + 1} from this transfer?`)) {
        remove(actualIndex);

        // Adjust current page if needed
        const newTotalPages = Math.ceil((fields.length - 1) / ITEMS_PER_PAGE);
        if (currentPage > newTotalPages && newTotalPages > 0) {
          setCurrentPage(newTotalPages);
        }
      }
    }
  };

  const calculateTotalItems = () => {
    return watchedLines.reduce((total, line) => total + (Number(line.qty) || 0), 0);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">
            {initialData ? 'Edit Transfer' : 'Create Transfer'}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            disabled={isSubmitting}
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit(handleFormSubmit)} className="flex-1 overflow-hidden flex flex-col">
          <div className="flex-1 overflow-y-auto p-6">
            {/* Transfer Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  From Location *
                </label>
                <select
                  {...register('from_location_id')}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select From Location</option>
                  {locationsData?.data?.map((location) => (
                    <option key={location.id} value={location.id}>
                      {location.name} ({location.code})
                    </option>
                  ))}
                </select>
                {errors.from_location_id && (
                  <p className="mt-1 text-sm text-red-600">{errors.from_location_id.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  To Location *
                </label>
                <select
                  {...register('to_location_id')}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select To Location</option>
                  {locationsData?.data?.map((location) => (
                    <option key={location.id} value={location.id}>
                      {location.name} ({location.code})
                    </option>
                  ))}
                </select>
                {errors.to_location_id && (
                  <p className="mt-1 text-sm text-red-600">{errors.to_location_id.message}</p>
                )}
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Notes
                </label>
                <textarea
                  {...register('notes')}
                  rows={3}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Optional notes about this transfer..."
                />
                {errors.notes && (
                  <p className="mt-1 text-sm text-red-600">{errors.notes.message}</p>
                )}
              </div>
            </div>

            {/* Items Section */}
            <div className="border border-gray-200 rounded-lg overflow-hidden">
              <div className="bg-gray-50 px-4 py-3 border-b border-gray-200 flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">
                  Transfer Items
                  <span className="ml-2 text-sm text-gray-500">
                    ({calculateTotalItems()} total items)
                  </span>
                </h3>
                <button
                  type="button"
                  onClick={addLineItem}
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add Item
                </button>
              </div>

              {/* Items Table */}
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Item
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Description
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Quantity
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {paginatedFields.map((field, index) => {
                      const actualIndex = startIndex + index;
                      return (
                        <tr key={field.id}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.item_identifier`)}
                              type="text"
                              placeholder="Enter SKU, barcode, or item name"
                              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            {errors.lines?.[actualIndex]?.item_identifier && (
                              <p className="mt-1 text-xs text-red-600">
                                {errors.lines[actualIndex]?.item_identifier?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.description`)}
                              type="text"
                              placeholder="Optional description"
                              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            {errors.lines?.[actualIndex]?.description && (
                              <p className="mt-1 text-xs text-red-600">
                                {errors.lines[actualIndex]?.description?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.qty`, { valueAsNumber: true })}
                              type="number"
                              min="1"
                              className="w-20 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            {errors.lines?.[actualIndex]?.qty && (
                              <p className="mt-1 text-xs text-red-600">
                                {errors.lines[actualIndex]?.qty?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <button
                              type="button"
                              onClick={() => removeLineItem(index)}
                              className="text-red-600 hover:text-red-900 text-sm"
                              disabled={fields.length <= 1}
                            >
                              Remove
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>

              {/* Pagination for items */}
              {totalPages > 1 && (
                <div className="bg-white px-4 py-3 border-t border-gray-200 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-gray-700">
                      Page {currentPage} of {totalPages}
                    </span>
                    <button
                      type="button"
                      onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                      disabled={currentPage <= 1}
                      className="px-2 py-1 text-sm border border-gray-300 rounded disabled:opacity-50"
                    >
                      Previous
                    </button>
                    <button
                      type="button"
                      onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                      disabled={currentPage >= totalPages}
                      className="px-2 py-1 text-sm border border-gray-300 rounded disabled:opacity-50"
                    >
                      Next
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-2 bg-blue-600 border border-transparent rounded-md text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
            >
              {isSubmitting && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              )}
              {initialData ? 'Update Transfer' : 'Create Transfer'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
