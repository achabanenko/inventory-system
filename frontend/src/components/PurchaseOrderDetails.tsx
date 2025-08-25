import { useQuery } from '@tanstack/react-query';
import { getPurchaseOrder, type PurchaseOrder } from '../api/purchaseOrders';

const STATUS_COLORS = {
  DRAFT: 'bg-gray-100 text-gray-800',
  APPROVED: 'bg-blue-100 text-blue-800',
  PARTIAL: 'bg-yellow-100 text-yellow-800',
  RECEIVED: 'bg-green-100 text-green-800',
  CLOSED: 'bg-gray-100 text-gray-800',
  CANCELED: 'bg-red-100 text-red-800',
};

interface PurchaseOrderDetailsProps {
  isOpen: boolean;
  onClose: () => void;
  purchaseOrderId: string | null;
  onEdit?: (po: PurchaseOrder) => void;
  onReceive?: (po: PurchaseOrder) => void;
}

export default function PurchaseOrderDetails({ 
  isOpen, 
  onClose, 
  purchaseOrderId, 
  onEdit, 
  onReceive 
}: PurchaseOrderDetailsProps) {
  const { data: purchaseOrder, isLoading, error } = useQuery({
    queryKey: ['purchaseOrder', purchaseOrderId],
    queryFn: () => purchaseOrderId ? getPurchaseOrder(purchaseOrderId) : null,
    enabled: isOpen && !!purchaseOrderId,
  });

  if (!isOpen) return null;

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg p-8">
          <div className="text-center">Loading...</div>
        </div>
      </div>
    );
  }

  if (error || !purchaseOrder) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg p-8">
          <div className="text-center text-red-600">
            Error loading purchase order details
          </div>
          <div className="mt-4 text-center">
            <button onClick={onClose} className="bg-gray-500 text-white px-4 py-2 rounded">
              Close
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <h2 className="text-xl font-semibold text-gray-900">
                Purchase Order Details
              </h2>
              <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${STATUS_COLORS[purchaseOrder.status]}`}>
                {purchaseOrder.status}
              </span>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              ✕
            </button>
          </div>
        </div>

        <div className="p-6">
          {/* Header Info */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            <div>
              <h3 className="text-lg font-medium text-gray-900 mb-4">Purchase Order Info</h3>
              <dl className="space-y-2">
                <div className="flex justify-between">
                  <dt className="text-sm font-medium text-gray-500">PO Number:</dt>
                  <dd className="text-sm text-gray-900">{purchaseOrder.number}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-sm font-medium text-gray-500">Status:</dt>
                  <dd className="text-sm text-gray-900">{purchaseOrder.status}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-sm font-medium text-gray-500">Created:</dt>
                  <dd className="text-sm text-gray-900">
                    {new Date(purchaseOrder.created_at).toLocaleDateString()}
                  </dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-sm font-medium text-gray-500">Expected Date:</dt>
                  <dd className="text-sm text-gray-900">
                    {purchaseOrder.expected_at 
                      ? new Date(purchaseOrder.expected_at).toLocaleDateString()
                      : 'Not set'
                    }
                  </dd>
                </div>
                {purchaseOrder.approved_at && (
                  <div className="flex justify-between">
                    <dt className="text-sm font-medium text-gray-500">Approved:</dt>
                    <dd className="text-sm text-gray-900">
                      {new Date(purchaseOrder.approved_at).toLocaleDateString()}
                    </dd>
                  </div>
                )}
              </dl>
            </div>

            <div>
              <h3 className="text-lg font-medium text-gray-900 mb-4">Supplier Info</h3>
              <dl className="space-y-2">
                <div className="flex justify-between">
                  <dt className="text-sm font-medium text-gray-500">Name:</dt>
                  <dd className="text-sm text-gray-900">
                    {purchaseOrder.supplier?.name || 'Unknown Supplier'}
                  </dd>
                </div>
              </dl>
            </div>
          </div>

          {/* Notes */}
          {purchaseOrder.notes && (
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-2">Notes</h3>
              <p className="text-sm text-gray-700 bg-gray-50 p-3 rounded-md">
                {purchaseOrder.notes}
              </p>
            </div>
          )}

          {/* Line Items */}
          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Line Items</h3>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Item
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Qty Ordered
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Qty Received
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Unit Cost
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Line Total
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {purchaseOrder.lines?.map((line) => (
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
                        <span className={`${line.qty_received === line.qty_ordered ? 'text-green-600' : 'text-orange-600'}`}>
                          {line.qty_received}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        ${parseFloat(line.unit_cost).toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        ${parseFloat(line.line_total).toFixed(2)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Total */}
          <div className="bg-gray-50 p-4 rounded-md mb-6">
            <div className="flex justify-between items-center">
              <span className="text-lg font-medium text-gray-900">Total:</span>
              <span className="text-xl font-bold text-gray-900">
                ${parseFloat(purchaseOrder.total).toFixed(2)}
              </span>
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end space-x-3">
            <button
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Close
            </button>
            
            {purchaseOrder.status === 'DRAFT' && onEdit && (
              <button
                onClick={() => onEdit(purchaseOrder)}
                className="px-4 py-2 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700"
              >
                Edit
              </button>
            )}
            
            {(purchaseOrder.status === 'APPROVED' || purchaseOrder.status === 'PARTIAL') && onReceive && (
              <button
                onClick={() => onReceive(purchaseOrder)}
                className="px-4 py-2 bg-green-600 text-white rounded-md text-sm font-medium hover:bg-green-700"
              >
                Receive Items
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
