import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { listReceipts, createReceipt, updateReceipt, deleteReceipt, approveReceipt, postReceipt, closeReceipt, type GoodsReceipt } from '../api/receipts';
import { listSuppliers, type Supplier } from '../api/suppliers';
import { listLocations, type Location } from '../api/locations';
import { toast } from 'react-hot-toast';
import { CheckCircle, Truck, XCircle, Eye, Plus, Edit, FileText } from 'lucide-react';
import GoodsReceiptForm from '../components/GoodsReceiptForm';
import ReceiptLineEditor from '../components/ReceiptLineEditor';
import type { CreateGoodsReceiptRequest } from '../api/receipts';

const STATUS_COLORS = {
  DRAFT: 'bg-gray-100 text-gray-800',
  APPROVED: 'bg-blue-100 text-blue-800',
  POSTED: 'bg-green-100 text-green-800',
  CLOSED: 'bg-gray-100 text-gray-800',
  CANCELED: 'bg-red-100 text-red-800',
};

export default function Receipts() {
  const navigate = useNavigate();
  const [rows, setRows] = useState<GoodsReceipt[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [formOpen, setFormOpen] = useState(false);
  const [fromPOFormOpen, setFromPOFormOpen] = useState(false);
  const [editing, setEditing] = useState<GoodsReceipt | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [lineEditorOpen, setLineEditorOpen] = useState(false);
  const [selectedReceipt, setSelectedReceipt] = useState<GoodsReceipt | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [debouncedSearch, setDebouncedSearch] = useState<string>('');

  // State variables for suppliers and locations (used in table display)
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);

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
      const data = await listReceipts(params);
      setRows(data.data || []);
      setTotalPages(data.total_pages || 1);
      setTotal(data.total || 0);
      if (targetPage) setPage(targetPage);
    } catch (error) {
      console.error('Failed to load receipts:', error);
      setRows([]);
      setTotalPages(1);
      setTotal(0);
      toast.error('Failed to load receipts');
    } finally { 
      setIsLoading(false); 
    }
  };

  useEffect(() => { (async () => {
    try {
      const s = await listSuppliers({ page_size: 200 }); 
      setSuppliers(s.data || []);
      const l = await listLocations({ page_size: 200 }); 
      setLocations(l.data || []);
    } catch (error) {
      console.error('Failed to load suppliers/locations:', error);
      setSuppliers([]);
      setLocations([]);
    }
  })(); }, []);
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

  // Legacy functions removed - now using handleCreate and handleEdit
  const onDelete = async (r: GoodsReceipt) => { 
    const statusText = r.status === 'DRAFT' ? 'draft' : 
                      r.status === 'APPROVED' ? 'approved' : 
                      r.status === 'POSTED' ? 'posted' : 
                      r.status === 'CLOSED' ? 'closed' : 
                      r.status === 'CANCELED' ? 'canceled' : 'receipt';
    
    const message = `Are you sure you want to delete this ${statusText} receipt (${r.number})? This action cannot be undone.`;
    
    if (!confirm(message)) return; 
    
    try {
      await deleteReceipt(r.id); 
      toast.success('Receipt deleted successfully'); 
      reload(); 
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to delete receipt');
    }
  };
  
  const onApprove = async (r: GoodsReceipt) => {
    setActionLoading(`approve-${r.id}`);
    try {
      await approveReceipt(r.id);
      toast.success('Receipt approved');
      reload();
    } catch {
      toast.error('Failed to approve receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const onPost = async (r: GoodsReceipt) => {
    setActionLoading(`post-${r.id}`);
    try {
      await postReceipt(r.id);
      toast.success('Receipt posted to inventory');
      reload();
    } catch {
      toast.error('Failed to post receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const onClose = async (r: GoodsReceipt) => {
    setActionLoading(`close-${r.id}`);
    try {
      await closeReceipt(r.id);
      toast.success('Receipt closed');
      reload();
    } catch {
      toast.error('Failed to close receipt');
    } finally {
      setActionLoading(null);
    }
  };



  // New form submission handlers
  const handleCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const handleEdit = (receipt: GoodsReceipt) => {
    setEditing(receipt);
    setFormOpen(true);
  };

  const handleCreateFromPO = () => {
    setEditing(null);
    setFromPOFormOpen(true);
  };

  const handleFormSubmit = async (data: CreateGoodsReceiptRequest) => {
    setIsSubmitting(true);
    try {
      if (editing) {
        await updateReceipt(editing.id, data);
        toast.success('Receipt updated successfully');
      } else {
        await createReceipt(data);
        toast.success('Receipt created successfully');
      }
      setFormOpen(false);
      setFromPOFormOpen(false);
      setEditing(null);
      await reload();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Operation failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleFormClose = () => {
    setFormOpen(false);
    setFromPOFormOpen(false);
    setEditing(null);
  };

  const handleEditLines = (receipt: GoodsReceipt) => {
    setSelectedReceipt(receipt);
    setLineEditorOpen(true);
  };

  const handleLineEditorClose = () => {
    setLineEditorOpen(false);
    setSelectedReceipt(null);
    // Reload receipts to show updated totals
    reload();
  };

  // Legacy form handler removed - now using handleFormSubmit

  // Removed unused state variables for old line management interface

  // Removed unused useEffect hooks for old line management interface

  return (
    <div>
      <div className="mb-8 flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Goods Receipts</h1>
          <p className="mt-1 text-sm text-gray-500">Create, edit, and manage goods receipt documents</p>
        </div>
        <div className="flex gap-2">
          <button 
            onClick={handleCreate} 
            className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <Plus className="h-4 w-4 mr-2" />
            New Receipt
          </button>
          <button 
            onClick={handleCreateFromPO} 
            className="inline-flex items-center px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500"
          >
            <FileText className="h-4 w-4 mr-2" />
            From Purchase Order
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
            placeholder="Search by receipt number..."
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
            <option value="APPROVED">Approved</option>
            <option value="POSTED">Posted</option>
            <option value="CLOSED">Closed</option>
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
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Supplier</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Location</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Total</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading ? (
                <tr><td colSpan={6} className="px-6 py-4 text-center text-sm text-gray-500">Loading...</td></tr>
              ) : !rows || rows.length === 0 ? (
                <tr><td colSpan={6} className="px-6 py-4 text-center text-sm text-gray-500">
                  {debouncedSearch || statusFilter ? 
                    `No receipts found matching your criteria` : 
                    'No receipts yet'
                  }
                </td></tr>
              ) : (
                rows.map((r) => (
                  <tr key={r.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <button
                        onClick={() => navigate(`/receipts/${r.id}`)}
                        className="text-blue-600 hover:text-blue-900 font-medium"
                      >
                        {r.number}
                      </button>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {r.supplier?.name || suppliers.find(s => s.id === r.supplier_id)?.name || '—'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {r.location ? `${r.location.name} (${r.location.code})` : locations.find(l => l.id === r.location_id)?.name || '—'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${STATUS_COLORS[r.status] || 'bg-gray-100 text-gray-800'}`}>
                        {r.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium text-gray-900">
                      ${(parseFloat(r.total?.toString() || '0') || 0).toFixed(2)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => navigate(`/receipts/${r.id}`)}
                          className="text-gray-600 hover:text-gray-900"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        {r.status === 'DRAFT' && (
                          <>
                            <button
                              onClick={() => onApprove(r)}
                              disabled={actionLoading === `approve-${r.id}`}
                              className="text-blue-600 hover:text-blue-900 disabled:opacity-50"
                            >
                              <CheckCircle className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => handleEdit(r)}
                              className="text-green-600 hover:text-green-900"
                              title="Edit Receipt"
                            >
                              <Edit className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => handleEditLines(r)}
                              className="text-blue-600 hover:text-blue-900"
                              title="Edit Lines"
                            >
                              <FileText className="h-4 w-4" />
                            </button>
                          </>
                        )}
                        {/* Delete button - available for all statuses */}
                        <button 
                          onClick={() => onDelete(r)} 
                          className="text-red-600 hover:text-red-900"
                          title="Delete Receipt"
                        >
                          Delete
                        </button>
                        {r.status === 'APPROVED' && (
                          <button
                            onClick={() => onPost(r)}
                            disabled={actionLoading === `post-${r.id}`}
                            className="text-green-600 hover:text-green-900 disabled:opacity-50"
                          >
                            <Truck className="h-4 w-4" />
                          </button>
                        )}
                        {r.status === 'POSTED' && (
                          <button
                            onClick={() => onClose(r)}
                            disabled={actionLoading === `close-${r.id}`}
                            className="text-purple-600 hover:text-purple-900 disabled:opacity-50"
                          >
                            <XCircle className="h-4 w-4" />
                          </button>
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
                <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page <= 1} className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">Previous</button>
                <button onClick={() => setPage(Math.min(totalPages, page + 1))} disabled={page >= totalPages} className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">Next</button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Old form modal removed - now using GoodsReceiptForm component */}

      {/* Receipt Line Editor */}
      {lineEditorOpen && selectedReceipt && (
        <ReceiptLineEditor
          isOpen={lineEditorOpen}
          onClose={handleLineEditorClose}
          receiptId={selectedReceipt.id}
          receiptNumber={selectedReceipt.number}
          receiptStatus={selectedReceipt.status}
        />
      )}

      {/* New Enhanced Forms */}
      <GoodsReceiptForm
        isOpen={formOpen}
        onClose={handleFormClose}
        onSubmit={handleFormSubmit}
        initialData={editing}
        isSubmitting={isSubmitting}
        createFromPO={false}
      />

      <GoodsReceiptForm
        isOpen={fromPOFormOpen}
        onClose={handleFormClose}
        onSubmit={handleFormSubmit}
        isSubmitting={isSubmitting}
        createFromPO={true}
      />
    </div>
  );
}


