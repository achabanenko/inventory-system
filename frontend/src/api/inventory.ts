import api from '../lib/api';

export interface InventoryLevelRow {
  item: {
    id: string;
    sku: string;
    name: string;
  };
  location: {
    id: string;
    code: string;
    name: string;
  };
  on_hand: number;
  allocated: number;
  available?: number; // if not provided, compute as on_hand - allocated
  reorder_point?: number;
  reorder_qty?: number;
  updated_at?: string;
}

export interface InventoryPaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total_pages: number;
  total: number;
}

export interface ListInventoryParams {
  item_id?: string;
  location_id?: string;
  page?: number;
  page_size?: number;
}

export async function listInventory(params: ListInventoryParams = {}): Promise<InventoryPaginatedResponse<InventoryLevelRow>> {
  const res = await api.get<InventoryPaginatedResponse<InventoryLevelRow>>('/inventory', { params });
  return res.data;
}

export interface ItemLocationBreakdownRow {
  location: {
    id: string;
    code: string;
    name: string;
  };
  on_hand: number;
  allocated: number;
  available?: number;
}

export async function getItemLocations(itemId: string): Promise<{ item_id: string; locations: ItemLocationBreakdownRow[] }> {
  const res = await api.get<{ item_id: string; locations: ItemLocationBreakdownRow[] }>(`/inventory/${itemId}/locations`);
  return res.data;
}

export type MovementReason = 'PO_RECEIPT' | 'ADJUSTMENT' | 'TRANSFER_OUT' | 'TRANSFER_IN' | 'COUNT';

export interface StockMovementRow {
  id: string;
  item_id: string;
  location_id: string;
  qty: number; // +receive / -issue
  reason: MovementReason;
  reference?: string;
  ref_id?: string;
  occurred_at: string;
}

export interface ListMovementsParams {
  item_id?: string;
  location_id?: string;
  reason?: MovementReason;
  from?: string; // ISO date
  to?: string;   // ISO date
  page?: number;
  page_size?: number;
}

export async function listMovements(params: ListMovementsParams = {}): Promise<InventoryPaginatedResponse<StockMovementRow>> {
  const res = await api.get<InventoryPaginatedResponse<StockMovementRow>>('/inventory/movements', { params });
  return res.data;
}


