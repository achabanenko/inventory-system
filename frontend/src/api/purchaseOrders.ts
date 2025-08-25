import api from '../lib/api';

export interface Supplier {
  id: string;
  name: string;
}

export interface Item {
  id: string;
  sku: string;
  name: string;
}

export interface PurchaseOrderLine {
  id: string;
  item_id: string;
  item?: Item;
  qty_ordered: number;
  qty_received: number;
  unit_cost: string; // Use string for decimal values
  tax?: any;
  line_total: string;
  created_at: string;
  updated_at: string;
}

export interface PurchaseOrder {
  id: string;
  number: string;
  status: 'DRAFT' | 'APPROVED' | 'PARTIAL' | 'RECEIVED' | 'CLOSED' | 'CANCELED';
  supplier_id: string;
  supplier?: Supplier;
  created_by: string;
  approved_by?: string;
  expected_at?: string;
  approved_at?: string;
  notes?: string;
  lines?: PurchaseOrderLine[];
  total: string;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_pages: number;
  total: number;
}

export interface ListPurchaseOrdersParams {
  q?: string;
  page?: number;
  page_size?: number;
  sort?: string;
  status?: string;
  supplier_id?: string;
}

export interface CreatePurchaseOrderLine {
  item_id: string;
  qty_ordered: number;
  unit_cost: string;
  tax?: any;
}

export interface CreatePurchaseOrderRequest {
  supplier_id: string;
  expected_at?: string;
  notes?: string;
  lines: CreatePurchaseOrderLine[];
}

export interface UpdatePurchaseOrderLine {
  id?: string;
  item_id: string;
  qty_ordered: number;
  unit_cost: string;
  tax?: any;
}

export interface UpdatePurchaseOrderRequest {
  supplier_id: string;
  expected_at?: string;
  notes?: string;
  lines: UpdatePurchaseOrderLine[];
}

export interface ReceiveLine {
  line_id: string;
  qty_received: number;
}

export interface ReceiveItemsRequest {
  lines: ReceiveLine[];
}

// API functions
export const listPurchaseOrders = async (params?: ListPurchaseOrdersParams): Promise<PaginatedResponse<PurchaseOrder>> => {
  const response = await api.get('/purchase-orders', { params });
  return response.data;
};

export const getPurchaseOrder = async (id: string): Promise<PurchaseOrder> => {
  const response = await api.get(`/purchase-orders/${id}`);
  return response.data;
};

export const createPurchaseOrder = async (data: CreatePurchaseOrderRequest): Promise<PurchaseOrder> => {
  const response = await api.post('/purchase-orders', data);
  return response.data;
};

export const updatePurchaseOrder = async (id: string, data: UpdatePurchaseOrderRequest): Promise<PurchaseOrder> => {
  const response = await api.put(`/purchase-orders/${id}`, data);
  return response.data;
};

export const deletePurchaseOrder = async (id: string): Promise<void> => {
  await api.delete(`/purchase-orders/${id}`);
};

export const approvePurchaseOrder = async (id: string): Promise<{ message: string }> => {
  const response = await api.post(`/purchase-orders/${id}/approve`);
  return response.data;
};

export const receivePurchaseOrder = async (id: string, data: ReceiveItemsRequest): Promise<{ message: string; status: string }> => {
  const response = await api.post(`/purchase-orders/${id}/receive`, data);
  return response.data;
};

export const closePurchaseOrder = async (id: string): Promise<{ message: string }> => {
  const response = await api.post(`/purchase-orders/${id}/close`);
  return response.data;
};
