import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { listBatches, createBatch, deleteBatch, listLines, addLine, updateLine, deleteLine, type CountBatch, type CountLine, type CountStatus } from '../api/counts';
import { listLocations, type Location } from '../api/locations';
import { listItems, type Item } from '../api/items';
import { toast } from 'react-hot-toast';

export default function Counts() {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState<CountStatus | ''>('');
  const [locationFilter, setLocationFilter] = useState<string>('');
  const [page] = useState(1);
  const [selectedBatch, setSelectedBatch] = useState<CountBatch | null>(null);
  const [addBatchOpen, setAddBatchOpen] = useState(false);
  const [newBatchLocation, setNewBatchLocation] = useState('');
  const [newBatchNotes, setNewBatchNotes] = useState('');

  const { data: locations } = useQuery({ queryKey: ['locations', { page_size: 200, is_active: true }], queryFn: () => listLocations({ page_size: 200, is_active: true }) });

  const params = useMemo(() => ({ status: statusFilter || undefined, location_id: locationFilter || undefined, page, page_size: 20 }), [statusFilter, locationFilter, page]);
  const { data, isLoading, error } = useQuery({ queryKey: ['count-batches', params], queryFn: () => listBatches(params) });

  const createBatchMut = useMutation({
    mutationFn: createBatch,
    onSuccess: () => { toast.success('Batch created'); setAddBatchOpen(false); setNewBatchLocation(''); setNewBatchNotes(''); queryClient.invalidateQueries({ queryKey: ['count-batches'] }); },
    onError: (e: any) => toast.error(e?.response?.data?.message || 'Failed to create batch'),
  });

  const deleteBatchMut = useMutation({
    mutationFn: deleteBatch,
    onSuccess: () => { toast.success('Batch deleted'); queryClient.invalidateQueries({ queryKey: ['count-batches'] }); setSelectedBatch(null); },
    onError: () => toast.error('Failed to delete batch'),
  });

  return (
    <div>
      <div className="mb-8 flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Stock Counts</h1>
          <p className="mt-1 text-sm text-gray-500">Manage stock counting batches and items</p>
        </div>
        <button onClick={() => setAddBatchOpen(true)} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">New Batch</button>
      </div>

      <div className="mb-6 bg-white shadow rounded-lg p-4 grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
          <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value as CountStatus | '')} className="w-full px-3 py-2 border rounded">
            <option value="">All</option>
            <option value="OPEN">Open</option>
            <option value="IN_PROGRESS">In Progress</option>
            <option value="COMPLETED">Completed</option>
            <option value="CANCELED">Canceled</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Location</label>
          <select value={locationFilter} onChange={(e) => setLocationFilter(e.target.value)} className="w-full px-3 py-2 border rounded">
            <option value="">All locations</option>
            {locations?.data?.map((l: Location) => (
              <option key={l.id} value={l.id}>{l.name} ({l.code})</option>
            ))}
          </select>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Number</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Location</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Notes</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading ? (
                <tr><td colSpan={5} className="px-6 py-4 text-center text-sm text-gray-500">Loading...</td></tr>
              ) : error ? (
                <tr><td colSpan={5} className="px-6 py-4 text-center text-sm text-red-600">Failed to load.</td></tr>
              ) : (data?.data?.length ?? 0) === 0 ? (
                <tr><td colSpan={5} className="px-6 py-4 text-center text-sm text-gray-500">No batches</td></tr>
              ) : (
                data!.data.map((b) => (
                  <tr key={b.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{b.number}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{locations?.data?.find(l => l.id === b.location_id)?.name || b.location_id}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{b.status}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{b.notes || '—'}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button onClick={() => setSelectedBatch(b)} className="text-blue-600 hover:text-blue-900 mr-3">Open</button>
                      <button onClick={() => deleteBatchMut.mutate(b.id)} className="text-red-600 hover:text-red-900">Delete</button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* New batch modal */}
      {addBatchOpen && (
        <div className="fixed inset-0 bg-black/20 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow w-full max-w-lg">
            <div className="p-4 border-b flex items-center justify-between">
              <h2 className="text-lg font-semibold">New Count Batch</h2>
              <button onClick={() => setAddBatchOpen(false)} className="text-gray-500">✕</button>
            </div>
            <form onSubmit={(e) => { e.preventDefault(); if (!newBatchLocation) { toast.error('Select location'); return; } createBatchMut.mutate({ location_id: newBatchLocation, notes: newBatchNotes || undefined }); }} className="p-4 space-y-3">
              <div>
                <label className="block text-sm text-gray-600 mb-1">Location</label>
                <select value={newBatchLocation} onChange={(e) => setNewBatchLocation(e.target.value)} className="w-full border rounded px-2 py-1" required>
                  <option value="">Select location...</option>
                  {locations?.data?.map((l: Location) => (
                    <option key={l.id} value={l.id}>{l.name} ({l.code})</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Notes</label>
                <textarea value={newBatchNotes} onChange={(e) => setNewBatchNotes(e.target.value)} className="w-full border rounded px-2 py-1" rows={3} />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setAddBatchOpen(false)} className="px-3 py-1.5 border rounded">Cancel</button>
                <button type="submit" className="px-3 py-1.5 bg-blue-600 text-white rounded">Create</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Batch drawer */}
      {selectedBatch && (
        <BatchDrawer batch={selectedBatch} onClose={() => setSelectedBatch(null)} />
      )}
    </div>
  );
}

function BatchDrawer({ batch, onClose }: { batch: CountBatch; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [itemInput, setItemInput] = useState('');
  const [expected, setExpected] = useState(0);
  const [counted, setCounted] = useState(0);

  const { data: items } = useQuery({ queryKey: ['items', { q: search }], queryFn: () => listItems({ q: search, page_size: 20 }), enabled: search.length > 0 });
  const { data } = useQuery({ queryKey: ['count-lines', batch.id], queryFn: () => listLines(batch.id) });

  const addLineMut = useMutation({
    mutationFn: (vars: { item_id: string; expected_on_hand: number; counted_qty: number }) => addLine(batch.id, vars),
    onSuccess: () => { toast.success('Line added'); queryClient.invalidateQueries({ queryKey: ['count-lines', batch.id] }); setItemInput(''); setExpected(0); setCounted(0); },
    onError: () => toast.error('Failed to add line'),
  });
  const updateLineMut = useMutation({
    mutationFn: (vars: { line_id: string; expected_on_hand?: number; counted_qty?: number }) => updateLine(batch.id, vars.line_id, { expected_on_hand: vars.expected_on_hand, counted_qty: vars.counted_qty }),
    onSuccess: () => { toast.success('Line updated'); queryClient.invalidateQueries({ queryKey: ['count-lines', batch.id] }); },
    onError: () => toast.error('Failed to update line'),
  });
  const deleteLineMut = useMutation({
    mutationFn: (lineId: string) => deleteLine(batch.id, lineId),
    onSuccess: () => { toast.success('Line deleted'); queryClient.invalidateQueries({ queryKey: ['count-lines', batch.id] }); },
    onError: () => toast.error('Failed to delete line'),
  });

  return (
    <div className="fixed inset-0 bg-black/20 flex items-center justify-end">
      <div className="bg-white w-full max-w-3xl h-full shadow-xl flex flex-col">
        <div className="p-4 border-b flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold">Batch {batch.number}</h2>
            <p className="text-sm text-gray-600">Status: {batch.status}</p>
          </div>
          <button onClick={onClose} className="text-gray-500">✕</button>
        </div>
        <div className="p-4 space-y-4 overflow-y-auto">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
            <div className="md:col-span-2">
              <label className="block text-sm text-gray-600 mb-1">Item (pick from suggestions)</label>
              <input
                value={itemInput}
                onChange={(e) => { setItemInput(e.target.value); setSearch(e.target.value); }}
                list="items-suggest"
                placeholder="Type name, SKU, barcode or paste Item ID"
                className="w-full border rounded px-2 py-1"
              />
              <datalist id="items-suggest">
                {items?.data?.map((it: Item) => (
                  <option key={it.id} value={it.id}>{it.name} ({it.sku})</option>
                ))}
              </datalist>
              <p className="text-xs text-gray-500 mt-1">Selecting a suggestion will set the Item ID automatically.</p>
            </div>
            <div>
              <label className="block text-sm text-gray-600 mb-1">Expected</label>
              <input type="number" value={expected} onChange={(e) => setExpected(parseInt(e.target.value || '0', 10))} className="w-full border rounded px-2 py-1" />
            </div>
            <div>
              <label className="block text-sm text-gray-600 mb-1">Counted</label>
              <input type="number" value={counted} onChange={(e) => setCounted(parseInt(e.target.value || '0', 10))} className="w-full border rounded px-2 py-1" />
            </div>
            <div className="md:col-span-4 flex justify-end">
              <button onClick={() => {
                if (!itemInput) { toast.error('Select an item'); return; }
                addLineMut.mutate({ item_id: itemInput, expected_on_hand: expected, counted_qty: counted });
              }} className="px-3 py-1.5 bg-blue-600 text-white rounded">Add Line</button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Item (SKU)</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Expected</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Counted</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Diff</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {data?.data?.map((ln: CountLine) => (
                  <tr key={ln.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{ln.item_sku || ln.item_id}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{ln.expected_on_hand}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{ln.counted_qty}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{ln.counted_qty - ln.expected_on_hand}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button onClick={() => updateLineMut.mutate({ line_id: ln.id, counted_qty: ln.counted_qty + 1 })} className="text-blue-600 hover:text-blue-900 mr-3">+1</button>
                      <button onClick={() => deleteLineMut.mutate(ln.id)} className="text-red-600 hover:text-red-900">Delete</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}


