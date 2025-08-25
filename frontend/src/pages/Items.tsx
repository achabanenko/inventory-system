import { useEffect, useState } from 'react';
import { toast } from 'react-hot-toast';
import { Plus, Search, Filter, X } from 'lucide-react';
import { listItems, createItem, updateItem, deleteItem, type Item } from '../api/items';
import { listCategories, type Category } from '../api/categories';

type FormState = {
  id?: string;
  sku: string;
  name: string;
  barcode?: string;
  uom: string;
  category_id?: string;
  cost: string;
  price: string;
  is_active: boolean;
};

export default function Items() {
  const [searchTerm, setSearchTerm] = useState('');
  const [items, setItems] = useState<Item[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [formState, setFormState] = useState<FormState | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [categories, setCategories] = useState<Category[]>([]);

  const reload = async (search?: string, targetPage?: number) => {
    setIsLoading(true);
    try {
      const effectivePage = targetPage ?? page;
      const data = await listItems({ q: search ?? searchTerm, page: effectivePage, page_size: 20, sort: 'created_at DESC' });
      setItems(data.data);
      setTotalPages(data.total_pages ?? 1);
      setTotal(data.total ?? 0);
      // keep page in sync if caller requested a specific page
      if (typeof targetPage === 'number') setPage(targetPage);
    } finally {
      setIsLoading(false);
    }
  };

  const loadCategories = async () => {
    try {
      const data = await listCategories({ page_size: 100 }); // Load up to 100 categories for dropdown
      setCategories(data.data);
    } catch (error) {
      console.error('Failed to load categories:', error);
    }
  };

  // Search with debounce
  useEffect(() => {
    const timer = setTimeout(() => {
      // reset to first page on new search
      reload(undefined, 1);
    }, 300); // 300ms debounce

    return () => clearTimeout(timer);
  }, [searchTerm]);

  useEffect(() => {
    reload();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  useEffect(() => {
    loadCategories();
  }, []);

  const onAdd = () => {
    setFormState({ sku: '', name: '', barcode: '', uom: 'EA', category_id: '', cost: '0.00', price: '0.00', is_active: true });
    setFormOpen(true);
  };

  const onEdit = (row: Item) => {
    setFormState({
      id: row.id,
      sku: row.sku,
      name: row.name,
      barcode: row.barcode ?? '',
      uom: row.uom,
      category_id: row.category_id ?? '',
      cost: typeof row.cost === 'string' ? row.cost : String(row.cost),
      price: typeof row.price === 'string' ? row.price : String(row.price),
      is_active: row.is_active,
    });
    setFormOpen(true);
  };

  const onDelete = async (row: Item) => {
    if (!confirm(`Delete item ${row.sku}?`)) return;
    await deleteItem(row.id);
    reload();
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formState) return;
    const payload = {
      sku: (formState.sku || '').trim().replace(/\s+/g, '-'),
      name: formState.name,
      barcode: formState.barcode || undefined,
      uom: formState.uom,
      category_id: formState.category_id || undefined,
      cost: formState.cost,
      price: formState.price,
      is_active: formState.is_active,
    };
    try {
      if (formState.id) {
        await updateItem(formState.id, payload);
        toast.success('Item updated');
      } else {
        await createItem(payload);
        toast.success('Item created');
      }
      setFormOpen(false);
      setFormState(null);
      reload(undefined, 1);
    } catch (err: any) {
      const status = err?.response?.status;
      if (status === 409) {
        toast.error('SKU already exists. Please use a unique SKU.');
      } else if (status === 400) {
        toast.error(err?.response?.data?.error?.message || 'Validation error');
      } else {
        toast.error('Failed to save item');
      }
    }
  };

  return (
    <div>
      <div className="mb-8 flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Items</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage your product catalog
          </p>
        </div>
        <button onClick={onAdd} className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Item
        </button>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="p-4 border-b border-gray-200">
          <div className="flex gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Search items by SKU, name, or barcode..."
                  className="pl-10 pr-10 py-2 border border-gray-300 rounded-md w-full focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                />
                {searchTerm && (
                  <button
                    onClick={() => setSearchTerm('')}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 hover:text-gray-600"
                  >
                    <X className="h-4 w-4" />
                  </button>
                )}
              </div>
            </div>
            <button className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              <Filter className="h-4 w-4 mr-2" />
              Filters
            </button>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  SKU
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Category
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Price
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Stock
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading && (
                <tr><td className="px-6 py-4 text-sm text-gray-500" colSpan={7}>
                  {searchTerm ? `Searching for "${searchTerm}"...` : 'Loading...'}
                </td></tr>
              )}
              {!isLoading && items.length === 0 && (
                <tr><td className="px-6 py-4 text-sm text-gray-500" colSpan={7}>
                  {searchTerm ? `No items found for "${searchTerm}"` : 'No items'}
                </td></tr>
              )}
              {!isLoading && items.map((row) => (
                <tr key={row.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{row.sku}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{row.name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{row.category?.name || '—'}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${String(row.price)}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">—</td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${row.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}`}>
                      {row.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button onClick={() => onEdit(row)} className="text-blue-600 hover:text-blue-900 mr-3">Edit</button>
                    <button onClick={() => onDelete(row)} className="text-red-600 hover:text-red-900">Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
            <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">
                  Page <span className="font-medium">{page}</span> of{' '}
                  <span className="font-medium">{totalPages}</span> — Total{' '}
                  <span className="font-medium">{total}</span>
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
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page >= totalPages}
                  className="relative inline-flex items-center px-3 py-2 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
      {formOpen && formState && (
        <div className="fixed inset-0 bg-black/20 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow w-full max-w-lg">
            <div className="p-4 border-b flex items-center justify-between">
              <h2 className="text-lg font-semibold">{formState.id ? 'Edit Item' : 'Add Item'}</h2>
              <button onClick={() => { setFormOpen(false); setFormState(null); }} className="text-gray-500">✕</button>
            </div>
            <form onSubmit={onSubmit} className="p-4 space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm text-gray-600 mb-1">SKU</label>
                  <input value={formState.sku} onChange={e => setFormState(s => ({...s!, sku: e.target.value}))} className="w-full border rounded px-2 py-1" required />
                </div>
                <div>
                  <label className="block text-sm text-gray-600 mb-1">UOM</label>
                  <input value={formState.uom} onChange={e => setFormState(s => ({...s!, uom: e.target.value}))} className="w-full border rounded px-2 py-1" required />
                </div>
                <div className="col-span-2">
                  <label className="block text-sm text-gray-600 mb-1">Name</label>
                  <input value={formState.name} onChange={e => setFormState(s => ({...s!, name: e.target.value}))} className="w-full border rounded px-2 py-1" required />
                </div>
                <div className="col-span-2">
                  <label className="block text-sm text-gray-600 mb-1">Barcode</label>
                  <input value={formState.barcode} onChange={e => setFormState(s => ({...s!, barcode: e.target.value}))} className="w-full border rounded px-2 py-1" />
                </div>
                <div className="col-span-2">
                  <label className="block text-sm text-gray-600 mb-1">Category</label>
                  <select 
                    value={formState.category_id || ''} 
                    onChange={e => setFormState(s => ({...s!, category_id: e.target.value || undefined}))} 
                    className="w-full border rounded px-2 py-1"
                  >
                    <option value="">No Category</option>
                    {categories.map(category => (
                      <option key={category.id} value={category.id}>{category.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm text-gray-600 mb-1">Cost</label>
                  <input type="number" step="0.01" value={formState.cost} onChange={e => setFormState(s => ({...s!, cost: e.target.value}))} className="w-full border rounded px-2 py-1" required />
                </div>
                <div>
                  <label className="block text-sm text-gray-600 mb-1">Price</label>
                  <input type="number" step="0.01" value={formState.price} onChange={e => setFormState(s => ({...s!, price: e.target.value}))} className="w-full border rounded px-2 py-1" required />
                </div>
                <div className="col-span-2 flex items-center gap-2">
                  <input id="active" type="checkbox" checked={formState.is_active} onChange={e => setFormState(s => ({...s!, is_active: e.target.checked}))} />
                  <label htmlFor="active" className="text-sm text-gray-700">Active</label>
                </div>
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button type="button" onClick={() => { setFormOpen(false); setFormState(null); }} className="px-3 py-1.5 border rounded">Cancel</button>
                <button type="submit" className="px-3 py-1.5 bg-blue-600 text-white rounded">Save</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}