import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getTransfer, approveTransfer, shipTransfer, receiveTransfer, deleteTransfer, type Transfer } from '../api/transfers';
import { toast } from 'react-hot-toast';
import {
  ArrowLeft,
  Edit,
  Trash2,
  CheckCircle,
  Truck,
  Package,
  User
} from 'lucide-react';

const STATUS_COLORS = {
  DRAFT: 'bg-gray-100 text-gray-800',
  IN_TRANSIT: 'bg-blue-100 text-blue-800',
  RECEIVED: 'bg-yellow-100 text-yellow-800',
  COMPLETED: 'bg-green-100 text-green-800',
  CANCELED: 'bg-red-100 text-red-800',
};

const STATUS_LABELS = {
  DRAFT: 'Draft',
  IN_TRANSIT: 'In Transit',
  RECEIVED: 'Received',
  COMPLETED: 'Completed',
  CANCELED: 'Canceled',
};

export default function TransferDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [transfer, setTransfer] = useState<Transfer | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    if (!id) {
      navigate('/transfers');
      return;
    }

    loadTransfer();
  }, [id]);

  const loadTransfer = async () => {
    if (!id) return;

    try {
      setIsLoading(true);
      const data = await getTransfer(id);
      setTransfer(data);
    } catch (error: any) {
      console.error('Failed to load transfer:', error);
      console.error('Error response:', error?.response?.data);
      toast.error(`Failed to load transfer: ${error?.response?.data?.message || error.message}`);
      navigate('/transfers');
    } finally {
      setIsLoading(false);
    }
  };

  const handleApprove = async () => {
    if (!transfer) return;

    setActionLoading('approve');
    try {
      await approveTransfer(transfer.id);
      toast.success('Transfer approved successfully');
      loadTransfer();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to approve transfer');
    } finally {
      setActionLoading(null);
    }
  };

  const handleShip = async () => {
    if (!transfer) return;

    setActionLoading('ship');
    try {
      await shipTransfer(transfer.id);
      toast.success('Transfer shipped successfully');
      loadTransfer();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to ship transfer');
    } finally {
      setActionLoading(null);
    }
  };

  const handleReceive = async () => {
    if (!transfer) return;

    setActionLoading('receive');
    try {
      await receiveTransfer(transfer.id);
      toast.success('Transfer received successfully');
      loadTransfer();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to receive transfer');
    } finally {
      setActionLoading(null);
    }
  };

  const handleDelete = async () => {
    if (!transfer) return;

    if (!confirm('Are you sure you want to delete this transfer? This action cannot be undone.')) {
      return;
    }

    try {
      await deleteTransfer(transfer.id);
      toast.success('Transfer deleted successfully');
      navigate('/transfers');
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to delete transfer');
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <div className="animate-pulse">
                <div className="h-8 bg-gray-200 rounded w-1/4"></div>
              </div>
            </div>
            <div className="p-6">
              <div className="animate-pulse space-y-4">
                <div className="h-4 bg-gray-200 rounded w-full"></div>
                <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                <div className="h-4 bg-gray-200 rounded w-1/2"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!transfer) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-white shadow rounded-lg p-6">
            <p className="text-gray-500">Transfer not found</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-4 mb-4">
            <button
              onClick={() => navigate('/transfers')}
              className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
            >
              <ArrowLeft className="h-5 w-5" />
              Back to Transfers
            </button>
          </div>

          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <div>
                  <h1 className="text-2xl font-bold text-gray-900">
                    Transfer {transfer.number}
                  </h1>
                  <div className="flex items-center gap-4 mt-2">
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${STATUS_COLORS[transfer.status]}`}>
                      {STATUS_LABELS[transfer.status]}
                    </span>
                    <span className="text-sm text-gray-500">
                      Created {new Date(transfer.created_at).toLocaleString()}
                    </span>
                  </div>
                </div>

                <div className="flex gap-2">
                  {transfer.status === 'DRAFT' && (
                    <>
                      <button
                        onClick={handleApprove}
                        disabled={actionLoading === 'approve'}
                        className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                      >
                        <CheckCircle className="h-4 w-4 mr-2" />
                        {actionLoading === 'approve' ? 'Approving...' : 'Approve'}
                      </button>
                      <button
                        onClick={() => navigate(`/transfers/${transfer.id}/edit`)}
                        className="inline-flex items-center px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700"
                      >
                        <Edit className="h-4 w-4 mr-2" />
                        Edit
                      </button>
                      <button
                        onClick={handleDelete}
                        className="inline-flex items-center px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </button>
                    </>
                  )}
                  {transfer.status === 'IN_TRANSIT' && (
                    <button
                      onClick={handleShip}
                      disabled={actionLoading === 'ship'}
                      className="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
                    >
                      <Truck className="h-4 w-4 mr-2" />
                      {actionLoading === 'ship' ? 'Shipping...' : 'Mark as Shipped'}
                    </button>
                  )}
                  {transfer.status === 'RECEIVED' && (
                    <button
                      onClick={handleReceive}
                      disabled={actionLoading === 'receive'}
                      className="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
                    >
                      <Package className="h-4 w-4 mr-2" />
                      {actionLoading === 'receive' ? 'Processing...' : 'Mark as Received'}
                    </button>
                  )}
                </div>
              </div>
            </div>

            {/* Transfer Details */}
            <div className="px-6 py-4">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                {/* From Location */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">From Location</h3>
                  <div className="space-y-1">
                    <p className="text-lg font-semibold text-gray-900">
                      {transfer.from_location?.name || 'Unknown Location'}
                    </p>
                    {transfer.from_location?.code && (
                      <p className="text-sm text-gray-500">
                        Code: {transfer.from_location.code}
                      </p>
                    )}
                  </div>
                </div>

                {/* To Location */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">To Location</h3>
                  <div className="space-y-1">
                    <p className="text-lg font-semibold text-gray-900">
                      {transfer.to_location?.name || 'Unknown Location'}
                    </p>
                    {transfer.to_location?.code && (
                      <p className="text-sm text-gray-500">
                        Code: {transfer.to_location.code}
                      </p>
                    )}
                  </div>
                </div>

                {/* Status & Dates */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">Status & Dates</h3>
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${STATUS_COLORS[transfer.status]}`}>
                        {STATUS_LABELS[transfer.status]}
                      </span>
                    </div>
                    {transfer.shipped_at && (
                      <div className="flex items-center gap-2 text-sm text-gray-600">
                        <Truck className="h-4 w-4" />
                        <span>Shipped: {new Date(transfer.shipped_at).toLocaleString()}</span>
                      </div>
                    )}
                    {transfer.received_at && (
                      <div className="flex items-center gap-2 text-sm text-gray-600">
                        <Package className="h-4 w-4" />
                        <span>Received: {new Date(transfer.received_at).toLocaleString()}</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Notes */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-900 mb-2">Notes</h3>
                  <p className="text-sm text-gray-600">
                    {transfer.notes || 'No notes provided'}
                  </p>
                </div>
              </div>

              {/* Transfer Items */}
              <div className="mt-8">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Transfer Items</h3>
                <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Item
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          SKU
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Description
                        </th>
                        <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Quantity
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {transfer.lines && transfer.lines.length > 0 ? (
                        transfer.lines.map((line, index) => (
                          <tr key={line.id || index}>
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                              {line.item?.name || line.item_identifier || 'Unknown Item'}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                              {line.item?.sku || (line.item ? 'N/A' : line.item_identifier)}
                              {!line.item && (
                                <span className="ml-2 px-2 py-1 text-xs bg-yellow-100 text-yellow-800 rounded">
                                  Not in system
                                </span>
                              )}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                              {line.description || '-'}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 text-right">
                              {line.qty}
                            </td>
                          </tr>
                        ))
                      ) : (
                        <tr>
                          <td colSpan={4} className="px-6 py-4 text-center text-sm text-gray-500">
                            No items in this transfer
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Timeline */}
              <div className="mt-8">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Timeline</h3>
                <div className="space-y-4">
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <div className="w-3 h-3 bg-gray-400 rounded-full"></div>
                      <span className="text-sm text-gray-600">Created</span>
                    </div>
                    <span className="text-sm text-gray-500">
                      {new Date(transfer.created_at).toLocaleString()}
                    </span>
                    {transfer.created_by && (
                      <div className="flex items-center gap-1 text-sm text-gray-500">
                        <User className="h-4 w-4" />
                        <span>by {transfer.created_by}</span>
                      </div>
                    )}
                  </div>

                  {transfer.approved_by && (
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
                        <span className="text-sm text-gray-600">Approved</span>
                      </div>
                      <div className="flex items-center gap-1 text-sm text-gray-500">
                        <User className="h-4 w-4" />
                        <span>by {transfer.approved_by}</span>
                      </div>
                    </div>
                  )}

                  {transfer.shipped_at && (
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                        <span className="text-sm text-gray-600">Shipped</span>
                      </div>
                      <span className="text-sm text-gray-500">
                        {new Date(transfer.shipped_at).toLocaleString()}
                      </span>
                    </div>
                  )}

                  {transfer.received_at && (
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-green-600 rounded-full"></div>
                        <span className="text-sm text-gray-600">Received</span>
                      </div>
                      <span className="text-sm text-gray-500">
                        {new Date(transfer.received_at).toLocaleString()}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
