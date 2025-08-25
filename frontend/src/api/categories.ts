import api from '../lib/api';

export interface Category {
  id: string;
  name: string;
  parent_id?: string | null;
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

export interface ListCategoriesParams {
  q?: string;
  page?: number;
  page_size?: number;
  sort?: string;
}

export interface UpsertCategoryPayload {
  name: string;
  parent_id?: string | null;
}

export async function listCategories(params: ListCategoriesParams = {}) {
  const res = await api.get<PaginatedResponse<Category>>('/categories', { params });
  return res.data;
}

export async function createCategory(payload: UpsertCategoryPayload) {
  const res = await api.post<Category>('/categories', payload);
  return res.data;
}

export async function updateCategory(id: string, payload: UpsertCategoryPayload) {
  const res = await api.put<Category>(`/categories/${id}`, payload);
  return res.data;
}

export async function deleteCategory(id: string) {
  await api.delete(`/categories/${id}`);
}

export async function getCategory(id: string) {
  const res = await api.get<Category>(`/categories/${id}`);
  return res.data;
}
