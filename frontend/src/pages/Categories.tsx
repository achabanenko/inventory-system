import { useEffect, useState } from 'react';
import { toast } from 'react-hot-toast';
import { Plus, Search, Filter, X, FolderOpen, Folder } from 'lucide-react';
import { listCategories, createCategory, updateCategory, deleteCategory, type Category } from '../api/categories';

type FormState = {
  id?: string;
  name: string;
  parent_id?: string;
};

export default function Categories() {
  const [searchTerm, setSearchTerm] = useState('');
  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [formState, setFormState] = useState<FormState | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);

  const reload = async (search?: string, targetPage?: number) => {
    setIsLoading(true);
    try {
      const effectivePage = targetPage ?? page;
      const data = await listCategories({ q: search ?? searchTerm, page: effectivePage, page_size: 20, sort: 'name ASC' });
      setCategories(data.data);
      setTotalPages(data.total_pages ?? 1);
      setTotal(data.total ?? 0);
      if (typeof targetPage === 'number') setPage(targetPage);
    } finally {
      setIsLoading(false);
    }
  };

  // Search with debounce
  useEffect(() => {
    const timer = setTimeout(() => {
      reload(undefined, 1);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  useEffect(() => {
    reload();
  }, [page]);

  const onAdd = () => {
    setFormState({ name: '', parent_id: '' });
    setFormOpen(true);
  };

  const onEdit = (row: Category) => {
    setFormState({
      id: row.id,
      name: row.name,
      parent_id: row.parent_id || '',
    });
    setFormOpen(true);
  };

  const onDelete = async (row: Category) => {
    if (!confirm(`Delete category "${row.name}"? This will fail if there are items assigned to this category.`)) return;
    try {
      await deleteCategory(row.id);
      toast.success('Category deleted');
      reload();
    } catch (error: any) {
      if (error?.response?.status === 409) {
        toast.error('Cannot delete category with items assigned to it');
      } else {
        toast.error('Failed to delete category');
      }
    }
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formState) return;
    
    const payload = {
      name: formState.name.trim(),
      parent_id: formState.parent_id || undefined,
    };

    try {
      if (formState.id) {
        await updateCategory(formState.id, payload);
        toast.success('Category updated');
      } else {
        await createCategory(payload);
        toast.success('Category created');
      }
      setFormOpen(false);
      setFormState(null);
      reload(undefined, 1);
    } catch (error: any) {
      const status = error?.response?.status;
      if (status === 409) {
        toast.error('Category name already exists');
      } else if (status === 400) {
        toast.error(error?.response?.data?.error?.message || 'Validation error');
      } else {
        toast.error('Failed to save category');
      }
    }
  };

  const getParentName = (parentId: string | null) => {
    if (!parentId) return '—';
    const parent = categories.find(c => c.id === parentId);
    return parent ? parent.name : 'Unknown';
  };

  const getCategoryIcon = (category: Category) => {
    return category.parent_id ? <Folder className="h-4 w-4" /> : <FolderOpen className="h-4 w-4" />;
  };

  return (
    <div>
      <div className="mb-8 flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Categories</h1>
          <p className="mt-1 text-sm text-gray-500">
            Organize your items into logical groups
          </p>
        </div>
        <button onClick={onAdd} className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Category
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
                  placeholder="Search categories by name..."
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
                  Category
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Parent Category
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Created
                </th>
                <th className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {isLoading && (
                <tr><td className="px-6 py-4 text-sm text-gray-500" colSpan={4}>
                  {searchTerm ? `Searching for "${searchTerm}"...` : 'Loading...'}
                </td></tr>
              )}
              {!isLoading && categories.length === 0 && (
                <tr><td className="px-6 py-4 text-sm text-gray-500" colSpan={4}>
                  {searchTerm ? `No categories found for "${searchTerm}"` : 'No categories'}
                </td></tr>
              )}
              {!isLoading && categories.map((row) => (
                <tr key={row.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      {getCategoryIcon(row)}
                      <span className="ml-2 text-sm font-medium text-gray-900">{row.name}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {getParentName(row.parent_id)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(row.created_at).toLocaleDateString()}
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
              <h2 className="text-lg font-semibold">{formState.id ? 'Edit Category' : 'Add Category'}</h2>
              <button onClick={() => { setFormOpen(false); setFormState(null); }} className="text-gray-500">✕</button>
            </div>
            <form onSubmit={onSubmit} className="p-4 space-y-3">
              <div>
                <label className="block text-sm text-gray-600 mb-1">Name</label>
                <input 
                  value={formState.name} 
                  onChange={e => setFormState(s => ({...s!, name: e.target.value}))} 
                  className="w-full border rounded px-2 py-1" 
                  required 
                />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Parent Category (Optional)</label>
                <select 
                  value={formState.parent_id || ''} 
                  onChange={e => setFormState(s => ({...s!, parent_id: e.target.value || undefined}))} 
                  className="w-full border rounded px-2 py-1"
                >
                  <option value="">No Parent</option>
                  {categories
                    .filter(cat => !cat.parent_id) // Only show top-level categories as parents
                    .filter(cat => cat.id !== formState.id) // Don't allow self-reference
                    .map(category => (
                      <option key={category.id} value={category.id}>{category.name}</option>
                    ))}
                </select>
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
