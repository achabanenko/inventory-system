import api from '../lib/api';

export type CountStatus = 'OPEN' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELED';

export interface CountBatch {
  id: string;
  number: string;
  location_id: string;
  status: CountStatus;
  notes?: string;
  created_by?: string;
  completed_at?: string;
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

export interface ListBatchesParams {
  status?: CountStatus;
  location_id?: string;
  page?: number;
  page_size?: number;
}

export const listBatches = async (params: ListBatchesParams = {}): Promise<PaginatedResponse<CountBatch>> => {
  const res = await api.get<PaginatedResponse<CountBatch>>('/counts', { params });
  return res.data;
};

export const createBatch = async (payload: { location_id: string; notes?: string }): Promise<CountBatch> => {
  const res = await api.post<CountBatch>('/counts', payload);
  return res.data;
};

export const updateBatch = async (id: string, payload: Partial<{ location_id: string; status: CountStatus; notes: string }>): Promise<CountBatch> => {
  const res = await api.put<CountBatch>(`/counts/${id}`, payload);
  return res.data;
};

export const deleteBatch = async (id: string): Promise<void> => {
  await api.delete(`/counts/${id}`);
};

export interface CountLine {
  id: string;
  batch_id: string;
  item_id: string;
  item_sku?: string;
  item_name?: string;
  expected_on_hand: number;
  counted_qty: number;
  created_at: string;
  updated_at: string;
}

export const listLines = async (batchId: string): Promise<{ data: CountLine[] }> => {
  const res = await api.get<{ data: CountLine[] }>(`/counts/${batchId}/lines`);
  return res.data;
};

export const addLine = async (batchId: string, payload: { item_id: string; expected_on_hand: number; counted_qty: number }): Promise<CountLine> => {
  const res = await api.post<CountLine>(`/counts/${batchId}/lines`, payload);
  return res.data;
};

export const updateLine = async (batchId: string, lineId: string, payload: Partial<{ expected_on_hand: number; counted_qty: number }>): Promise<CountLine> => {
  const res = await api.put<CountLine>(`/counts/${batchId}/lines/${lineId}`, payload);
  return res.data;
};

export const deleteLine = async (batchId: string, lineId: string): Promise<void> => {
  await api.delete(`/counts/${batchId}/lines/${lineId}`);
};


