import api from '../lib/api';

export interface Location {
  id: string;
  code: string;
  name: string;
  is_active: boolean;
  address?: any;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_pages: number;
  total: number;
}

export interface ListLocationsParams {
  q?: string;
  page?: number;
  page_size?: number;
  is_active?: boolean;
}

export async function listLocations(params: ListLocationsParams = {}): Promise<PaginatedResponse<Location>> {
  const res = await api.get<PaginatedResponse<Location>>('/locations', { params });
  return res.data;
}

export const getLocation = async (id: string): Promise<Location> => {
  const res = await api.get<Location>(`/locations/${id}`);
  return res.data;
};

export interface UpsertLocationPayload {
  code: string;
  name: string;
  is_active?: boolean;
  address?: Record<string, unknown> | null;
}

export const createLocation = async (payload: UpsertLocationPayload): Promise<Location> => {
  const res = await api.post<Location>('/locations', payload);
  return res.data;
};

export const updateLocation = async (id: string, payload: UpsertLocationPayload): Promise<Location> => {
  const res = await api.put<Location>(`/locations/${id}`, payload);
  return res.data;
};

export const deleteLocation = async (id: string): Promise<void> => {
  await api.delete(`/locations/${id}`);
};


