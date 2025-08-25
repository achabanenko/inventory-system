import api from '../lib/api';

export interface Category {
  id: string;
  name: string;
  parent_id?: string | null;
  created_at: string;
  updated_at: string;
}

export interface Item {
  id: string;
  sku: string;
  name: string;
  barcode?: string | null;
  uom: string;
  category_id?: string | null;
  category?: Category | null;
  cost: string; // backend uses decimal; we treat as string
  price: string; // backend uses decimal; we treat as string
  attributes?: Record<string, unknown> | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_pages: number;
  total: number;
}

export interface ListItemsParams {
  q?: string;
  page?: number;
  page_size?: number;
  sort?: string;
}

export interface UpsertItemPayload {
  sku: string;
  name: string;
  barcode?: string | null;
  uom: string;
  category_id?: string | null;
  cost: string | number; // will be coerced to string
  price: string | number; // will be coerced to string
  attributes?: Record<string, unknown> | null;
  is_active?: boolean;
}

export async function listItems(params: ListItemsParams) {
  const res = await api.get<PaginatedResponse<Item>>('/items', { params });
  return res.data;
}

export async function createItem(payload: UpsertItemPayload) {
  const res = await api.post<Item>('/items', normalizeMoney(payload));
  return res.data;
}

export async function updateItem(id: string, payload: UpsertItemPayload) {
  const res = await api.put<Item>(`/items/${id}`, normalizeMoney(payload));
  return res.data;
}

export async function deleteItem(id: string) {
  await api.delete(`/items/${id}`);
}

function normalizeMoney(payload: UpsertItemPayload): UpsertItemPayload {
  return {
    ...payload,
    cost: typeof payload.cost === 'number' ? payload.cost.toFixed(2) : payload.cost,
    price: typeof payload.price === 'number' ? payload.price.toFixed(2) : payload.price,
  };
}


