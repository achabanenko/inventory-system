import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { listSuppliers } from '../api/suppliers';
import { listItems } from '../api/items';
import type { CreatePurchaseOrderRequest, PurchaseOrder } from '../api/purchaseOrders';

const ITEMS_PER_PAGE = 10;

const purchaseOrderSchema = z.object({
  supplier_id: z.string().min(1, 'Supplier is required'),
  expected_at: z.string().optional(),
  notes: z.string().optional(),
  lines: z.array(z.object({
    item_id: z.string().min(1, 'Item is required'),
    qty_ordered: z.number().min(1, 'Quantity must be at least 1'),
    unit_cost: z.string().min(1, 'Unit cost is required'),
    tax: z.any().optional(),
  })).min(1, 'At least one line item is required'),
});

type PurchaseOrderFormData = z.infer<typeof purchaseOrderSchema>;

interface PurchaseOrderFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreatePurchaseOrderRequest) => void;
  initialData?: PurchaseOrder;
  isSubmitting?: boolean;
}

export default function PurchaseOrderForm({ 
  isOpen, 
  onClose, 
  onSubmit, 
  initialData, 
  isSubmitting = false 
}: PurchaseOrderFormProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);

  const { data: suppliersData } = useQuery({
    queryKey: ['suppliers', { is_active: true }],
    queryFn: () => listSuppliers({ is_active: true, page_size: 100 }),
  });

  const { data: itemsData } = useQuery({
    queryKey: ['items', { q: searchTerm }],
    queryFn: () => listItems({ q: searchTerm, page_size: 50 }),
    enabled: searchTerm.length > 0,
  });

  // Calculate initial values based on props
  const getInitialValues = (): PurchaseOrderFormData => {
    if (initialData && initialData.lines && initialData.lines.length > 0) {
      return {
        supplier_id: initialData.supplier_id || '',
        expected_at: initialData.expected_at ? initialData.expected_at.split('T')[0] : '',
        notes: initialData.notes || '',
        lines: initialData.lines.map(line => ({
          item_id: line.item_id || '',
          qty_ordered: line.qty_ordered || 1,
          unit_cost: String(line.unit_cost || ''),
          tax: line.tax || null,
        })),
      };
    }
    return {
      supplier_id: '',
      expected_at: '',
      notes: '',
      lines: [{ item_id: '', qty_ordered: 1, unit_cost: '', tax: null }],
    };
  };

  const {
    register,
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<PurchaseOrderFormData>({
    resolver: zodResolver(purchaseOrderSchema),
    defaultValues: getInitialValues(),
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'lines',
  });

  // Form is now properly initialized via key prop and getInitialValues
  // No need for useEffect reset

  const watchedLines = watch('lines');

  // Pagination logic
  const totalPages = Math.ceil(fields.length / ITEMS_PER_PAGE);
  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const endIndex = startIndex + ITEMS_PER_PAGE;
  const paginatedFields = fields.slice(startIndex, endIndex);

  const calculateTotal = () => {
    return watchedLines.reduce((total, line) => {
      const qty = Number(line.qty_ordered) || 0;
      const cost = parseFloat(line.unit_cost) || 0;
      return total + (qty * cost);
    }, 0);
  };

  const handleFormSubmit = (data: PurchaseOrderFormData) => {
    const formattedData: CreatePurchaseOrderRequest = {
      supplier_id: data.supplier_id,
      expected_at: data.expected_at || undefined,
      notes: data.notes || undefined,
      lines: data.lines.map(line => ({
        item_id: line.item_id,
        qty_ordered: line.qty_ordered,
        unit_cost: line.unit_cost,
        tax: line.tax,
      })),
    };
    onSubmit(formattedData);
  };

  const addLineItem = () => {
    append({ item_id: '', qty_ordered: 1, unit_cost: '', tax: null });
    // Navigate to the page containing the new item
    const newItemIndex = fields.length; // This will be the index after append
    const newItemPage = Math.ceil((newItemIndex + 1) / ITEMS_PER_PAGE);
    setCurrentPage(newItemPage);
  };

  const removeLineItem = (index: number) => {
    if (fields.length > 1) {
      const existingLine = initialData?.lines?.[index];
      const itemName = existingLine?.item?.name || existingLine?.item?.sku || `Line ${index + 1}`;
      
      if (window.confirm(`Are you sure you want to remove "${itemName}" from this order?`)) {
        remove(index);
      }
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-900">
              {initialData ? 'Edit Purchase Order' : 'Create Purchase Order'}
            </h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
              disabled={isSubmitting}
            >
              ✕
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit(handleFormSubmit)} className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            {/* Supplier */}
            <div>
              <label htmlFor="supplier_id" className="block text-sm font-medium text-gray-700 mb-2">
                Supplier *
              </label>
              <select
                {...register('supplier_id')}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">Select a supplier...</option>
                {suppliersData?.data.map((supplier) => (
                  <option key={supplier.id} value={supplier.id}>
                    {supplier.name} ({supplier.code})
                  </option>
                ))}
              </select>
              {errors.supplier_id && (
                <p className="mt-1 text-sm text-red-600">{errors.supplier_id.message}</p>
              )}
            </div>

            {/* Expected Date */}
            <div>
              <label htmlFor="expected_at" className="block text-sm font-medium text-gray-700 mb-2">
                Expected Date
              </label>
              <input
                type="date"
                {...register('expected_at')}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>

          {/* Notes */}
          <div className="mb-6">
            <label htmlFor="notes" className="block text-sm font-medium text-gray-700 mb-2">
              Notes
            </label>
            <textarea
              {...register('notes')}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Enter any notes or special instructions..."
            />
          </div>

          {/* Line Items */}
          <div className="mb-6">
            <div className="flex items-center justify-between mb-4">
              <div className="flex-1">
                <h3 className="text-lg font-medium text-gray-900">Line Items</h3>
                <p className="text-sm text-gray-500 mt-1">
                  {initialData ? 'Edit quantities and costs. Existing items cannot be changed. Add new lines for additional items.' : 'Add items to this purchase order.'}
                </p>
                {fields.length > ITEMS_PER_PAGE && (
                  <p className="text-xs text-gray-500 mt-1">
                    Showing {startIndex + 1}-{Math.min(endIndex, fields.length)} of {fields.length} items
                  </p>
                )}
              </div>
              <div className="flex items-center gap-3">
                {fields.length > ITEMS_PER_PAGE && (
                  <div className="flex items-center gap-2 text-sm">
                    <button
                      type="button"
                      onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                      disabled={currentPage <= 1}
                      className="px-2 py-1 border border-gray-300 rounded text-xs hover:bg-gray-50 disabled:opacity-50"
                    >
                      ‹ Prev
                    </button>
                    <span className="text-xs text-gray-500">
                      Page {currentPage} of {totalPages}
                    </span>
                    <button
                      type="button"
                      onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                      disabled={currentPage >= totalPages}
                      className="px-2 py-1 border border-gray-300 rounded text-xs hover:bg-gray-50 disabled:opacity-50"
                    >
                      Next ›
                    </button>
                  </div>
                )}
                <button
                  type="button"
                  onClick={addLineItem}
                  className="bg-green-600 text-white px-4 py-2 rounded-md text-sm hover:bg-green-700 transition-colors flex items-center gap-2"
                  disabled={isSubmitting}
                >
                  <span>➕</span> Add New Item
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <div className="bg-gray-50 px-4 py-2 rounded-md">
                <div className="grid grid-cols-12 gap-2 text-xs font-medium text-gray-600 uppercase">
                  <div className="col-span-5">Item</div>
                  <div className="col-span-2 text-center">Quantity</div>
                  <div className="col-span-2 text-center">Unit Cost</div>
                  <div className="col-span-2 text-center">Line Total</div>
                  <div className="col-span-1 text-center">Action</div>
                </div>
              </div>
              
              {paginatedFields.map((field, pageIndex) => {
                const actualIndex = startIndex + pageIndex;
                const existingLine = initialData?.lines?.[actualIndex];
                return (
                <div key={field.id} className={`border rounded-md p-3 transition-all ${
                  existingLine ? 'border-blue-200 bg-blue-50/30' : 'border-gray-200 bg-white'
                }`}>
                  <div className="grid grid-cols-12 gap-2 items-center">
                    {/* Item - Read Only */}
                    <div className="col-span-5">
                      {existingLine ? (
                        <div>
                          <div className="flex items-center">
                            <div className="flex-1">
                              <div className="font-medium text-sm text-gray-900">
                                {existingLine.item?.name || 'Unknown Item'}
                              </div>
                              <div className="text-xs text-gray-500">
                                SKU: {existingLine.item?.sku || existingLine.item_id}
                              </div>
                            </div>
                            <span className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded">Fixed</span>
                          </div>
                          {/* Hidden input to preserve existing item_id */}
                          <input type="hidden" {...register(`lines.${actualIndex}.item_id`)} value={existingLine.item_id} />
                        </div>
                      ) : (
                        <div className="flex items-center gap-2">
                          <input
                            list={`items-list-${actualIndex}`}
                            {...register(`lines.${actualIndex}.item_id`)}
                            placeholder="Enter SKU or search..."
                            className="flex-1 px-2 py-1 text-sm border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                            onChange={(e) => {
                              setSearchTerm(e.target.value);
                              setValue(`lines.${actualIndex}.item_id`, e.target.value);
                            }}
                          />
                          <datalist id={`items-list-${actualIndex}`}>
                            {itemsData?.data.map((item) => (
                              <option key={item.id} value={item.id}>
                                {item.name} ({item.sku})
                              </option>
                            ))}
                          </datalist>
                        </div>
                      )}
                      {errors.lines?.[actualIndex]?.item_id && (
                        <p className="text-xs text-red-600 mt-1">
                          {errors.lines[actualIndex]?.item_id?.message}
                        </p>
                      )}
                    </div>

                    {/* Quantity */}
                    <div className="col-span-2 text-center">
                      <input
                        type="number"
                        min="1"
                        {...register(`lines.${actualIndex}.qty_ordered`, { valueAsNumber: true })}
                        className="w-full px-2 py-1 text-sm text-center border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Qty"
                      />
                      {existingLine && (
                        <div className="text-xs text-gray-500 mt-1">
                          was: {existingLine.qty_ordered}
                        </div>
                      )}
                      {errors.lines?.[actualIndex]?.qty_ordered && (
                        <p className="text-xs text-red-600 mt-1">
                          {errors.lines[actualIndex]?.qty_ordered?.message}
                        </p>
                      )}
                    </div>

                    {/* Unit Cost */}
                    <div className="col-span-2 text-center">
                      <div className="relative">
                        <span className="absolute left-2 top-1/2 transform -translate-y-1/2 text-gray-500 text-xs">$</span>
                        <input
                          type="number"
                          step="0.01"
                          min="0"
                          {...register(`lines.${actualIndex}.unit_cost`)}
                          className="w-full pl-5 pr-2 py-1 text-sm text-center border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                          placeholder="0.00"
                        />
                      </div>
                      {existingLine && (
                        <div className="text-xs text-gray-500 mt-1">
                          was: ${existingLine.unit_cost}
                        </div>
                      )}
                      {errors.lines?.[actualIndex]?.unit_cost && (
                        <p className="text-xs text-red-600 mt-1">
                          {errors.lines[actualIndex]?.unit_cost?.message}
                        </p>
                      )}
                    </div>

                    {/* Line Total */}
                    <div className="col-span-2 text-center">
                      <div className="px-2 py-1 bg-gray-50 border border-gray-200 rounded text-sm font-medium">
                        ${((watchedLines[actualIndex]?.qty_ordered || 0) * (parseFloat(watchedLines[actualIndex]?.unit_cost) || 0)).toFixed(2)}
                      </div>
                    </div>
                    
                    {/* Remove Button */}
                    <div className="col-span-1 text-center">
                      {fields.length > 1 ? (
                        <button
                          type="button"
                          onClick={() => removeLineItem(actualIndex)}
                          className="text-red-600 hover:text-red-800 p-1 rounded hover:bg-red-50 transition-colors"
                          disabled={isSubmitting}
                          title={`Remove line item ${actualIndex + 1}`}
                        >
                          🗑️
                        </button>
                      ) : (
                        <span className="text-xs text-gray-400">—</span>
                      )}
                    </div>

                  </div>
                </div>
                );
              })}
            </div>

            {errors.lines && (
              <p className="mt-2 text-sm text-red-600">
                {errors.lines.message || 'Please fix the errors in line items'}
              </p>
            )}
          </div>

          {/* Total */}
          <div className="mb-6 bg-gray-50 p-4 rounded-md">
            <div className="flex justify-between items-center">
              <span className="text-lg font-medium text-gray-900">Total:</span>
              <span className="text-xl font-bold text-gray-900">
                ${calculateTotal().toFixed(2)}
              </span>
            </div>
          </div>

          {/* Form Actions */}
          <div className="flex justify-end space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Saving...' : (initialData ? 'Update' : 'Create')} Purchase Order
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
