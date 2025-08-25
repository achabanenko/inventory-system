import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { PurchaseOrder, ReceiveItemsRequest } from '../api/purchaseOrders';

const receiveItemsSchema = z.object({
  lines: z.array(z.object({
    line_id: z.string(),
    qty_received: z.number().min(0, 'Quantity must be 0 or more'),
  })),
});

type ReceiveItemsFormData = z.infer<typeof receiveItemsSchema>;

interface ReceiveItemsFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: ReceiveItemsRequest) => void;
  purchaseOrder: PurchaseOrder | null;
  isSubmitting?: boolean;
}

export default function ReceiveItemsForm({ 
  isOpen, 
  onClose, 
  onSubmit, 
  purchaseOrder, 
  isSubmitting = false 
}: ReceiveItemsFormProps) {
  const {
    register,
    control,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ReceiveItemsFormData>({
    resolver: zodResolver(receiveItemsSchema),
    defaultValues: {
      lines: purchaseOrder?.lines?.map(line => ({
        line_id: line.id,
        qty_received: 0,
      })) || [],
    },
  });

  useFieldArray({
    control,
    name: 'lines',
  });

  const watchedLines = watch('lines');

  const handleFormSubmit = (data: ReceiveItemsFormData) => {
    // Only include lines with qty_received > 0
    const linesToReceive = data.lines.filter(line => line.qty_received > 0);
    
    if (linesToReceive.length === 0) {
      alert('Please enter quantities to receive for at least one item.');
      return;
    }

    onSubmit({
      lines: linesToReceive,
    });
  };

  const getMaxReceiveQuantity = (lineIndex: number) => {
    const poLine = purchaseOrder?.lines?.[lineIndex];
    if (!poLine) return 0;
    return poLine.qty_ordered - poLine.qty_received;
  };

  if (!isOpen || !purchaseOrder) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-900">
              Receive Items - {purchaseOrder.number}
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

        <div className="p-6">
          <div className="mb-4">
            <h3 className="text-lg font-medium text-gray-900 mb-2">Purchase Order Summary</h3>
            <dl className="grid grid-cols-2 gap-4">
              <div>
                <dt className="text-sm font-medium text-gray-500">Supplier:</dt>
                <dd className="text-sm text-gray-900">{purchaseOrder.supplier?.name}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Total:</dt>
                <dd className="text-sm text-gray-900">${parseFloat(purchaseOrder.total).toFixed(2)}</dd>
              </div>
            </dl>
          </div>

          <form onSubmit={handleSubmit(handleFormSubmit)}>
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Items to Receive</h3>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Item
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Ordered
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Previously Received
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Remaining
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Receive Now
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {purchaseOrder.lines?.map((line, index) => {
                      const remainingQty = line.qty_ordered - line.qty_received;
                      const maxReceive = getMaxReceiveQuantity(index);
                      
                      return (
                        <tr key={line.id}>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            <div>
                              <div className="font-medium">{line.item?.name || 'Unknown Item'}</div>
                              <div className="text-gray-500">{line.item?.sku}</div>
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {line.qty_ordered}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {line.qty_received}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            <span className={remainingQty === 0 ? 'text-green-600' : 'text-orange-600'}>
                              {remainingQty}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {remainingQty > 0 ? (
                              <div>
                                <input
                                  type="hidden"
                                  {...register(`lines.${index}.line_id`)}
                                  value={line.id}
                                />
                                <input
                                  type="number"
                                  min="0"
                                  max={maxReceive}
                                  {...register(`lines.${index}.qty_received`, { valueAsNumber: true })}
                                  className="w-20 px-2 py-1 border border-gray-300 rounded text-center focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                                />
                                {errors.lines?.[index]?.qty_received && (
                                  <p className="mt-1 text-sm text-red-600">
                                    {errors.lines[index]?.qty_received?.message}
                                  </p>
                                )}
                              </div>
                            ) : (
                              <span className="text-green-600 font-medium">Complete</span>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Summary of what will be received */}
            <div className="mb-6 bg-blue-50 p-4 rounded-md">
              <h4 className="text-sm font-medium text-blue-900 mb-2">Receiving Summary</h4>
              <div className="space-y-1">
                {watchedLines.map((watchedLine, index) => {
                  const line = purchaseOrder.lines?.[index];
                  if (!line || !watchedLine.qty_received || watchedLine.qty_received <= 0) return null;
                  
                  return (
                    <div key={line.id} className="text-sm text-blue-800">
                      {line.item?.name} ({line.item?.sku}): {watchedLine.qty_received} units
                    </div>
                  );
                })}
                {watchedLines.every(line => !line.qty_received || line.qty_received <= 0) && (
                  <div className="text-sm text-blue-800">No items selected for receiving</div>
                )}
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
                className="px-4 py-2 bg-green-600 text-white rounded-md text-sm font-medium hover:bg-green-700 disabled:opacity-50"
                disabled={isSubmitting || watchedLines.every(line => !line.qty_received || line.qty_received <= 0)}
              >
                {isSubmitting ? 'Receiving...' : 'Receive Items'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
