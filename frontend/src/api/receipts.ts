import api from '../lib/api';

export type ReceiptStatus = 'DRAFT' | 'APPROVED' | 'POSTED' | 'CLOSED' | 'CANCELED';

export interface Supplier {
  id: string;
  name: string;
}

export interface Location {
  id: string;
  code: string;
  name: string;
}

export interface GoodsReceipt {
  id: string;
  number: string;
  supplier_id?: string;
  supplier?: Supplier;
  location_id?: string;
  location?: Location;
  status: ReceiptStatus;
  reference?: string;
  notes?: string;
  created_by?: string;
  approved_by?: string;
  posted_by?: string;
  approved_at?: string;
  posted_at?: string;
  lines?: GoodsReceiptLine[];
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

export const listReceipts = async (params?: { page?: number; page_size?: number }) => {
  const res = await api.get<PaginatedResponse<GoodsReceipt>>('/receipts', { params });
  return res.data;
};

export const createReceipt = async (payload: { supplier_id?: string; location_id?: string; reference?: string; notes?: string }) => {
  const res = await api.post<GoodsReceipt>('/receipts', payload);
  return res.data;
};

export const updateReceipt = async (id: string, payload: Partial<{ supplier_id: string; location_id: string; status: ReceiptStatus; reference: string; notes: string }>) => {
  const res = await api.put<GoodsReceipt>(`/receipts/${id}`, payload);
  return res.data;
};

export const deleteReceipt = async (id: string) => {
  await api.delete(`/receipts/${id}`);
};

export interface Item {
  id: string;
  sku: string;
  name: string;
}

export interface GoodsReceiptLine {
  id: string;
  receipt_id: string;
  item_id: string;
  item?: Item;
  qty: number;
  unit_cost: number | string;
  line_total: number | string;
  created_at: string;
  updated_at: string;
}

export const listReceiptLines = async (id: string) => {
  const res = await api.get<{ data: GoodsReceiptLine[] }>(`/receipts/${id}/lines`);
  return res.data;
};

export const addReceiptLine = async (id: string, payload: { item_id: string; qty: number; unit_cost: string }) => {
  const res = await api.post<GoodsReceiptLine>(`/receipts/${id}/lines`, payload);
  return res.data;
};

export const updateReceiptLine = async (id: string, lineId: string, payload: Partial<{ qty: number; unit_cost: string }>) => {
  const res = await api.put<GoodsReceiptLine>(`/receipts/${id}/lines/${lineId}`, payload);
  return res.data;
};

export const deleteReceiptLine = async (id: string, lineId: string) => {
  await api.delete(`/receipts/${id}/lines/${lineId}`);
};

export const createReceiptFromPO = async (payload: { purchase_order_id: string; location_id: string; reference?: string; notes?: string }) => {
  const res = await api.post<GoodsReceipt>('/receipts/from-po', payload);
  return res.data;
};

export const getReceipt = async (id: string) => {
  const res = await api.get<GoodsReceipt>(`/receipts/${id}`);
  return res.data;
};

export const approveReceipt = async (id: string) => {
  const res = await api.post(`/receipts/${id}/approve`);
  return res.data;
};

export const postReceipt = async (id: string) => {
  const res = await api.post(`/receipts/${id}/post`);
  return res.data;
};

export const closeReceipt = async (id: string) => {
  const res = await api.post(`/receipts/${id}/close`);
  return res.data;
};


