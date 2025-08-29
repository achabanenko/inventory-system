import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { toast } from 'react-hot-toast';
import {
  listAdjustments,
  deleteAdjustment,
  approveAdjustment,
  createAdjustment,
  updateAdjustment,
  getAdjustmentStatusLabel,
  getAdjustmentStatusColor,
  getAdjustmentReasonLabel,
  type Adjustment,
  type ListAdjustmentsParams,
  type PaginatedResponse,
  ADJUSTMENT_STATUSES,
  ADJUSTMENT_REASONS,
} from '../api/adjustments';
import { AdjustmentForm } from '../components/AdjustmentForm';

export default function Adjustments() {
  const [adjustments, setAdjustments] = useState<Adjustment[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);

  // Filters
  const [statusFilter, setStatusFilter] = useState('');
  const [reasonFilter, setReasonFilter] = useState('');
  const [searchFilter, setSearchFilter] = useState('');

  // Form state
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Adjustment | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const reload = async () => {
    try {
      setLoading(true);
      const params: ListAdjustmentsParams = {
        page,
        limit,
      };

      if (statusFilter) params.status = statusFilter;
      if (reasonFilter) params.reason = reasonFilter;
      if (searchFilter) params.search = searchFilter;

      const response: PaginatedResponse<Adjustment> = await listAdjustments(params);
      setAdjustments(response.data || []);
      setTotal(response.total || 0);
    } catch (error: any) {
      console.error('Failed to load adjustments:', error);
      toast.error('Failed to load adjustments');
      setAdjustments([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    reload();
  }, [page, statusFilter, reasonFilter, searchFilter]);

  const handleCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const handleEdit = (adjustment: Adjustment) => {
    setEditing(adjustment);
    setFormOpen(true);
  };

  const handleFormClose = () => {
    setFormOpen(false);
    setEditing(null);
  };

  const handleFormSubmit = async (data: any) => {
    setIsSubmitting(true);
    try {
      if (editing) {
        await updateAdjustment(editing.id, data);
        toast.success('Adjustment updated successfully');
      } else {
        await createAdjustment(data);
        toast.success('Adjustment created successfully');
      }
      handleFormClose();
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to save adjustment');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this adjustment?')) return;

    try {
      await deleteAdjustment(id);
      toast.success('Adjustment deleted successfully');
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to delete adjustment');
    }
  };

  const handleApprove = async (id: string) => {
    if (!confirm('Are you sure you want to approve this adjustment? This will update inventory levels.')) return;

    try {
      await approveAdjustment(id);
      toast.success('Adjustment approved successfully');
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to approve adjustment');
    }
  };

  const totalPages = Math.max(1, Math.ceil((total || 0) / limit));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Adjustments</h1>
          <p className="text-gray-600">Manage inventory adjustments</p>
        </div>
        <button
          onClick={handleCreate}
          className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
        >
          New Adjustment
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white p-4 rounded-lg shadow-sm border border-gray-200">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Search
            </label>
            <input
              type="text"
              value={searchFilter}
              onChange={(e) => setSearchFilter(e.target.value)}
              placeholder="Search by number, location, or notes..."
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Status
            </label>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All Statuses</option>
              {Object.entries(ADJUSTMENT_STATUSES).map(([key, value]) => (
                <option key={key} value={value}>
                  {getAdjustmentStatusLabel(value)}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Reason
            </label>
            <select
              value={reasonFilter}
              onChange={(e) => setReasonFilter(e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All Reasons</option>
              {Object.entries(ADJUSTMENT_REASONS).map(([key, value]) => (
                <option key={key} value={value}>
                  {getAdjustmentReasonLabel(value)}
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-end">
            <button
              onClick={() => {
                setStatusFilter('');
                setReasonFilter('');
                setSearchFilter('');
                setPage(1);
              }}
              className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 transition-colors"
            >
              Clear Filters
            </button>
          </div>
        </div>
      </div>

      {/* Adjustments Table */}
      <div className="bg-white shadow-sm rounded-lg border border-gray-200">
        {loading ? (
          <div className="p-8 text-center">
            <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            <p className="mt-2 text-sm text-gray-600">Loading adjustments...</p>
          </div>
        ) : !adjustments || adjustments.length === 0 ? (
          <div className="p-8 text-center">
            <p className="text-gray-500">No adjustments found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Number
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Location
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Reason
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {(adjustments || []).map((adjustment) => (
                  <tr key={adjustment.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <Link
                        to={`/adjustments/${adjustment.id}`}
                        className="text-blue-600 hover:text-blue-900 font-medium"
                      >
                        {adjustment.number}
                      </Link>
                      <span className="text-xs text-gray-400 ml-2">
                        Status: {adjustment.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {adjustment.location?.name || 'Unknown Location'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {getAdjustmentReasonLabel(adjustment.reason)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getAdjustmentStatusColor(adjustment.status)}`}>
                        {getAdjustmentStatusLabel(adjustment.status)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(adjustment.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                      {adjustment.status === 'DRAFT' && (
                        <>
                          <button
                            onClick={() => handleEdit(adjustment)}
                            className="text-blue-600 hover:text-blue-900"
                            title="Edit"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                          </button>
                          <button
                            onClick={() => handleDelete(adjustment.id)}
                            className="text-red-600 hover:text-red-900"
                            title="Delete"
                          >
                            Delete
                          </button>
                          <button
                            onClick={() => handleApprove(adjustment.id)}
                            className="text-green-600 hover:text-green-900"
                            title="Approve"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                          </button>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {total > limit && (
          <div className="px-6 py-3 border-t border-gray-200 flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Showing {((page - 1) * limit) + 1} to {Math.min(page * limit, total)} of {total} results
            </div>
            <div className="flex space-x-2">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page === 1}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Previous
              </button>
              <span className="px-3 py-1 text-sm text-gray-700">
                Page {page} of {totalPages}
              </span>
              <button
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Form Modal */}
      {formOpen && (
        <AdjustmentForm
          adjustment={editing}
          onClose={handleFormClose}
          onSubmit={handleFormSubmit}
          isSubmitting={isSubmitting}
        />
      )}
    </div>
  );
}