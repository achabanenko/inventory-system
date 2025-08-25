import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { listReceipts, createReceipt, updateReceipt, deleteReceipt, listReceiptLines, addReceiptLine, updateReceiptLine, deleteReceiptLine, createReceiptFromPO, approveReceipt, postReceipt, closeReceipt, type GoodsReceipt, type GoodsReceiptLine } from '../api/receipts';
import { listSuppliers, type Supplier } from '../api/suppliers';
import { listLocations, type Location } from '../api/locations';
import { listItems, type Item } from '../api/items';
import { toast } from 'react-hot-toast';
import { CheckCircle, Truck, XCircle, Eye } from 'lucide-react';

export default function Receipts() {
  const navigate = useNavigate();
  const [rows, setRows] = useState<GoodsReceipt[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<GoodsReceipt | null>(null);
  const [supplierId, setSupplierId] = useState('');
  const [locationId, setLocationId] = useState('');
  const [reference, setReference] = useState('');
  const [notes, setNotes] = useState('');

  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [fromPOOpen, setFromPOOpen] = useState(false);
  const [poId, setPoId] = useState('');
  const [poLocationId, setPoLocationId] = useState('');
  const [poReference, setPoReference] = useState('');
  const [poNotes, setPoNotes] = useState('');

  const reload = async (targetPage?: number) => {
    setIsLoading(true);
    try {
      const p = targetPage ?? page;
      const data = await listReceipts({ page: p, page_size: 20 });
      setRows(data.data || []);
      setTotalPages(data.total_pages || 1);
      setTotal(data.total || 0);
      if (targetPage) setPage(targetPage);
    } catch (error) {
      console.error('Failed to load receipts:', error);
      setRows([]);
      setTotalPages(1);
      setTotal(0);
      toast.error('Failed to load receipts');
    } finally { 
      setIsLoading(false); 
    }
  };

  useEffect(() => { (async () => {
    try {
      const s = await listSuppliers({ page_size: 200 }); 
      setSuppliers(s.data || []);
      const l = await listLocations({ page_size: 200 }); 
      setLocations(l.data || []);
    } catch (error) {
      console.error('Failed to load suppliers/locations:', error);
      setSuppliers([]);
      setLocations([]);
    }
  })(); }, []);
  useEffect(() => { reload(); }, [page]);

  const onAdd = () => { setEditing(null); setSupplierId(''); setLocationId(''); setReference(''); setNotes(''); setFormOpen(true); };
  const onEdit = (r: GoodsReceipt) => { setEditing(r); setSupplierId(r.supplier_id || ''); setLocationId(r.location_id || ''); setReference(r.reference || ''); setNotes(r.notes || ''); setFormOpen(true); };
  const onDelete = async (r: GoodsReceipt) => { if (!confirm('Delete receipt?')) return; await deleteReceipt(r.id); toast.success('Deleted'); reload(); };
  
  const onApprove = async (r: GoodsReceipt) => {
    setActionLoading(`approve-${r.id}`);
    try {
      await approveReceipt(r.id);
      toast.success('Receipt approved');
      reload();
    } catch {
      toast.error('Failed to approve receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const onPost = async (r: GoodsReceipt) => {
    setActionLoading(`post-${r.id}`);
    try {
      await postReceipt(r.id);
      toast.success('Receipt posted to inventory');
      reload();
    } catch {
      toast.error('Failed to post receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const onClose = async (r: GoodsReceipt) => {
    setActionLoading(`close-${r.id}`);
    try {
      await closeReceipt(r.id);
      toast.success('Receipt closed');
      reload();
    } catch {
      toast.error('Failed to close receipt');
    } finally {
      setActionLoading(null);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'DRAFT':
        return 'bg-gray-100 text-gray-800';
      case 'APPROVED':
        return 'bg-blue-100 text-blue-800';
      case 'POSTED':
        return 'bg-green-100 text-green-800';
      case 'CLOSED':
        return 'bg-purple-100 text-purple-800';
      case 'CANCELED':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editing) { await updateReceipt(editing.id, { supplier_id: supplierId || undefined, location_id: locationId || undefined, reference: reference || undefined, notes: notes || undefined }); toast.success('Updated'); }
      else { await createReceipt({ supplier_id: supplierId || undefined, location_id: locationId || undefined, reference: reference || undefined, notes: notes || undefined }); toast.success('Created'); }
      setFormOpen(false); reload();
    } catch { toast.error('Failed to save receipt'); }
  };

  const [linesOpen, setLinesOpen] = useState(false);
  const [current] = useState<GoodsReceipt | null>(null);
  const [lines, setLines] = useState<GoodsReceiptLine[]>([]);
  const [itemQuery, setItemQuery] = useState('');
  const [itemChoices, setItemChoices] = useState<Item[]>([]);
  const [newItemId, setNewItemId] = useState('');
  const [newQty, setNewQty] = useState(1);
  const [newCost, setNewCost] = useState('0.00');

  useEffect(() => { (async () => { if (!current) return; const res = await listReceiptLines(current.id); setLines(res.data); })(); }, [current]);
  useEffect(() => { (async () => { if (!itemQuery) { setItemChoices([]); return; } const res = await listItems({ q: itemQuery, page_size: 20 }); setItemChoices(res.data); })(); }, [itemQuery]);

  return (
    <div>
      <div className="mb-8 flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Goods Receipts</h1>
          <p className="mt-1 text-sm text-gray-500">Create, edit, and manage goods receipt documents</p>
        </div>
        <div className="flex gap-2">
          <button onClick={onAdd} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">New Receipt</button>
          <button onClick={() => setFromPOOpen(true)} className="px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700">From Purchase Order</button>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Number</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Supplier</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Location</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Total</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading ? (
                <tr><td colSpan={6} className="px-6 py-4 text-center text-sm text-gray-500">Loading...</td></tr>
              ) : !rows || rows.length === 0 ? (
                <tr><td colSpan={6} className="px-6 py-4 text-center text-sm text-gray-500">No receipts</td></tr>
              ) : (
                rows.map((r) => (
                  <tr key={r.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <button
                        onClick={() => navigate(`/receipts/${r.id}`)}
                        className="text-blue-600 hover:text-blue-900 font-medium"
                      >
                        {r.number}
                      </button>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {r.supplier?.name || suppliers.find(s => s.id === r.supplier_id)?.name || '—'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {r.location ? `${r.location.name} (${r.location.code})` : locations.find(l => l.id === r.location_id)?.name || '—'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(r.status)}`}>
                        {r.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium text-gray-900">
                      ${(parseFloat(r.total?.toString() || '0') || 0).toFixed(2)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => navigate(`/receipts/${r.id}`)}
                          className="text-gray-600 hover:text-gray-900"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        {r.status === 'DRAFT' && (
                          <>
                            <button
                              onClick={() => onApprove(r)}
                              disabled={actionLoading === `approve-${r.id}`}
                              className="text-blue-600 hover:text-blue-900 disabled:opacity-50"
                            >
                              <CheckCircle className="h-4 w-4" />
                            </button>
                            <button onClick={() => onEdit(r)} className="text-green-600 hover:text-green-900">Edit</button>
                            <button onClick={() => onDelete(r)} className="text-red-600 hover:text-red-900">Delete</button>
                          </>
                        )}
                        {r.status === 'APPROVED' && (
                          <button
                            onClick={() => onPost(r)}
                            disabled={actionLoading === `post-${r.id}`}
                            className="text-green-600 hover:text-green-900 disabled:opacity-50"
                          >
                            <Truck className="h-4 w-4" />
                          </button>
                        )}
                        {r.status === 'POSTED' && (
                          <button
                            onClick={() => onClose(r)}
                            disabled={actionLoading === `close-${r.id}`}
                            className="text-purple-600 hover:text-purple-900 disabled:opacity-50"
                          >
                            <XCircle className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {totalPages > 1 && (
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
            <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">Page <span className="font-medium">{page}</span> of <span className="font-medium">{totalPages}</span> — Total <span className="font-medium">{total}</span></p>
              </div>
              <div className="flex gap-2">
                <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page <= 1} className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">Previous</button>
                <button onClick={() => setPage(Math.min(totalPages, page + 1))} disabled={page >= totalPages} className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">Next</button>
              </div>
            </div>
          </div>
        )}
      </div>

      {formOpen && (
        <div className="fixed inset-0 bg-black/20 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow w-full max-w-lg">
            <div className="p-4 border-b flex items-center justify-between">
              <h2 className="text-lg font-semibold">{editing ? 'Edit Receipt' : 'New Receipt'}</h2>
              <button onClick={() => setFormOpen(false)} className="text-gray-500">✕</button>
            </div>
            <form onSubmit={onSubmit} className="p-4 space-y-3">
              <div>
                <label className="block text-sm text-gray-600 mb-1">Supplier</label>
                <select value={supplierId} onChange={e => setSupplierId(e.target.value)} className="w-full border rounded px-2 py-1">
                  <option value="">—</option>
                  {suppliers.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Location</label>
                <select value={locationId} onChange={e => setLocationId(e.target.value)} className="w-full border rounded px-2 py-1">
                  <option value="">—</option>
                  {locations.map(l => <option key={l.id} value={l.id}>{l.name} ({l.code})</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Reference</label>
                <input value={reference} onChange={e => setReference(e.target.value)} className="w-full border rounded px-2 py-1" />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Notes</label>
                <textarea value={notes} onChange={e => setNotes(e.target.value)} className="w-full border rounded px-2 py-1" rows={3} />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setFormOpen(false)} className="px-3 py-1.5 border rounded">Cancel</button>
                <button type="submit" className="px-3 py-1.5 bg-blue-600 text-white rounded">Save</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {fromPOOpen && (
        <div className="fixed inset-0 bg-black/20 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow w-full max-w-lg">
            <div className="p-4 border-b flex items-center justify-between">
              <h2 className="text-lg font-semibold">Create from Purchase Order</h2>
              <button onClick={() => setFromPOOpen(false)} className="text-gray-500">✕</button>
            </div>
            <form onSubmit={async (e) => { e.preventDefault(); try { await createReceiptFromPO({ purchase_order_id: poId, location_id: poLocationId, reference: poReference || undefined, notes: poNotes || undefined }); toast.success('Receipt created from PO'); setFromPOOpen(false); setPoId(''); setPoLocationId(''); setPoReference(''); setPoNotes(''); reload(1); } catch { toast.error('Failed to create from PO'); } }} className="p-4 space-y-3">
              <div>
                <label className="block text-sm text-gray-600 mb-1">Purchase Order ID</label>
                <input value={poId} onChange={e => setPoId(e.target.value)} className="w-full border rounded px-2 py-1" placeholder="Paste PO ID" required />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Location</label>
                <select value={poLocationId} onChange={e => setPoLocationId(e.target.value)} className="w-full border rounded px-2 py-1" required>
                  <option value="">—</option>
                  {locations.map(l => <option key={l.id} value={l.id}>{l.name} ({l.code})</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Reference</label>
                <input value={poReference} onChange={e => setPoReference(e.target.value)} className="w-full border rounded px-2 py-1" />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Notes</label>
                <textarea value={poNotes} onChange={e => setPoNotes(e.target.value)} className="w-full border rounded px-2 py-1" rows={3} />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => setFromPOOpen(false)} className="px-3 py-1.5 border rounded">Cancel</button>
                <button type="submit" className="px-3 py-1.5 bg-purple-600 text-white rounded">Create</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {linesOpen && current && (
        <div className="fixed inset-0 bg-black/20 flex items-center justify-end">
          <div className="bg-white w-full max-w-3xl h-full shadow-xl flex flex-col">
            <div className="p-4 border-b flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">Receipt {current.number}</h2>
                <p className="text-sm text-gray-600">Status: {current.status}</p>
              </div>
              <button onClick={() => setLinesOpen(false)} className="text-gray-500">✕</button>
            </div>
            <div className="p-4 space-y-4 overflow-y-auto">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                <div className="md:col-span-2">
                  <label className="block text-sm text-gray-600 mb-1">Item (pick suggestion)</label>
                  <input value={newItemId} onChange={e => { setNewItemId(e.target.value); setItemQuery(e.target.value); }} list="receipt-items" placeholder="Type name/SKU/barcode or ID" className="w-full border rounded px-2 py-1" />
                  <datalist id="receipt-items">
                    {itemChoices.map((it) => (
                      <option key={it.id} value={it.id}>{it.name} ({it.sku})</option>
                    ))}
                  </datalist>
                </div>
                <div>
                  <label className="block text-sm text-gray-600 mb-1">Qty</label>
                  <input type="number" min={1} value={newQty} onChange={e => setNewQty(parseInt(e.target.value || '1', 10))} className="w-full border rounded px-2 py-1" />
                </div>
                <div>
                  <label className="block text-sm text-gray-600 mb-1">Unit Cost</label>
                  <input type="number" step="0.01" min={0} value={newCost} onChange={e => setNewCost(e.target.value)} className="w-full border rounded px-2 py-1" />
                </div>
                <div className="md:col-span-4 flex justify-end">
                  <button onClick={async () => { try { await addReceiptLine(current.id, { item_id: newItemId, qty: newQty, unit_cost: newCost }); toast.success('Line added'); const res = await listReceiptLines(current.id); setLines(res.data); setNewItemId(''); setNewQty(1); setNewCost('0.00'); } catch { toast.error('Failed to add line'); } }} className="px-3 py-1.5 bg-blue-600 text-white rounded">Add Line</button>
                </div>
              </div>

              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Item</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Qty</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Unit Cost</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {lines.map((ln) => (
                      <tr key={ln.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{ln.item_id}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{ln.qty}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-900">{ln.unit_cost}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <button onClick={async () => { try { await updateReceiptLine(current.id, ln.id, { qty: ln.qty + 1 }); const res = await listReceiptLines(current.id); setLines(res.data); } catch {} }} className="text-blue-600 hover:text-blue-900 mr-3">+1</button>
                          <button onClick={async () => { try { await deleteReceiptLine(current.id, ln.id); const res = await listReceiptLines(current.id); setLines(res.data); } catch {} }} className="text-red-600 hover:text-red-900">Delete</button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}


