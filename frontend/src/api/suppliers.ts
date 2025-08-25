import api from '../lib/api';

export interface Supplier {
  id: string;
  code: string;
  name: string;
  contact?: any;
  is_active: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_pages: number;
  total: number;
}

export interface ListSuppliersParams {
  q?: string;
  page?: number;
  page_size?: number;
  is_active?: boolean;
}

export const listSuppliers = async (params?: ListSuppliersParams): Promise<PaginatedResponse<Supplier>> => {
  const response = await api.get('/suppliers', { params });
  return response.data;
};

export const getSupplier = async (id: string): Promise<Supplier> => {
  const response = await api.get(`/suppliers/${id}`);
  return response.data;
};

export interface UpsertSupplierPayload {
  code: string;
  name: string;
  contact?: Record<string, unknown> | null;
  is_active?: boolean;
}

export const createSupplier = async (payload: UpsertSupplierPayload): Promise<Supplier> => {
  const response = await api.post('/suppliers', payload);
  return response.data;
};

export const updateSupplier = async (id: string, payload: UpsertSupplierPayload): Promise<Supplier> => {
  const response = await api.put(`/suppliers/${id}`, payload);
  return response.data;
};

export const deleteSupplier = async (id: string): Promise<void> => {
  await api.delete(`/suppliers/${id}`);
};
