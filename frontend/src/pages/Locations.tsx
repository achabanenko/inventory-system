import { useEffect, useState } from 'react';
import { listLocations, createLocation, updateLocation, deleteLocation, type Location, type UpsertLocationPayload } from '../api/locations';

type FormState = {
  id?: string;
  code: string;
  name: string;
  is_active: boolean;
  address_json: string; // freeform JSON
};

export default function Locations() {
  const [searchTerm, setSearchTerm] = useState('');
  const [rows, setRows] = useState<Location[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [formState, setFormState] = useState<FormState | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const reload = async (search?: string) => {
    setIsLoading(true);
    try {
      const data = await listLocations({ q: search ?? searchTerm, page: 1, page_size: 100 });
      setRows(data.data);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => { reload(); }, []);
  useEffect(() => {
    const t = setTimeout(() => reload(), 300);
    return () => clearTimeout(t);
  }, [searchTerm]);

  const onAdd = () => {
    setFormState({ code: '', name: '', is_active: true, address_json: '{}' });
    setFormOpen(true);
  };

  const onEdit = (row: Location) => {
    setFormState({ id: row.id, code: row.code, name: row.name, is_active: row.is_active, address_json: row.address ? JSON.stringify(row.address, null, 2) : '{}' });
    setFormOpen(true);
  };

  const onDelete = async (row: Location) => {
    if (!confirm(`Delete location ${row.code}?`)) return;
    await deleteLocation(row.id);
    reload();
  };

  return (
    <div>
      <div className="mb-8 flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Locations</h1>
          <p className="mt-1 text-sm text-gray-500">Manage warehouse and storage locations</p>
        </div>
        <button onClick={onAdd} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Add Location</button>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="p-4 border-b border-gray-200">
          <input value={searchTerm} onChange={e => setSearchTerm(e.target.value)} placeholder="Search by code or name..." className="w-full px-3 py-2 border rounded" />
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Code</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Active</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading && (
                <tr><td className="px-6 py-4 text-sm text-gray-500" colSpan={3}>Loading...</td></tr>
              )}
              {!isLoading && rows.length === 0 && (
                <tr><td className="px-6 py-4 text-sm text-gray-500" colSpan={3}>No locations</td></tr>
              )}
              {!isLoading && rows.map((row) => (
                <tr key={row.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{row.code}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{row.name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{row.is_active ? 'Yes' : 'No'}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button onClick={() => onEdit(row)} className="text-blue-600 hover:text-blue-900 mr-3">Edit</button>
                    <button onClick={() => onDelete(row)} className="text-red-600 hover:text-red-900">Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {formOpen && formState && (
        <div className="fixed inset-0 bg-black/20 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow w-full max-w-lg">
            <div className="p-4 border-b flex items-center justify-between">
              <h2 className="text-lg font-semibold">{formState.id ? 'Edit Location' : 'Add Location'}</h2>
              <button onClick={() => { setFormOpen(false); setFormState(null); }} className="text-gray-500">✕</button>
            </div>
            <form onSubmit={async (e) => {
              e.preventDefault();
              if (!formState) return;
              setSubmitting(true);
              let address: Record<string, unknown> | null = null;
              if (formState.address_json.trim()) {
                try { address = JSON.parse(formState.address_json); } catch { alert('Address must be valid JSON'); setSubmitting(false); return; }
              }
              const payload: UpsertLocationPayload = { code: formState.code, name: formState.name, is_active: formState.is_active, address: address ?? undefined };
              if (formState.id) { await updateLocation(formState.id, payload); } else { await createLocation(payload); }
              setSubmitting(false);
              setFormOpen(false);
              setFormState(null);
              reload();
            }} className="p-4 space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm text-gray-600 mb-1">Code</label>
                  <input value={formState.code} onChange={e => setFormState(s => ({...s!, code: e.target.value}))} className="w-full border rounded px-2 py-1" required />
                </div>
                <div>
                  <label className="block text-sm text-gray-600 mb-1">Active</label>
                  <input type="checkbox" checked={formState.is_active} onChange={e => setFormState(s => ({...s!, is_active: e.target.checked}))} />
                </div>
                <div className="col-span-2">
                  <label className="block text-sm text-gray-600 mb-1">Name</label>
                  <input value={formState.name} onChange={e => setFormState(s => ({...s!, name: e.target.value}))} className="w-full border rounded px-2 py-1" required />
                </div>
                <div className="col-span-2">
                  <label className="block text-sm text-gray-600 mb-1">Address (JSON)</label>
                  <textarea rows={4} value={formState.address_json} onChange={e => setFormState(s => ({...s!, address_json: e.target.value}))} className="w-full border rounded px-2 py-1 font-mono text-xs" />
                </div>
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => { setFormOpen(false); setFormState(null); }} className="px-3 py-1.5 border rounded">Cancel</button>
                <button disabled={submitting} type="submit" className="px-3 py-1.5 bg-blue-600 text-white rounded disabled:opacity-50">{submitting ? 'Saving...' : 'Save'}</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}