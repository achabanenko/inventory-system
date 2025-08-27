import { useState, useEffect } from 'react';
import { listReceiptLines, addReceiptLine, updateReceiptLine, deleteReceiptLine } from '../api/receipts';
import { toast } from 'react-hot-toast';
import { X, Plus, Edit2, Trash2, Save, XCircle } from 'lucide-react';

interface ReceiptLineEditorProps {
  isOpen: boolean;
  onClose: () => void;
  receiptId: string;
  receiptNumber: string;
  receiptStatus: string;
}

interface EditableLine {
  id?: string;
  item_identifier: string;
  qty: number;
  unit_cost: string;
  isEditing: boolean;
  isNew: boolean;
}

export default function ReceiptLineEditor({ 
  isOpen, 
  onClose, 
  receiptId, 
  receiptNumber, 
  receiptStatus 
}: ReceiptLineEditorProps) {
  const [lines, setLines] = useState<EditableLine[]>([]);
  // Removed unused searchTerm state
  const [isLoading, setIsLoading] = useState(false);

  // Removed unused itemsData query since we're not using search suggestions

  // Load existing lines when component opens
  useEffect(() => {
    if (isOpen && receiptId) {
      loadLines();
    }
  }, [isOpen, receiptId]);

  const loadLines = async () => {
    try {
      const response = await listReceiptLines(receiptId);
      const existingLines = response.data.map(line => ({
        id: line.id,
        item_identifier: line.item?.sku || line.item?.name || line.item_id || 'Unknown Item',
        qty: line.qty,
        unit_cost: line.unit_cost ? String(line.unit_cost) : '',
        isEditing: false,
        isNew: false,
      }));
      setLines(existingLines);
    } catch (error) {
      toast.error('Failed to load receipt lines');
    }
  };

  const addNewLine = () => {
    const newLine: EditableLine = {
      item_identifier: '',
      qty: 1,
      unit_cost: '0.00',
      isEditing: true,
      isNew: true,
    };
    setLines([...lines, newLine]);
  };

  const startEditLine = (index: number) => {
    const updatedLines = [...lines];
    updatedLines[index].isEditing = true;
    setLines(updatedLines);
  };

  const cancelEditLine = (index: number) => {
    const updatedLines = [...lines];
    if (updatedLines[index].isNew) {
      // Remove new line if canceling
      updatedLines.splice(index, 1);
    } else {
      // Reset to original values
      updatedLines[index].isEditing = false;
    }
    setLines(updatedLines);
  };

  const saveLine = async (index: number) => {
    const line = lines[index];
    
    if (!line.item_identifier.trim()) {
      toast.error('Item identifier is required');
      return;
    }
    
    if (line.qty <= 0) {
      toast.error('Quantity must be greater than 0');
      return;
    }
    
    if (line.unit_cost !== '' && parseFloat(line.unit_cost) < 0) {
      toast.error('Unit cost must be non-negative');
      return;
    }

    setIsLoading(true);
    try {
      if (line.isNew) {
        // Create new line
        await addReceiptLine(receiptId, {
          item_id: line.item_identifier,
          qty: line.qty,
          unit_cost: line.unit_cost,
        });
        toast.success('Line added successfully');
      } else if (line.id) {
        // Update existing line
        await updateReceiptLine(receiptId, line.id, {
          qty: line.qty,
          unit_cost: line.unit_cost,
        });
        toast.success('Line updated successfully');
      }
      
      // Reload lines to get updated data
      await loadLines();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to save line');
    } finally {
      setIsLoading(false);
    }
  };

  const deleteLine = async (index: number) => {
    const line = lines[index];
    
    if (line.isNew) {
      // Just remove from local state
      const updatedLines = [...lines];
      updatedLines.splice(index, 1);
      setLines(updatedLines);
      return;
    }

    if (!line.id) {
      toast.error('Cannot delete line without ID');
      return;
    }

    if (!window.confirm('Are you sure you want to delete this line?')) {
      return;
    }

    setIsLoading(true);
    try {
      await deleteReceiptLine(receiptId, line.id);
      toast.success('Line deleted successfully');
      await loadLines();
    } catch (error: any) {
      toast.error(error?.response?.data?.message || 'Failed to delete line');
    } finally {
      setIsLoading(false);
    }
  };

  const updateLineField = (index: number, field: keyof EditableLine, value: any) => {
    const updatedLines = [...lines];
    updatedLines[index] = { ...updatedLines[index], [field]: value };
    setLines(updatedLines);
  };

  const calculateLineTotal = (line: EditableLine) => {
    if (!line.unit_cost || line.unit_cost === '') {
      return 'N/A';
    }
    return (line.qty * parseFloat(line.unit_cost)).toFixed(2);
  };

  const calculateTotal = () => {
    return lines.reduce((total, line) => {
      if (!line.unit_cost || line.unit_cost === '') {
        return total;
      }
      return total + (line.qty * parseFloat(line.unit_cost));
    }, 0).toFixed(2);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-6xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-900">
              Edit Receipt Lines - {receiptNumber}
            </h2>
            <p className="text-sm text-gray-500">Status: {receiptStatus}</p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            disabled={isLoading}
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Add New Line Button */}
          <div className="mb-6">
            <button
              onClick={addNewLine}
              disabled={isLoading}
              className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add New Line
            </button>
          </div>

          {/* Lines Table */}
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Item (SKU/Barcode/Code)
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Quantity
                  </th>
                                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Unit Cost (Optional)
                    </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Line Total
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {lines.map((line, index) => (
                  <tr key={index} className={line.isNew ? 'bg-blue-50' : ''}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {line.isEditing ? (
                        <div>
                          <input
                            type="text"
                            value={line.item_identifier}
                            onChange={(e) => updateLineField(index, 'item_identifier', e.target.value)}
                            placeholder="Enter SKU, barcode, or item code"
                            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          />
                          {/* Item suggestions removed for simplicity */}
                        </div>
                      ) : (
                        <span className="text-sm text-gray-900">{line.item_identifier}</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {line.isEditing ? (
                        <input
                          type="number"
                          min="1"
                          value={line.qty}
                          onChange={(e) => updateLineField(index, 'qty', parseInt(e.target.value) || 1)}
                          className="w-20 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                      ) : (
                        <span className="text-sm text-gray-900">{line.qty}</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {line.isEditing ? (
                                                  <input
                            type="number"
                            step="0.01"
                            min="0"
                            value={line.unit_cost}
                            onChange={(e) => updateLineField(index, 'unit_cost', e.target.value)}
                            placeholder="Optional"
                            className="w-24 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          />
                                              ) : (
                          <span className="text-sm text-gray-900">
                            {line.unit_cost && line.unit_cost !== '' ? `$${line.unit_cost}` : 'N/A'}
                          </span>
                        )}
                    </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-medium">
                          {calculateLineTotal(line) === 'N/A' ? 'N/A' : `$${calculateLineTotal(line)}`}
                        </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center space-x-2">
                        {line.isEditing ? (
                          <>
                            <button
                              onClick={() => saveLine(index)}
                              disabled={isLoading}
                              className="text-green-600 hover:text-green-900 disabled:opacity-50"
                              title="Save"
                            >
                              <Save className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => cancelEditLine(index)}
                              disabled={isLoading}
                              className="text-gray-600 hover:text-gray-900 disabled:opacity-50"
                              title="Cancel"
                            >
                              <XCircle className="h-4 w-4" />
                            </button>
                          </>
                        ) : (
                          <>
                            <button
                              onClick={() => startEditLine(index)}
                              disabled={isLoading}
                              className="text-blue-600 hover:text-blue-900 disabled:opacity-50"
                              title="Edit"
                            >
                              <Edit2 className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => deleteLine(index)}
                              disabled={isLoading}
                              className="text-red-600 hover:text-red-900 disabled:opacity-50"
                              title="Delete"
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
                {lines.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-6 py-4 text-center text-sm text-gray-500">
                      No lines added yet. Click "Add New Line" to get started.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Total */}
          {lines.length > 0 && (
            <div className="mt-6 bg-gray-50 rounded-lg p-4">
              <div className="flex justify-between items-center">
                <span className="text-lg font-medium text-gray-900">Total:</span>
                <span className="text-xl font-bold text-gray-900">
                  {calculateTotal() === '0.00' ? 'N/A (No costs provided)' : `$${calculateTotal()}`}
                </span>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-200 flex justify-end">
          <button
            onClick={onClose}
            disabled={isLoading}
            className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
