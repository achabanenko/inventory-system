import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  listTransfers,
  createTransfer,
  updateTransfer,
  deleteTransfer,
  approveTransfer,
  shipTransfer,
  receiveTransfer,
  type Transfer,
  type CreateTransferRequest
} from '../api/transfers';


import { toast } from 'react-hot-toast';
import {
  CheckCircle,
  Truck,
  Eye,
  Plus,
  Edit,
  ArrowRight,
  Package
} from 'lucide-react';
import TransferForm from '../components/TransferForm';



export default function Transfers() {
  const navigate = useNavigate();
  const [rows, setRows] = useState<Transfer[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Transfer | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [debouncedSearch, setDebouncedSearch] = useState<string>('');



  const reload = async (targetPage?: number) => {
    setIsLoading(true);
    try {
      const p = targetPage ?? page;
      const params: any = { page: p, page_size: 20 };
      if (statusFilter) {
        params.status = statusFilter;
      }
      if (debouncedSearch) {
        params.q = debouncedSearch;
      }
      const data = await listTransfers(params);
      setRows(data.data || []);
      setTotalPages(data.total_pages || 1);
      setTotal(data.total || 0);
      if (targetPage) setPage(targetPage);
    } catch (error) {
      console.error('Failed to load transfers:', error);
      setRows([]);
      setTotalPages(1);
      setTotal(0);
      toast.error('Failed to load transfers');
    } finally {
      setIsLoading(false);
    }
  };


  useEffect(() => { reload(); }, [page]);
  useEffect(() => {
    setPage(1); // Reset to first page when filter changes
    reload(1);
  }, [statusFilter]);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchTerm);
      setPage(1); // Reset to first page when searching
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  // Reload when search changes
  useEffect(() => {
    reload(1);
  }, [debouncedSearch]);



  const onDelete = async (transfer: Transfer) => {
    const statusText = transfer.status === 'DRAFT' ? 'draft' :
                      transfer.status === 'IN_TRANSIT' ? 'in-transit' :
                      transfer.status === 'RECEIVED' ? 'received' :
                      transfer.status === 'COMPLETED' ? 'completed' : 'canceled';

    const message = `Are you sure you want to delete this ${statusText} transfer (${transfer.number})? This action cannot be undone.`;

    if (!confirm(message)) return;

    try {
      await deleteTransfer(transfer.id);
      toast.success('Transfer deleted successfully');
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to delete transfer');
    }
  };

  const onApprove = async (transfer: Transfer) => {
    setActionLoading(`approve-${transfer.id}`);
    try {
      await approveTransfer(transfer.id);
      toast.success('Transfer approved successfully');
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to approve transfer');
    } finally {
      setActionLoading(null);
    }
  };

  const onShip = async (transfer: Transfer) => {
    setActionLoading(`ship-${transfer.id}`);
    try {
      await shipTransfer(transfer.id);
      toast.success('Transfer shipped successfully');
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to ship transfer');
    } finally {
      setActionLoading(null);
    }
  };

  const onReceive = async (transfer: Transfer) => {
    setActionLoading(`receive-${transfer.id}`);
    try {
      await receiveTransfer(transfer.id);
      toast.success('Transfer received successfully');
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to receive transfer');
    } finally {
      setActionLoading(null);
    }
  };

  const handleCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const handleEdit = (transfer: Transfer) => {
    setEditing(transfer);
    setFormOpen(true);
  };

  const handleFormClose = () => {
    setFormOpen(false);
    setEditing(null);
  };

  const handleFormSubmit = async (data: CreateTransferRequest) => {
    setIsSubmitting(true);
    try {
      if (editing) {
        await updateTransfer(editing.id, data);
        toast.success('Transfer updated successfully');
      } else {
        await createTransfer(data);
        toast.success('Transfer created successfully');
      }
      handleFormClose();
      reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to save transfer');
    } finally {
      setIsSubmitting(false);
    }
  };



  const getStatusColor = (status: string) => {
    switch (status) {
      case 'DRAFT':
        return 'bg-gray-100 text-gray-800';
      case 'IN_TRANSIT':
        return 'bg-blue-100 text-blue-800';
      case 'RECEIVED':
        return 'bg-yellow-100 text-yellow-800';
      case 'COMPLETED':
        return 'bg-green-100 text-green-800';
      case 'CANCELED':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div>
      <div className="mb-8 flex justify-between items-start">
        <div>
        <h1 className="text-2xl font-bold text-gray-900">Transfers</h1>
          <p className="mt-1 text-sm text-gray-500">Move items between locations and manage transfer workflows</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleCreate}
            className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <Plus className="h-4 w-4 mr-2" />
            New Transfer
          </button>
        </div>
      </div>

      {/* Search and Status Filter */}
      <div className="mb-4 flex items-center gap-4">
        <div className="flex items-center gap-2">
          <label htmlFor="search" className="text-sm font-medium text-gray-700">
            Search:
          </label>
          <input
            id="search"
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search by transfer number..."
            className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
          />
          {searchTerm && (
            <button
              onClick={() => setSearchTerm('')}
              className="text-sm text-gray-500 hover:text-gray-700 underline"
            >
              Clear
            </button>
          )}
        </div>
        <div className="flex items-center gap-2">
          <label htmlFor="status-filter" className="text-sm font-medium text-gray-700">
            Status:
          </label>
          <select
            id="status-filter"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">All Statuses</option>
            <option value="DRAFT">Draft</option>
            <option value="IN_TRANSIT">In Transit</option>
            <option value="RECEIVED">Received</option>
            <option value="COMPLETED">Completed</option>
            <option value="CANCELED">Canceled</option>
          </select>
        </div>
        {(statusFilter || searchTerm) && (
          <button
            onClick={() => { setStatusFilter(''); setSearchTerm(''); }}
            className="text-sm text-gray-500 hover:text-gray-700 underline"
          >
            Clear All Filters
          </button>
        )}
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Number</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">From → To</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Items</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading ? (
                <tr><td colSpan={6} className="px-6 py-4 text-center text-sm text-gray-500">Loading...</td></tr>
              ) : !rows || rows.length === 0 ? (
                <tr><td colSpan={6} className="px-6 py-4 text-center text-sm text-gray-500">
                  {debouncedSearch || statusFilter ?
                    `No transfers found matching your criteria` :
                    'No transfers yet'
                  }
                </td></tr>
              ) : (
                rows.map((transfer) => (
                  <tr key={transfer.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <button
                        onClick={() => navigate(`/transfers/${transfer.id}`)}
                        className="text-blue-600 hover:text-blue-900 font-medium"
                      >
                        {transfer.number}
                      </button>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="flex items-center gap-1">
                        <span className="text-sm">
                          {transfer.from_location?.name || transfer.from_location?.code || 'Unknown'}
                        </span>
                        <ArrowRight className="h-4 w-4 text-gray-400" />
                        <span className="text-sm">
                          {transfer.to_location?.name || transfer.to_location?.code || 'Unknown'}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(transfer.status)}`}>
                        {transfer.status.replace('_', ' ')}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="flex items-center gap-1">
                        <Package className="h-4 w-4 text-gray-400" />
                        <span>{transfer.lines?.length || 0} items</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {new Date(transfer.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => navigate(`/transfers/${transfer.id}`)}
                          className="text-gray-600 hover:text-gray-900"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        {transfer.status === 'DRAFT' && (
                          <>
                            <button
                              onClick={() => onApprove(transfer)}
                              disabled={actionLoading === `approve-${transfer.id}`}
                              className="text-blue-600 hover:text-blue-900 disabled:opacity-50"
                              title="Approve Transfer"
                            >
                              <CheckCircle className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => handleEdit(transfer)}
                              className="text-green-600 hover:text-green-900"
                              title="Edit Transfer"
                            >
                              <Edit className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => onDelete(transfer)}
                              className="text-red-600 hover:text-red-900"
                              title="Delete Transfer"
                            >
                              Delete
                            </button>
                          </>
                        )}
                        {/* Debug: Show transfer status */}
                        <span className="text-xs text-gray-400 ml-2">
                          Status: {transfer.status}
                        </span>
                        {transfer.status === 'IN_TRANSIT' && (
                          <button
                            onClick={() => onShip(transfer)}
                            disabled={actionLoading === `ship-${transfer.id}`}
                            className="text-blue-600 hover:text-blue-900 disabled:opacity-50"
                            title="Mark as Shipped"
                          >
                            <Truck className="h-4 w-4" />
                          </button>
                        )}
                        {transfer.status === 'RECEIVED' && (
                          <button
                            onClick={() => onReceive(transfer)}
                            disabled={actionLoading === `receive-${transfer.id}`}
                            className="text-green-600 hover:text-green-900 disabled:opacity-50"
                            title="Mark as Received"
                          >
                            <CheckCircle className="h-4 w-4" />
                          </button>
                        )}
                        {transfer.status === 'COMPLETED' && (
                          <span className="text-green-600 text-xs font-medium">✓ Completed</span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {totalPages > 1 && (
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
            <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">Page <span className="font-medium">{page}</span> of <span className="font-medium">{totalPages}</span> — Total <span className="font-medium">{total}</span></p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page <= 1}
                  className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page >= totalPages}
                  className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      <TransferForm
        isOpen={formOpen}
        onClose={handleFormClose}
        onSubmit={handleFormSubmit}
        initialData={editing}
        isSubmitting={isSubmitting}
      />
    </div>
  );
}