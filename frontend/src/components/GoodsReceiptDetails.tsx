import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getReceipt, approveReceipt, postReceipt, closeReceipt, type GoodsReceipt } from '../api/receipts';
import { toast } from 'react-hot-toast';
import { ArrowLeft, CheckCircle, Truck, XCircle } from 'lucide-react';

export default function GoodsReceiptDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [receipt, setReceipt] = useState<GoodsReceipt | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const loadReceipt = async () => {
    if (!id) return;
    setIsLoading(true);
    try {
      const data = await getReceipt(id);
      setReceipt(data);
    } catch (error) {
      toast.error('Failed to load receipt');
      navigate('/receipts');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadReceipt();
  }, [id]);

  const handleApprove = async () => {
    if (!receipt) return;
    setActionLoading('approve');
    try {
      await approveReceipt(receipt.id);
      toast.success('Receipt approved successfully');
      loadReceipt();
    } catch (error) {
      toast.error('Failed to approve receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const handlePost = async () => {
    if (!receipt) return;
    setActionLoading('post');
    try {
      await postReceipt(receipt.id);
      toast.success('Receipt posted to inventory successfully');
      loadReceipt();
    } catch (error) {
      toast.error('Failed to post receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const handleClose = async () => {
    if (!receipt) return;
    setActionLoading('close');
    try {
      await closeReceipt(receipt.id);
      toast.success('Receipt closed successfully');
      loadReceipt();
    } catch (error) {
      toast.error('Failed to close receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'DRAFT':
        return 'bg-gray-100 text-gray-800';
      case 'APPROVED':
        return 'bg-blue-100 text-blue-800';
      case 'POSTED':
        return 'bg-green-100 text-green-800';
      case 'CLOSED':
        return 'bg-purple-100 text-purple-800';
      case 'CANCELED':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const canApprove = receipt?.status === 'DRAFT';
  const canPost = receipt?.status === 'APPROVED';
  const canClose = receipt?.status === 'POSTED';

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-sm text-gray-500">Loading receipt...</p>
        </div>
      </div>
    );
  }

  if (!receipt) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Receipt not found</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center">
          <button
            onClick={() => navigate('/receipts')}
            className="mr-4 p-2 hover:bg-gray-100 rounded-full"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Receipt {receipt.number}</h1>
            <div className="flex items-center mt-1 space-x-4">
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(receipt.status)}`}>
                {receipt.status}
              </span>
              <span className="text-sm text-gray-500">
                Created {new Date(receipt.created_at).toLocaleDateString()}
              </span>
            </div>
          </div>
        </div>

        <div className="flex space-x-2">
          {canApprove && (
            <button
              onClick={handleApprove}
              disabled={actionLoading === 'approve'}
              className="flex items-center px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              <CheckCircle className="h-4 w-4 mr-2" />
              {actionLoading === 'approve' ? 'Approving...' : 'Approve'}
            </button>
          )}
          {canPost && (
            <button
              onClick={handlePost}
              disabled={actionLoading === 'post'}
              className="flex items-center px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
            >
              <Truck className="h-4 w-4 mr-2" />
              {actionLoading === 'post' ? 'Posting...' : 'Post to Inventory'}
            </button>
          )}
          {canClose && (
            <button
              onClick={handleClose}
              disabled={actionLoading === 'close'}
              className="flex items-center px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 disabled:opacity-50"
            >
              <XCircle className="h-4 w-4 mr-2" />
              {actionLoading === 'close' ? 'Closing...' : 'Close'}
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Receipt Information */}
        <div className="lg:col-span-1">
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Receipt Information</h2>
            <dl className="space-y-3 text-sm">
              <div>
                <dt className="text-gray-500">Number</dt>
                <dd className="text-gray-900 font-medium">{receipt.number}</dd>
              </div>
              <div>
                <dt className="text-gray-500">Status</dt>
                <dd>
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(receipt.status)}`}>
                    {receipt.status}
                  </span>
                </dd>
              </div>
              {receipt.supplier && (
                <div>
                  <dt className="text-gray-500">Supplier</dt>
                  <dd className="text-gray-900">{receipt.supplier.name}</dd>
                </div>
              )}
              {receipt.location && (
                <div>
                  <dt className="text-gray-500">Location</dt>
                  <dd className="text-gray-900">{receipt.location.name} ({receipt.location.code})</dd>
                </div>
              )}
              {receipt.reference && (
                <div>
                  <dt className="text-gray-500">Reference</dt>
                  <dd className="text-gray-900">{receipt.reference}</dd>
                </div>
              )}
              {receipt.notes && (
                <div>
                  <dt className="text-gray-500">Notes</dt>
                  <dd className="text-gray-900">{receipt.notes}</dd>
                </div>
              )}
              <div>
                <dt className="text-gray-500">Total</dt>
                <dd className="text-gray-900 font-bold text-lg">${typeof receipt.total === 'number' ? receipt.total.toFixed(2) : (parseFloat(receipt.total || '0') || 0).toFixed(2)}</dd>
              </div>
            </dl>
          </div>

          {/* Workflow Information */}
          <div className="bg-white shadow rounded-lg p-6 mt-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Workflow</h2>
            <div className="space-y-4">
              <div className="flex items-center">
                <div className="flex-shrink-0 w-2 h-2 bg-green-400 rounded-full"></div>
                <div className="ml-3 text-sm text-gray-900">
                  Created {new Date(receipt.created_at).toLocaleDateString()}
                </div>
              </div>
              {receipt.approved_at && (
                <div className="flex items-center">
                  <div className="flex-shrink-0 w-2 h-2 bg-blue-400 rounded-full"></div>
                  <div className="ml-3 text-sm text-gray-900">
                    Approved {new Date(receipt.approved_at).toLocaleDateString()}
                  </div>
                </div>
              )}
              {receipt.posted_at && (
                <div className="flex items-center">
                  <div className="flex-shrink-0 w-2 h-2 bg-green-400 rounded-full"></div>
                  <div className="ml-3 text-sm text-gray-900">
                    Posted {new Date(receipt.posted_at).toLocaleDateString()}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Receipt Lines */}
        <div className="lg:col-span-2">
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-medium text-gray-900">Items</h2>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Item
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Quantity
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Unit Cost
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Line Total
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {receipt.lines?.map((line) => (
                    <tr key={line.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        <div>
                          <div className="font-medium">{line.item?.name || line.item_id}</div>
                          {line.item && (
                            <div className="text-gray-500">{line.item.sku}</div>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                        {line.qty}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">
                        ${typeof line.unit_cost === 'number' ? line.unit_cost.toFixed(2) : (parseFloat(line.unit_cost?.toString() || '0') || 0).toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900 font-medium">
                        ${typeof line.line_total === 'number' ? line.line_total.toFixed(2) : (parseFloat(line.line_total?.toString() || '0') || 0).toFixed(2)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}