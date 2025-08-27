import api from '../lib/api';

export interface Adjustment {
  id: string;
  number: string;
  location_id: string;
  location?: Location;
  tenant_id: string;
  reason: string;
  status: string;
  notes?: string;
  created_by?: string;
  approved_by?: string;
  approved_at?: string;
  created_at: string;
  updated_at: string;
  lines?: AdjustmentLine[];
}

export interface AdjustmentLine {
  id: string;
  adjustment_id: string;
  item_id?: string;
  item_identifier: string;
  item?: Item;
  qty_expected: number;
  qty_actual: number;
  qty_diff: number;
  notes?: string;
}

export interface Location {
  id: string;
  name: string;
  code: string;
}

export interface Item {
  id: string;
  sku: string;
  name: string;
  uom: string;
}

export interface CreateAdjustmentRequest {
  location_id: string;
  reason: string;
  notes: string;
  lines: {
    item_id: string;
    qty_expected: number;
    qty_actual: number;
    notes: string;
  }[];
}

export interface UpdateAdjustmentRequest {
  location_id: string;
  reason: string;
  notes: string;
  lines: {
    item_id: string;
    qty_expected: number;
    qty_actual: number;
    notes: string;
  }[];
}

export interface ListAdjustmentsParams {
  page?: number;
  limit?: number;
  status?: string;
  reason?: string;
  search?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_pages: number;
  total: number;
}

// Adjustment reasons enum
export const ADJUSTMENT_REASONS = {
  COUNT: 'COUNT',
  DAMAGE: 'DAMAGE',
  CORRECTION: 'CORRECTION',
  EXPIRY: 'EXPIRY',
  THEFT: 'THEFT',
  OTHER: 'OTHER',
} as const;

export type AdjustmentReason = typeof ADJUSTMENT_REASONS[keyof typeof ADJUSTMENT_REASONS];

// Adjustment statuses enum
export const ADJUSTMENT_STATUSES = {
  DRAFT: 'DRAFT',
  APPROVED: 'APPROVED',
  CANCELED: 'CANCELED',
} as const;

export type AdjustmentStatus = typeof ADJUSTMENT_STATUSES[keyof typeof ADJUSTMENT_STATUSES];

// API functions
export async function listAdjustments(params?: ListAdjustmentsParams): Promise<PaginatedResponse<Adjustment>> {
  const searchParams = new URLSearchParams();
  
  if (params?.page) searchParams.append('page', params.page.toString());
  if (params?.limit) searchParams.append('limit', params.limit.toString());
  if (params?.status) searchParams.append('status', params.status);
  if (params?.reason) searchParams.append('reason', params.reason);
  if (params?.search) searchParams.append('search', params.search);

  const response = await api.get(`/adjustments?${searchParams.toString()}`);
  return response.data;
}

export async function getAdjustment(id: string): Promise<Adjustment> {
  const response = await api.get(`/adjustments/${id}`);
  return response.data;
}

export async function createAdjustment(data: CreateAdjustmentRequest): Promise<Adjustment> {
  const response = await api.post('/adjustments', data);
  return response.data;
}

export async function updateAdjustment(id: string, data: UpdateAdjustmentRequest): Promise<void> {
  await api.put(`/adjustments/${id}`, data);
}

export async function deleteAdjustment(id: string): Promise<void> {
  await api.delete(`/adjustments/${id}`);
}

export async function approveAdjustment(id: string): Promise<void> {
  await api.post(`/adjustments/${id}/approve`);
}

// Utility functions
export function getAdjustmentReasonLabel(reason: string): string {
  switch (reason) {
    case 'COUNT':
      return 'Stock Count';
    case 'DAMAGE':
      return 'Damage';
    case 'CORRECTION':
      return 'Correction';
    case 'EXPIRY':
      return 'Expiry';
    case 'THEFT':
      return 'Theft';
    case 'OTHER':
      return 'Other';
    default:
      return reason;
  }
}

export function getAdjustmentStatusLabel(status: string): string {
  switch (status) {
    case 'DRAFT':
      return 'Draft';
    case 'APPROVED':
      return 'Approved';
    case 'CANCELED':
      return 'Canceled';
    default:
      return status;
  }
}

export function getAdjustmentStatusColor(status: string): string {
  switch (status) {
    case 'DRAFT':
      return 'bg-yellow-100 text-yellow-800';
    case 'APPROVED':
      return 'bg-green-100 text-green-800';
    case 'CANCELED':
      return 'bg-red-100 text-red-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}
