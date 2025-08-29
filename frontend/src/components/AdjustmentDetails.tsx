import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { toast } from 'react-hot-toast';
import {
  getAdjustment,
  deleteAdjustment,
  approveAdjustment,
  getAdjustmentStatusLabel,
  getAdjustmentStatusColor,
  getAdjustmentReasonLabel,
  type Adjustment,
} from '../api/adjustments';
import { AdjustmentForm } from './AdjustmentForm';

export function AdjustmentDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [adjustment, setAdjustment] = useState<Adjustment | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [formOpen, setFormOpen] = useState(false);

  useEffect(() => {
    if (!id) return;

    const loadAdjustment = async () => {
      try {
        setIsLoading(true);
        const data = await getAdjustment(id);
        setAdjustment(data);
      } catch (error: any) {
        console.error('Failed to load adjustment:', error);
        console.error('Error response:', error?.response?.data);
        toast.error(`Failed to load adjustment: ${error?.response?.data?.message || error.message}`);
        navigate('/adjustments');
      } finally {
        setIsLoading(false);
      }
    };

    loadAdjustment();
  }, [id, navigate]);

  const handleEdit = () => {
    setFormOpen(true);
  };

  const handleFormClose = () => {
    setFormOpen(false);
  };

  const handleFormSubmit = async () => {
    // This will be handled by the parent Adjustments component
    // For now, just close the form and reload
    setFormOpen(false);
    if (id) {
      const data = await getAdjustment(id);
      setAdjustment(data);
    }
  };

  const handleDelete = async () => {
    if (!adjustment || !confirm('Are you sure you want to delete this adjustment?')) return;

    try {
      await deleteAdjustment(adjustment.id);
      toast.success('Adjustment deleted successfully');
      navigate('/adjustments');
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to delete adjustment');
    }
  };

  const handleApprove = async () => {
    if (!adjustment || !confirm('Are you sure you want to approve this adjustment? This will update inventory levels.')) return;

    try {
      await approveAdjustment(adjustment.id);
      toast.success('Adjustment approved successfully');
      // Reload the adjustment to get updated status
      if (id) {
        const data = await getAdjustment(id);
        setAdjustment(data);
      }
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to approve adjustment');
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600">Loading adjustment details...</p>
        </div>
      </div>
    );
  }

  if (!adjustment) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Adjustment not found</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Adjustment {adjustment.number}
          </h1>
          <p className="text-gray-600">
            {getAdjustmentReasonLabel(adjustment.reason)} • {adjustment.location?.name}
          </p>
        </div>
        <div className="flex items-center space-x-3">
          {adjustment.status === 'DRAFT' && (
            <>
              <button
                onClick={handleEdit}
                className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
              >
                Edit
              </button>
              <button
                onClick={handleDelete}
                className="bg-red-600 text-white px-4 py-2 rounded-md hover:bg-red-700 transition-colors"
              >
                Delete
              </button>
              <button
                onClick={handleApprove}
                className="bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 transition-colors"
              >
                Approve
              </button>
            </>
          )}
        </div>
      </div>

      {/* Status and Details */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Status</h3>
          <span className={`inline-flex px-3 py-1 text-sm font-semibold rounded-full ${getAdjustmentStatusColor(adjustment.status)}`}>
            {getAdjustmentStatusLabel(adjustment.status)}
          </span>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Location</h3>
          <p className="text-lg font-semibold text-gray-900">
            {adjustment.location?.name || 'Unknown Location'}
          </p>
          <p className="text-sm text-gray-600">
            {adjustment.location?.code || ''}
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Reason</h3>
          <p className="text-lg font-semibold text-gray-900">
            {getAdjustmentReasonLabel(adjustment.reason)}
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <h3 className="text-sm font-medium text-gray-500 mb-2">Created</h3>
          <p className="text-lg font-semibold text-gray-900">
            {new Date(adjustment.created_at).toLocaleDateString()}
          </p>
          <p className="text-sm text-gray-600">
            {new Date(adjustment.created_at).toLocaleTimeString()}
          </p>
        </div>
      </div>

      {/* Notes */}
      {adjustment.notes && (
        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <h3 className="text-lg font-medium text-gray-900 mb-3">Notes</h3>
          <p className="text-gray-700 whitespace-pre-wrap">{adjustment.notes}</p>
        </div>
      )}

      {/* Line Items */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Line Items</h3>
        </div>
        
        {!adjustment.lines || adjustment.lines.length === 0 ? (
          <div className="p-6 text-center text-gray-500">
            No items in this adjustment
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Item
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Expected
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actual
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Difference
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Notes
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {adjustment.lines.map((line) => (
                  <tr key={line.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="text-sm font-medium text-gray-900">
                          {line.item?.name || line.item_identifier || 'Unknown Item'}
                        </div>
                        <div className="text-sm text-gray-500">
                          SKU: {line.item?.sku || (line.item ? 'N/A' : line.item_identifier)}
                        </div>
                        {!line.item && (
                          <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800">
                            Not in system
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {line.qty_expected}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {line.qty_actual}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                        line.qty_diff > 0 ? 'bg-green-100 text-green-800' :
                        line.qty_diff < 0 ? 'bg-red-100 text-red-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {line.qty_diff > 0 ? '+' : ''}{line.qty_diff}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {line.notes || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Timeline */}
      <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Timeline</h3>
        <div className="space-y-4">
          <div className="flex items-start space-x-3">
            <div className="flex-shrink-0 w-2 h-2 bg-blue-600 rounded-full mt-2"></div>
            <div>
              <p className="text-sm font-medium text-gray-900">Adjustment Created</p>
              <p className="text-sm text-gray-500">
                {new Date(adjustment.created_at).toLocaleString()}
              </p>
            </div>
          </div>
          
          {adjustment.approved_at && (
            <div className="flex items-start space-x-3">
              <div className="flex-shrink-0 w-2 h-2 bg-green-600 rounded-full mt-2"></div>
              <div>
                <p className="text-sm font-medium text-gray-900">Adjustment Approved</p>
                <p className="text-sm text-gray-500">
                  {new Date(adjustment.approved_at).toLocaleString()}
                </p>
                {adjustment.approved_by && (
                  <p className="text-sm text-gray-500">by {adjustment.approved_by}</p>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Form Modal */}
      {formOpen && (
        <AdjustmentForm
          adjustment={adjustment}
          onClose={handleFormClose}
          onSubmit={handleFormSubmit}
          isSubmitting={false}
        />
      )}
    </div>
  );
}
