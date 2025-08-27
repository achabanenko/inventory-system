import api from '../lib/api';

export type TransferStatus = 'DRAFT' | 'IN_TRANSIT' | 'RECEIVED' | 'COMPLETED' | 'CANCELED';

export interface Location {
  id: string;
  name: string;
  code: string;
}

export interface Item {
  id: string;
  sku: string;
  name: string;
}

export interface TransferLine {
  id: string;
  item_id?: string;
  item_identifier: string;
  description: string;
  item?: Item;
  qty: number;
}

export interface Transfer {
  id: string;
  number: string;
  from_location_id: string;
  from_location?: Location;
  to_location_id: string;
  to_location?: Location;
  status: TransferStatus;
  notes: string;
  created_by?: string;
  approved_by?: string;
  shipped_at?: string;
  received_at?: string;
  lines?: TransferLine[];
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

export const listTransfers = async (params?: {
  page?: number;
  page_size?: number;
  q?: string;
  status?: string;
  from_location_id?: string;
  to_location_id?: string;
}) => {
  const res = await api.get<PaginatedResponse<Transfer>>('/transfers', { params });
  return res.data;
};

export interface CreateTransferRequest {
  from_location_id: string;
  to_location_id: string;
  notes?: string;
  lines: {
    item_id: string;
    description: string;
    qty: number;
  }[];
}

export const createTransfer = async (payload: CreateTransferRequest) => {
  const res = await api.post<Transfer>('/transfers', payload);
  return res.data;
};

export const getTransfer = async (id: string) => {
  const res = await api.get<Transfer>(`/transfers/${id}`);
  return res.data;
};

export interface UpdateTransferRequest {
  notes?: string;
  lines: {
    item_id: string;
    description: string;
    qty: number;
  }[];
}

export const updateTransfer = async (id: string, payload: UpdateTransferRequest) => {
  const res = await api.put<Transfer>(`/transfers/${id}`, payload);
  return res.data;
};

export const deleteTransfer = async (id: string) => {
  await api.delete(`/transfers/${id}`);
};

export const approveTransfer = async (id: string) => {
  const res = await api.post<{ message: string }>(`/transfers/${id}/approve`);
  return res.data;
};

export const shipTransfer = async (id: string) => {
  const res = await api.post<{ message: string }>(`/transfers/${id}/ship`);
  return res.data;
};

export const receiveTransfer = async (id: string) => {
  const res = await api.post<{ message: string }>(`/transfers/${id}/receive`);
  return res.data;
};
