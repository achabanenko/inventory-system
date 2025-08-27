import { useState, useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { listSuppliers } from '../api/suppliers';
import { listLocations } from '../api/locations';
import { listReceiptLines } from '../api/receipts';
import { listPurchaseOrders } from '../api/purchaseOrders';
import type { GoodsReceipt, CreateGoodsReceiptRequest } from '../api/receipts';
import { X, Plus, Trash2, Search } from 'lucide-react';

const ITEMS_PER_PAGE = 10;

const goodsReceiptSchema = z.object({
  supplier_id: z.string().optional(),
  location_id: z.string().optional(),
  reference: z.string().optional(),
  notes: z.string().optional(),
  purchase_order_id: z.string().optional(),
  lines: z.array(z.object({
    item_identifier: z.string().min(1, 'Item code/SKU is required'),
    qty: z.number().min(1, 'Quantity must be at least 1'),
    unit_cost: z.string().min(1, 'Unit cost is required'),
  })).min(1, 'At least one line item is required'),
});

type GoodsReceiptFormData = z.infer<typeof goodsReceiptSchema>;

// CreateGoodsReceiptRequest is now imported from ../api/receipts

interface GoodsReceiptFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateGoodsReceiptRequest) => void;
  initialData?: GoodsReceipt | null;
  isSubmitting?: boolean;
  createFromPO?: boolean;
}

export default function GoodsReceiptForm({ 
  isOpen, 
  onClose, 
  onSubmit, 
  initialData, 
  isSubmitting = false,
  createFromPO = false
}: GoodsReceiptFormProps) {
  
  const queryClient = useQueryClient();
  const [currentPage, setCurrentPage] = useState(1);
  const [searchTerm, setSearchTerm] = useState('');
  // Removed unused selectedPO state

  const { data: suppliersData } = useQuery({
    queryKey: ['suppliers', { is_active: true }],
    queryFn: () => listSuppliers({ is_active: true, page_size: 100 }),
  });

  const { data: locationsData } = useQuery({
    queryKey: ['locations'],
    queryFn: () => listLocations({ page_size: 100 }),
  });

  // Removed unused itemsData query

  const { data: purchaseOrdersData } = useQuery({
    queryKey: ['purchaseOrders', { status: 'APPROVED' }],
    queryFn: () => listPurchaseOrders({ status: 'APPROVED', page_size: 100 }),
    enabled: createFromPO,
  });

  // Fetch receipt lines when editing
  const { data: receiptLinesData, refetch, isLoading: isLoadingLines, error: linesError } = useQuery({
    queryKey: ['receiptLines', initialData?.id],
    queryFn: () => listReceiptLines(initialData!.id),
    enabled: !!initialData?.id && isOpen,
    staleTime: 0, // Always consider data stale
    refetchOnMount: true, // Refetch when component mounts
    refetchOnWindowFocus: false, // Don't refetch on window focus
  });

  // Calculate initial values based on props
  const getInitialValues = (): GoodsReceiptFormData => {
    // If editing and we have receipt lines data, use that
    if (initialData && receiptLinesData?.data && receiptLinesData.data.length > 0) {
      return {
        supplier_id: initialData.supplier_id || '',
        location_id: initialData.location_id || '',
        reference: initialData.reference || '',
        notes: initialData.notes || '',
        purchase_order_id: '',
        lines: receiptLinesData.data.map(line => ({
          item_identifier: line.item?.sku || line.item?.name || line.item_id || '',
          qty: line.qty || 1,
          unit_cost: line.unit_cost ? String(line.unit_cost) : '',
        })),
      };
    }
    // If editing but no lines data yet, use header info only
    if (initialData) {
      return {
        supplier_id: initialData.supplier_id || '',
        location_id: initialData.location_id || '',
        reference: initialData.reference || '',
        notes: initialData.notes || '',
        purchase_order_id: '',
        lines: [{ item_identifier: '', qty: 1, unit_cost: '' }],
      };
    }
    // Default for new receipts
    return {
      supplier_id: '',
      location_id: '',
      reference: '',
      notes: '',
      purchase_order_id: '',
      lines: [{ item_identifier: '', qty: 1, unit_cost: '' }],
    };
  };

  const {
    register,
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<GoodsReceiptFormData>({
    resolver: zodResolver(goodsReceiptSchema),
    defaultValues: getInitialValues(),
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'lines',
  });

  // Reset form when lines data changes (for editing)
  useEffect(() => {
    if (receiptLinesData?.data && initialData) {
      console.log('Receipt lines data loaded:', receiptLinesData.data);
      const newValues = getInitialValues();
      console.log('Setting form lines to:', newValues.lines);
      setValue('lines', newValues.lines);
    }
  }, [receiptLinesData?.data, initialData, setValue]);

  // Force refetch when form opens for editing
  useEffect(() => {
    if (isOpen && initialData?.id) {
      console.log('Form opened for editing, invalidating cache and refetching for receipt:', initialData.id);
      // Invalidate the cache and refetch
      queryClient.invalidateQueries({ queryKey: ['receiptLines', initialData.id] });
      refetch();
    }
  }, [isOpen, initialData?.id, queryClient, refetch]);

  const watchedLines = watch('lines');
  // Removed unused watchedPO

  // Pagination logic
  const totalPages = Math.ceil(fields.length / ITEMS_PER_PAGE);
  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const endIndex = startIndex + ITEMS_PER_PAGE;
  const paginatedFields = fields.slice(startIndex, endIndex);

  const calculateTotal = () => {
    return watchedLines.reduce((total, line) => {
      const qty = Number(line.qty) || 0;
      const cost = parseFloat(line.unit_cost) || 0;
      return total + (qty * cost);
    }, 0);
  };

  const handleFormSubmit = (data: GoodsReceiptFormData) => {
    const formattedData: CreateGoodsReceiptRequest = {
      supplier_id: data.supplier_id || undefined,
      location_id: data.location_id || undefined,
      reference: data.reference || undefined,
      notes: data.notes || undefined,
      purchase_order_id: data.purchase_order_id || undefined,
      lines: data.lines.map(line => ({
        item_id: line.item_identifier, // This can be either item ID or SKU
        qty: line.qty,
        unit_cost: line.unit_cost,
      })),
    };
    onSubmit(formattedData);
  };

  const addLineItem = () => {
    append({ item_identifier: '', qty: 1, unit_cost: '' });
    // Navigate to the page containing the new item
    const newItemIndex = fields.length; // This will be the index after append
    const newItemPage = Math.ceil((newItemIndex + 1) / ITEMS_PER_PAGE);
    setCurrentPage(newItemPage);
  };

  const removeLineItem = (index: number) => {
    if (fields.length > 1) {
      const actualIndex = startIndex + index;
      const existingLine = initialData?.lines?.[actualIndex];
      const itemName = existingLine?.item?.name || existingLine?.item?.sku || `Line ${actualIndex + 1}`;
      
      if (window.confirm(`Are you sure you want to remove "${itemName}" from this receipt?`)) {
        remove(actualIndex);
        
        // Adjust current page if needed
        const newTotalPages = Math.ceil((fields.length - 1) / ITEMS_PER_PAGE);
        if (currentPage > newTotalPages && newTotalPages > 0) {
          setCurrentPage(newTotalPages);
        }
      }
    }
  };

  const handlePOSelection = (poId: string) => {
    // Removed unused setSelectedPO call
    setValue('purchase_order_id', poId);
    
    // Find the selected PO and populate supplier
    const selectedPOData = purchaseOrdersData?.data?.find(po => po.id === poId);
    if (selectedPOData) {
      setValue('supplier_id', selectedPOData.supplier_id || '');
      
      // Populate lines from PO
      if (selectedPOData.lines && selectedPOData.lines.length > 0) {
        // Clear existing lines first
        while (fields.length > 0) {
          remove(0);
        }
        
        // Add lines from PO
        selectedPOData.lines.forEach(line => {
          const remainingQty = (line.qty_ordered || 0) - (line.qty_received || 0);
          if (remainingQty > 0) {
            append({
              item_identifier: line.item?.sku || line.item_id,
              qty: remainingQty,
              unit_cost: String(line.unit_cost || ''),
            });
          }
        });
      }
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-6xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">
            {initialData ? 'Edit Goods Receipt' : (createFromPO ? 'Create Receipt from Purchase Order' : 'Create Goods Receipt')}
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
            {/* Basic Information */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
              {createFromPO && (
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Purchase Order
                  </label>
                  <select
                    {...register('purchase_order_id')}
                    onChange={(e) => handlePOSelection(e.target.value)}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value="">Select Purchase Order</option>
                    {purchaseOrdersData?.data?.map((po) => (
                      <option key={po.id} value={po.id}>
                        {po.number} - {po.supplier?.name} (${parseFloat(po.total).toFixed(2)})
                      </option>
                    ))}
                  </select>
                  {errors.purchase_order_id && (
                    <p className="mt-1 text-sm text-red-600">{errors.purchase_order_id.message}</p>
                  )}
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Supplier
                </label>
                <select
                  {...register('supplier_id')}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select Supplier</option>
                  {suppliersData?.data?.map((supplier) => (
                    <option key={supplier.id} value={supplier.id}>
                      {supplier.name}
                    </option>
                  ))}
                </select>
                {errors.supplier_id && (
                  <p className="mt-1 text-sm text-red-600">{errors.supplier_id.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Location
                </label>
                <select
                  {...register('location_id')}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select Location</option>
                  {locationsData?.data?.map((location) => (
                    <option key={location.id} value={location.id}>
                      {location.name} ({location.code})
                    </option>
                  ))}
                </select>
                {errors.location_id && (
                  <p className="mt-1 text-sm text-red-600">{errors.location_id.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Reference
                </label>
                <input
                  {...register('reference')}
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Reference number or document"
                />
                {errors.reference && (
                  <p className="mt-1 text-sm text-red-600">{errors.reference.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Notes
                </label>
                <textarea
                  {...register('notes')}
                  rows={3}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Additional notes"
                />
                {errors.notes && (
                  <p className="mt-1 text-sm text-red-600">{errors.notes.message}</p>
                )}
              </div>
            </div>

            {/* Line Items Section */}
            <div className="border border-gray-200 rounded-lg overflow-hidden">
              <div className="bg-gray-50 px-4 py-3 border-b border-gray-200 flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">Line Items</h3>
                <button
                  type="button"
                  onClick={addLineItem}
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add Item
                </button>
              </div>

              {/* Loading and Error States for Editing */}
              {initialData && (
                <>
                  {isLoadingLines && (
                    <div className="bg-blue-50 px-4 py-3 border-b border-blue-200">
                      <div className="flex items-center text-blue-800">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
                        Loading existing line items...
                      </div>
                    </div>
                  )}
                  {linesError && (
                    <div className="bg-red-50 px-4 py-3 border-b border-red-200">
                      <div className="flex items-center text-red-800">
                        <span className="text-sm">Failed to load line items. Please try refreshing.</span>
                        <button
                          onClick={() => refetch()}
                          className="ml-2 text-sm text-red-600 hover:text-red-800 underline"
                        >
                          Retry
                        </button>
                      </div>
                    </div>
                  )}
                </>
              )}

              {/* Items Search */}
              <div className="bg-white px-4 py-3 border-b border-gray-200">
                <div className="relative">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Search items by name or SKU..."
                    className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>

              {/* Line Items Table */}
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Item
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Quantity
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Unit Cost
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Line Total
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {paginatedFields.map((field, index) => {
                      const actualIndex = startIndex + index;
                      const lineTotal = (Number(watchedLines[actualIndex]?.qty) || 0) * (parseFloat(watchedLines[actualIndex]?.unit_cost) || 0);
                      
                      return (
                        <tr key={field.id}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <input
                              {...register(`lines.${actualIndex}.item_identifier`)}
                              type="text"
                              placeholder="Enter SKU, barcode, or item code"
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
                            <input
                              {...register(`lines.${actualIndex}.unit_cost`)}
                              type="number"
                              step="0.01"
                              min="0"
                              className="w-24 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            {errors.lines?.[actualIndex]?.unit_cost && (
                              <p className="mt-1 text-xs text-red-600">
                                {errors.lines[actualIndex]?.unit_cost?.message}
                              </p>
                            )}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            ${lineTotal.toFixed(2)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <button
                              type="button"
                              onClick={() => removeLineItem(index)}
                              disabled={fields.length <= 1}
                              className="text-red-600 hover:text-red-900 disabled:text-gray-400 disabled:cursor-not-allowed"
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="bg-white px-4 py-3 border-t border-gray-200 flex items-center justify-between">
                  <div className="flex items-center text-sm text-gray-700">
                    Showing {startIndex + 1} to {Math.min(endIndex, fields.length)} of {fields.length} items
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      type="button"
                      onClick={() => setCurrentPage(currentPage - 1)}
                      disabled={currentPage === 1}
                      className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Previous
                    </button>
                    <span className="text-sm text-gray-700">
                      Page {currentPage} of {totalPages}
                    </span>
                    <button
                      type="button"
                      onClick={() => setCurrentPage(currentPage + 1)}
                      disabled={currentPage === totalPages}
                      className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Next
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* Total */}
            <div className="mt-6 bg-gray-50 rounded-lg p-4">
              <div className="flex justify-between items-center">
                <span className="text-lg font-medium text-gray-900">Total:</span>
                <span className="text-xl font-bold text-gray-900">
                  ${calculateTotal().toFixed(2)}
                </span>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {isSubmitting ? 'Saving...' : (initialData ? 'Update Receipt' : 'Create Receipt')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
