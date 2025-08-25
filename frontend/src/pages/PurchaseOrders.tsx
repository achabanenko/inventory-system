import { useState, useEffect, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  listPurchaseOrders, 
  getPurchaseOrder,
  createPurchaseOrder,
  updatePurchaseOrder,
  deletePurchaseOrder, 
  approvePurchaseOrder, 
  receivePurchaseOrder,
  closePurchaseOrder,
  type PurchaseOrder,
  type ListPurchaseOrdersParams,
  type CreatePurchaseOrderRequest,
  type UpdatePurchaseOrderRequest,
  type ReceiveItemsRequest
} from '../api/purchaseOrders';
import { toast } from 'react-hot-toast';
import PurchaseOrderForm from '../components/PurchaseOrderForm';
import PurchaseOrderDetails from '../components/PurchaseOrderDetails';
import ReceiveItemsForm from '../components/ReceiveItemsForm';

const STATUS_COLORS = {
  DRAFT: 'bg-gray-100 text-gray-800',
  APPROVED: 'bg-blue-100 text-blue-800',
  PARTIAL: 'bg-yellow-100 text-yellow-800',
  RECEIVED: 'bg-green-100 text-green-800',
  CLOSED: 'bg-gray-100 text-gray-800',
  CANCELED: 'bg-red-100 text-red-800',
};

export default function PurchaseOrders() {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [page, setPage] = useState(1);
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [formOpen, setFormOpen] = useState(false);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [receiveOpen, setReceiveOpen] = useState(false);
  const [selectedPO, setSelectedPO] = useState<PurchaseOrder | null>(null);
  const [selectedPOId, setSelectedPOId] = useState<string | null>(null);
  const [editingPO, setEditingPO] = useState<PurchaseOrder | null>(null);
  const queryClient = useQueryClient();

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchTerm);
      setPage(1); // Reset to first page when searching
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  const queryParams: ListPurchaseOrdersParams = useMemo(() => ({
    q: debouncedSearch || undefined,
    status: statusFilter || undefined,
    page,
    page_size: 20,
    sort: 'created_at DESC',
  }), [debouncedSearch, statusFilter, page]);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['purchaseOrders', queryParams],
    queryFn: () => listPurchaseOrders(queryParams),
  });

  const createMutation = useMutation({
    mutationFn: createPurchaseOrder,
    onSuccess: () => {
      toast.success('Purchase order created successfully');
      queryClient.invalidateQueries({ queryKey: ['purchaseOrders'] });
      setFormOpen(false);
      setEditingPO(null);
    },
    onError: (error: any) => {
      toast.error(error?.response?.data?.message || 'Failed to create purchase order');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdatePurchaseOrderRequest }) => 
      updatePurchaseOrder(id, data),
    onSuccess: (_, { id }) => {
      toast.success('Purchase order updated successfully');
      // Invalidate both the list and the specific PO details
      queryClient.invalidateQueries({ queryKey: ['purchaseOrders'] });
      queryClient.invalidateQueries({ queryKey: ['purchaseOrder', id] });
      setFormOpen(false);
      setEditingPO(null);
    },
    onError: (error: any) => {
      toast.error(error?.response?.data?.message || 'Failed to update purchase order');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deletePurchaseOrder,
    onSuccess: () => {
      toast.success('Purchase order deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['purchaseOrders'] });
    },
    onError: (error: any) => {
      toast.error(error?.response?.data?.message || 'Failed to delete purchase order');
    },
  });

  const approveMutation = useMutation({
    mutationFn: approvePurchaseOrder,
    onSuccess: () => {
      toast.success('Purchase order approved successfully');
      queryClient.invalidateQueries({ queryKey: ['purchaseOrders'] });
    },
    onError: (error: any) => {
      toast.error(error?.response?.data?.message || 'Failed to approve purchase order');
    },
  });

  const receiveMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: ReceiveItemsRequest }) => 
      receivePurchaseOrder(id, data),
    onSuccess: () => {
      toast.success('Items received successfully');
      queryClient.invalidateQueries({ queryKey: ['purchaseOrders'] });
      queryClient.invalidateQueries({ queryKey: ['purchaseOrder'] });
      setReceiveOpen(false);
      setSelectedPO(null);
    },
    onError: (error: any) => {
      toast.error(error?.response?.data?.message || 'Failed to receive items');
    },
  });

  const closeMutation = useMutation({
    mutationFn: closePurchaseOrder,
    onSuccess: () => {
      toast.success('Purchase order closed successfully');
      queryClient.invalidateQueries({ queryKey: ['purchaseOrders'] });
    },
    onError: (error: any) => {
      toast.error(error?.response?.data?.message || 'Failed to close purchase order');
    },
  });

  const handleCreate = () => {
    setEditingPO(null);
    setFormOpen(true);
  };

  const handleEdit = async (po: PurchaseOrder) => {
    try {
      // Fetch full PO details with lines before editing
      const fullPO = await queryClient.fetchQuery({
        queryKey: ['purchaseOrder', po.id],
        queryFn: () => getPurchaseOrder(po.id),
      });
      setEditingPO(fullPO);
      setFormOpen(true);
    } catch (error) {
      toast.error('Failed to load purchase order details for editing');
    }
  };

  const handleView = (id: string) => {
    setSelectedPOId(id);
    setDetailsOpen(true);
  };

  const handleReceive = (po: PurchaseOrder) => {
    setSelectedPO(po);
    setReceiveOpen(true);
  };

  const handleFormSubmit = (data: CreatePurchaseOrderRequest) => {
    if (editingPO) {
      updateMutation.mutate({ 
        id: editingPO.id, 
        data: data as UpdatePurchaseOrderRequest 
      });
    } else {
      createMutation.mutate(data);
    }
  };

  const handleReceiveSubmit = (data: ReceiveItemsRequest) => {
    if (selectedPO) {
      receiveMutation.mutate({ id: selectedPO.id, data });
    }
  };

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this purchase order?')) {
      deleteMutation.mutate(id);
    }
  };

  const handleApprove = async (id: string) => {
    if (window.confirm('Are you sure you want to approve this purchase order?')) {
      approveMutation.mutate(id);
    }
  };

  const handleClose = async (id: string) => {
    if (window.confirm('Are you sure you want to close this purchase order?')) {
      closeMutation.mutate(id);
    }
  };

  const clearSearch = () => {
    setSearchTerm('');
  };

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-600">Error loading purchase orders: {(error as any)?.message}</p>
        <button 
          onClick={() => refetch()} 
          className="mt-4 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Purchase Orders</h1>
        <p className="mt-1 text-sm text-gray-500">
          Manage purchase orders and receiving
        </p>
      </div>

      {/* Search and Filters */}
      <div className="mb-6 bg-white shadow rounded-lg p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1 relative">
            <input
              type="text"
              placeholder="Search purchase orders..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-3 pr-10 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
            />
            {searchTerm && (
              <button
                onClick={clearSearch}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
              >
                ✕
              </button>
            )}
          </div>
          <div className="sm:w-48">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All Statuses</option>
              <option value="DRAFT">Draft</option>
              <option value="APPROVED">Approved</option>
              <option value="PARTIAL">Partial</option>
              <option value="RECEIVED">Received</option>
              <option value="CLOSED">Closed</option>
              <option value="CANCELED">Canceled</option>
            </select>
          </div>
          <button 
            onClick={handleCreate}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          >
            Create Purchase Order
          </button>
        </div>
      </div>

      {/* Purchase Orders Table */}
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Number
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Supplier
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Expected Date
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="px-6 py-4 text-center text-sm text-gray-500">
                    Loading purchase orders...
                  </td>
                </tr>
              ) : (data?.data?.length ?? 0) === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-4 text-center text-sm text-gray-500">
                    {debouncedSearch || statusFilter ? 'No purchase orders found matching your criteria' : 'No purchase orders yet'}
                  </td>
                </tr>
              ) : (
                data?.data?.map((po: PurchaseOrder) => (
                  <tr key={po.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {po.number}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {po.supplier?.name || 'Unknown Supplier'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${STATUS_COLORS[po.status]}`}>
                        {po.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      ${parseFloat(po.total).toFixed(2)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {po.expected_at ? new Date(po.expected_at).toLocaleDateString() : 'Not set'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {new Date(po.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex justify-end gap-2">
                        <button 
                          onClick={() => handleView(po.id)}
                          className="text-blue-600 hover:text-blue-900"
                        >
                          View
                        </button>
                        {po.status === 'DRAFT' && (
                          <>
                            <button 
                              onClick={() => handleEdit(po)}
                              className="text-green-600 hover:text-green-900"
                            >
                              Edit
                            </button>
                            <button 
                              onClick={() => handleApprove(po.id)}
                              className="text-purple-600 hover:text-purple-900"
                              disabled={approveMutation.isPending}
                            >
                              Approve
                            </button>
                            <button 
                              onClick={() => handleDelete(po.id)}
                              className="text-red-600 hover:text-red-900"
                              disabled={deleteMutation.isPending}
                            >
                              Delete
                            </button>
                          </>
                        )}
                        {(po.status === 'APPROVED' || po.status === 'PARTIAL') && (
                          <button 
                            onClick={() => handleReceive(po)}
                            className="text-orange-600 hover:text-orange-900"
                          >
                            Receive
                          </button>
                        )}
                        {(po.status === 'RECEIVED' || po.status === 'PARTIAL') && (
                          <button 
                            onClick={() => handleClose(po.id)}
                            className="text-gray-600 hover:text-gray-900"
                            disabled={closeMutation.isPending}
                          >
                            Close
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

        {/* Pagination */}
        {data && (data.total_pages ?? 0) > 1 && (
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
            <div className="flex-1 flex justify-between sm:hidden">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              <button
                onClick={() => setPage(Math.min(data.total_pages, page + 1))}
                disabled={page >= data.total_pages}
                className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
            <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">
                  Showing{' '}
                  <span className="font-medium">{(page - 1) * 20 + 1}</span>
                  {' '}to{' '}
                  <span className="font-medium">
                    {Math.min(page * 20, data.total ?? 0)}
                  </span>
                  {' '}of{' '}
                  <span className="font-medium">{data.total ?? 0}</span>
                  {' '}results
                </p>
              </div>
              <div>
                <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
                  <button
                    onClick={() => setPage(Math.max(1, page - 1))}
                    disabled={page <= 1}
                    className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Previous
                  </button>
                  {Array.from({ length: Math.min(5, data.total_pages ?? 0) }, (_, i) => {
                    const totalPages = data.total_pages ?? 0;
                    const pageNum = Math.max(1, Math.min(totalPages - 4, page - 2)) + i;
                    return (
                      <button
                        key={pageNum}
                        onClick={() => setPage(pageNum)}
                        className={`relative inline-flex items-center px-4 py-2 border text-sm font-medium ${
                          pageNum === page
                            ? 'z-10 bg-blue-50 border-blue-500 text-blue-600'
                            : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                        }`}
                      >
                        {pageNum}
                      </button>
                    );
                  })}
                  <button
                    onClick={() => setPage(Math.min(data.total_pages, page + 1))}
                    disabled={page >= data.total_pages}
                    className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Next
                  </button>
                </nav>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Form and Modal Components */}
      <PurchaseOrderForm
        key={editingPO?.id || 'new'} // Force re-initialization when editing different POs
        isOpen={formOpen}
        onClose={() => {
          setFormOpen(false);
          setEditingPO(null);
        }}
        onSubmit={handleFormSubmit}
        initialData={editingPO || undefined}
        isSubmitting={createMutation.isPending || updateMutation.isPending}
      />

      <PurchaseOrderDetails
        isOpen={detailsOpen}
        onClose={() => {
          setDetailsOpen(false);
          setSelectedPOId(null);
        }}
        purchaseOrderId={selectedPOId}
        onEdit={(po) => {
          setDetailsOpen(false);
          handleEdit(po);
        }}
        onReceive={(po) => {
          setDetailsOpen(false);
          handleReceive(po);
        }}
      />

      <ReceiveItemsForm
        isOpen={receiveOpen}
        onClose={() => {
          setReceiveOpen(false);
          setSelectedPO(null);
        }}
        onSubmit={handleReceiveSubmit}
        purchaseOrder={selectedPO}
        isSubmitting={receiveMutation.isPending}
      />
    </div>
  );
}