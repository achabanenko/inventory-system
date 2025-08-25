import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listInventory, type InventoryLevelRow } from '../api/inventory';
import { listItems, type Item } from '../api/items';
import { listLocations, type Location } from '../api/locations';

export default function Inventory() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [selectedItemId, setSelectedItemId] = useState<string>('');
  const [selectedLocationId, setSelectedLocationId] = useState<string>('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(t);
  }, [search]);

  // Autocomplete items when searching
  const { data: itemsList } = useQuery({
    queryKey: ['items', { q: debouncedSearch }],
    queryFn: () => listItems({ q: debouncedSearch, page_size: 20 }),
    enabled: debouncedSearch.length > 0,
  });

  const { data: locationsList } = useQuery({
    queryKey: ['locations', { is_active: true, page_size: 200 }],
    queryFn: () => listLocations({ is_active: true, page_size: 200 }),
  });

  const queryParams = useMemo(() => ({
    item_id: selectedItemId || undefined,
    location_id: selectedLocationId || undefined,
    page,
    page_size: 20,
  }), [selectedItemId, selectedLocationId, page]);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['inventory', queryParams],
    queryFn: () => listInventory(queryParams),
  });

  const rows = (data?.data || []).map((row) => ({
    ...row,
    available: row.available ?? (row.on_hand - row.allocated),
  } as InventoryLevelRow));

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Inventory</h1>
        <p className="mt-1 text-sm text-gray-500">
          Track stock levels across all locations
        </p>
      </div>

      <div className="mb-6 bg-white shadow rounded-lg p-4">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">Search Items</label>
            <input
              type="text"
              placeholder="Search by name, SKU, or barcode"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
              list="inventory-items"
            />
            <datalist id="inventory-items">
              {itemsList?.data?.map((it: Item) => (
                <option key={it.id} value={it.id}>{it.name} ({it.sku})</option>
              ))}
            </datalist>
            <p className="text-xs text-gray-500 mt-1">Tip: selecting from suggestions sets the Item filter below.</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Filter by Item</label>
            <select
              value={selectedItemId}
              onChange={(e) => setSelectedItemId(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All items</option>
              {itemsList?.data?.map((it: Item) => (
                <option key={it.id} value={it.id}>{it.name} ({it.sku})</option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Filter by Location</label>
            <select
              value={selectedLocationId}
              onChange={(e) => setSelectedLocationId(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All locations</option>
              {locationsList?.data?.map((loc: Location) => (
                <option key={loc.id} value={loc.id}>{loc.name} ({loc.code})</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Item</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">SKU</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Location</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">On Hand</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Allocated</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Available</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Reorder Pt.</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="px-6 py-4 text-center text-sm text-gray-500">Loading inventory...</td>
                </tr>
              ) : error ? (
                <tr>
                  <td colSpan={7} className="px-6 py-4 text-center text-sm text-red-600">
                    Failed to load inventory. <button className="text-blue-600 underline" onClick={() => refetch()}>Retry</button>
                  </td>
                </tr>
              ) : rows.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-6 py-4 text-center text-sm text-gray-500">No inventory records</td>
                </tr>
              ) : (
                rows.map((r) => (
                  <tr key={`${r.item.id}-${r.location.id}`} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{r.item.name}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{r.item.sku}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{r.location.name} ({r.location.code})</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{r.on_hand}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{r.allocated}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium text-gray-900">{r.available}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{r.reorder_point ?? '-'}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {data && (data.total_pages ?? 0) > 1 && (
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
            <div className="flex-1 flex justify-between sm:hidden">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              <button
                onClick={() => setPage(Math.min(data.total_pages, page + 1))}
                disabled={page >= data.total_pages}
                className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
            <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">
                  Page <span className="font-medium">{page}</span> of <span className="font-medium">{data.total_pages}</span>
                </p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page <= 1}
                  className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage(Math.min(data.total_pages, page + 1))}
                  disabled={page >= data.total_pages}
                  className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}