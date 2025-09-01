import 'package:flutter/material.dart';
import 'package:flutter_barcode_scanner/flutter_barcode_scanner.dart';
import 'package:provider/provider.dart';
import '../models/purchase_order.dart';
import '../models/supplier.dart';
import '../models/item.dart';
import '../services/api_service.dart';

class PurchaseOrderCreateScreen extends StatefulWidget {
  final Item? preSelectedItem;

  const PurchaseOrderCreateScreen({
    super.key,
    this.preSelectedItem,
  });

  @override
  State<PurchaseOrderCreateScreen> createState() => _PurchaseOrderCreateScreenState();
}

class _PurchaseOrderCreateScreenState extends State<PurchaseOrderCreateScreen> {
  final _formKey = GlobalKey<FormState>();
  final TextEditingController _notesController = TextEditingController();
  final TextEditingController _searchController = TextEditingController();

  Supplier? _selectedSupplier;
  List<Supplier> _suppliers = [];
  List<PurchaseOrderLine> _lines = [];
  DateTime? _expectedDate;
  bool _isLoading = false;
  bool _isSaving = false;

  // Item search and selection
  List<Item> _searchResults = [];
  bool _isSearching = false;

  @override
  void initState() {
    super.initState();
    _loadSuppliers();

    // Add pre-selected item if provided
    if (widget.preSelectedItem != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _addItemToOrder(widget.preSelectedItem!);
      });
    }
  }

  @override
  void dispose() {
    _notesController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadSuppliers() async {
    setState(() => _isLoading = true);
    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      final suppliers = await apiService.getSuppliers(limit: 100);
      setState(() => _suppliers = suppliers);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error loading suppliers: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _searchItems() async {
    final query = _searchController.text.trim();
    if (query.isEmpty) {
      setState(() => _searchResults = []);
      return;
    }

    setState(() => _isSearching = true);
    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      final items = await apiService.getItems(search: query, limit: 20);
      setState(() => _searchResults = items);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error searching items: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      setState(() => _isSearching = false);
    }
  }

  Future<void> _scanBarcode() async {
    try {
      final barcode = await FlutterBarcodeScanner.scanBarcode(
        '#ff6666',
        'Cancel',
        true,
        ScanMode.BARCODE,
      );

      if (barcode == '-1') return; // User cancelled

      // Search for item by barcode
      final apiService = Provider.of<ApiService>(context, listen: false);
      final item = await apiService.getItemByBarcode(barcode);

      if (item != null && mounted) {
        _addItemToOrder(item);
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('No item found for barcode: $barcode'),
            backgroundColor: Colors.orange,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Scan failed: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  void _addItemToOrder(Item item) {
    // Check if item is already in the order
    final existingLineIndex = _lines.indexWhere((line) => line.itemId == item.id);

    if (existingLineIndex >= 0) {
      // Update existing line
      final existingLine = _lines[existingLineIndex];
      final updatedLine = PurchaseOrderLine(
        id: existingLine.id,
        poId: existingLine.poId,
        itemId: item.id,
        itemName: item.name,
        itemSku: item.sku,
        barcode: item.barcode,
        qtyOrdered: existingLine.qtyOrdered + 1,
        qtyReceived: existingLine.qtyReceived,
        unitCost: item.costPrice,
        tax: existingLine.tax,
      );
      setState(() {
        _lines[existingLineIndex] = updatedLine;
      });
    } else {
      // Add new line
      final newLine = PurchaseOrderLine(
        id: DateTime.now().millisecondsSinceEpoch.toString(), // Temporary ID
        poId: '', // Will be set when PO is created
        itemId: item.id,
        itemName: item.name,
        itemSku: item.sku,
        barcode: item.barcode,
        qtyOrdered: 1,
        qtyReceived: 0,
        unitCost: item.costPrice,
      );
      setState(() {
        _lines.add(newLine);
      });
    }

    // Clear search
    _searchController.clear();
    setState(() => _searchResults = []);

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Added ${item.name} to order')),
    );
  }

  void _updateLineQuantity(int index, int quantity) {
    if (quantity <= 0) {
      setState(() => _lines.removeAt(index));
    } else {
      final line = _lines[index];
      final updatedLine = PurchaseOrderLine(
        id: line.id,
        poId: line.poId,
        itemId: line.itemId,
        itemName: line.itemName,
        itemSku: line.itemSku,
        barcode: line.barcode,
        qtyOrdered: quantity,
        qtyReceived: line.qtyReceived,
        unitCost: line.unitCost,
        tax: line.tax,
      );
      setState(() => _lines[index] = updatedLine);
    }
  }

  void _removeLine(int index) {
    setState(() => _lines.removeAt(index));
  }

  Future<void> _selectExpectedDate() async {
    final pickedDate = await showDatePicker(
      context: context,
      initialDate: _expectedDate ?? DateTime.now().add(const Duration(days: 7)),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );

    if (pickedDate != null) {
      setState(() => _expectedDate = pickedDate);
    }
  }

  Future<void> _savePurchaseOrder() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedSupplier == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select a supplier'), backgroundColor: Colors.red),
      );
      return;
    }
    if (_lines.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please add at least one item'), backgroundColor: Colors.red),
      );
      return;
    }

    setState(() => _isSaving = true);

    try {
      final apiService = Provider.of<ApiService>(context, listen: false);

      // Convert lines to the format expected by the API
      final linesData = _lines.map((line) => {
        'item_id': line.itemId,
        'qty_ordered': line.qtyOrdered,
        'unit_cost': line.unitCost,
      }).toList();

      final purchaseOrder = await apiService.createPurchaseOrder(
        supplierId: _selectedSupplier!.id,
        expectedAt: _expectedDate,
        notes: _notesController.text.trim().isEmpty ? null : _notesController.text.trim(),
        lines: linesData,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Purchase Order ${purchaseOrder.number} created successfully!')),
        );
        Navigator.pop(context, purchaseOrder);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error creating purchase order: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      setState(() => _isSaving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Create Purchase Order'),
        actions: [
          TextButton(
            onPressed: _isSaving ? null : _savePurchaseOrder,
            child: _isSaving
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Save', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Form(
              key: _formKey,
              child: Column(
                children: [
                  // Supplier and Date Section
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Theme.of(context).cardColor,
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.05),
                          blurRadius: 4,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Column(
                      children: [
                        // Supplier Selection
                        DropdownButtonFormField<Supplier>(
                          value: _selectedSupplier,
                          decoration: const InputDecoration(
                            labelText: 'Supplier *',
                            prefixIcon: Icon(Icons.business),
                          ),
                          items: _suppliers.map((supplier) {
                            return DropdownMenuItem(
                              value: supplier,
                              child: Text(supplier.displayName),
                            );
                          }).toList(),
                          onChanged: (supplier) {
                            setState(() => _selectedSupplier = supplier);
                          },
                          validator: (value) {
                            if (value == null) return 'Please select a supplier';
                            return null;
                          },
                        ),

                        const SizedBox(height: 16),

                        // Expected Date
                        InkWell(
                          onTap: _selectExpectedDate,
                          child: InputDecorator(
                            decoration: const InputDecoration(
                              labelText: 'Expected Date',
                              prefixIcon: Icon(Icons.calendar_today),
                            ),
                            child: Text(
                              _expectedDate != null
                                  ? '${_expectedDate!.month}/${_expectedDate!.day}/${_expectedDate!.year}'
                                  : 'Select expected date',
                              style: TextStyle(
                                color: _expectedDate != null ? null : Colors.grey,
                              ),
                            ),
                          ),
                        ),

                        const SizedBox(height: 16),

                        // Notes
                        TextFormField(
                          controller: _notesController,
                          maxLines: 2,
                          decoration: const InputDecoration(
                            labelText: 'Notes',
                            prefixIcon: Icon(Icons.note),
                          ),
                        ),
                      ],
                    ),
                  ),

                  // Items Section
                  Expanded(
                    child: Column(
                      children: [
                        // Add Items Section
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: Colors.grey[50],
                            border: Border(bottom: BorderSide(color: Colors.grey[200]!)),
                          ),
                          child: Column(
                            children: [
                              Row(
                                children: [
                                  Expanded(
                                    child: TextField(
                                      controller: _searchController,
                                      onChanged: (_) => _searchItems(),
                                      decoration: const InputDecoration(
                                        hintText: 'Search items...',
                                        prefixIcon: Icon(Icons.search),
                                        border: OutlineInputBorder(),
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  ElevatedButton.icon(
                                    onPressed: _scanBarcode,
                                    icon: const Icon(Icons.qr_code_scanner),
                                    label: const Text('Scan'),
                                  ),
                                ],
                              ),

                              // Search Results
                              if (_searchResults.isNotEmpty) ...[
                                const SizedBox(height: 8),
                                Container(
                                  constraints: const BoxConstraints(maxHeight: 200),
                                  decoration: BoxDecoration(
                                    color: Colors.white,
                                    border: Border.all(color: Colors.grey[300]!),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: ListView.builder(
                                    shrinkWrap: true,
                                    itemCount: _searchResults.length,
                                    itemBuilder: (context, index) {
                                      final item = _searchResults[index];
                                      return ListTile(
                                        title: Text(item.name),
                                        subtitle: Text('SKU: ${item.sku} • \$${item.price.toStringAsFixed(2)}'),
                                        trailing: IconButton(
                                          icon: const Icon(Icons.add),
                                          onPressed: () => _addItemToOrder(item),
                                        ),
                                        onTap: () => _addItemToOrder(item),
                                      );
                                    },
                                  ),
                                ),
                              ],
                            ],
                          ),
                        ),

                        // Order Lines
                        Expanded(
                          child: _lines.isEmpty
                              ? Center(
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      Icon(
                                        Icons.inventory_2,
                                        size: 64,
                                        color: Colors.grey[400],
                                      ),
                                      const SizedBox(height: 16),
                                      Text(
                                        'No items added yet',
                                        style: TextStyle(
                                          fontSize: 18,
                                          color: Colors.grey[600],
                                        ),
                                      ),
                                      const SizedBox(height: 8),
                                      const Text(
                                        'Search for items or scan barcodes to add them to the order',
                                        textAlign: TextAlign.center,
                                        style: TextStyle(color: Colors.grey),
                                      ),
                                    ],
                                  ),
                                )
                              : ListView.builder(
                                  padding: const EdgeInsets.all(16),
                                  itemCount: _lines.length,
                                  itemBuilder: (context, index) {
                                    final line = _lines[index];
                                    return Card(
                                      margin: const EdgeInsets.only(bottom: 8),
                                      child: Padding(
                                        padding: const EdgeInsets.all(12),
                                        child: Row(
                                          children: [
                                            Expanded(
                                              child: Column(
                                                crossAxisAlignment: CrossAxisAlignment.start,
                                                children: [
                                                  Text(
                                                    line.itemName ?? 'Unknown Item',
                                                    style: const TextStyle(fontWeight: FontWeight.bold),
                                                  ),
                                                  Text(
                                                    'SKU: ${line.itemSku ?? 'N/A'}',
                                                    style: TextStyle(color: Colors.grey[600], fontSize: 12),
                                                  ),
                                                  Text(
                                                    '\$${line.unitCost.toStringAsFixed(2)} each',
                                                    style: const TextStyle(color: Colors.green),
                                                  ),
                                                ],
                                              ),
                                            ),
                                            Row(
                                              children: [
                                                IconButton(
                                                  icon: const Icon(Icons.remove),
                                                  onPressed: () => _updateLineQuantity(index, line.qtyOrdered - 1),
                                                ),
                                                SizedBox(
                                                  width: 60,
                                                  child: Text(
                                                    '${line.qtyOrdered}',
                                                    textAlign: TextAlign.center,
                                                    style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                                                  ),
                                                ),
                                                IconButton(
                                                  icon: const Icon(Icons.add),
                                                  onPressed: () => _updateLineQuantity(index, line.qtyOrdered + 1),
                                                ),
                                                IconButton(
                                                  icon: const Icon(Icons.delete, color: Colors.red),
                                                  onPressed: () => _removeLine(index),
                                                ),
                                              ],
                                            ),
                                          ],
                                        ),
                                      ),
                                    );
                                  },
                                ),
                        ),

                        // Order Summary
                        if (_lines.isNotEmpty)
                          Container(
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: Theme.of(context).cardColor,
                              boxShadow: [
                                BoxShadow(
                                  color: Colors.black.withOpacity(0.05),
                                  blurRadius: 4,
                                  offset: const Offset(0, -2),
                                ),
                              ],
                            ),
                            child: Row(
                              children: [
                                Expanded(
                                  child: Text(
                                    '${_lines.length} item${_lines.length == 1 ? '' : 's'} • Total: \$${_lines.fold<double>(0, (sum, line) => sum + line.lineTotal).toStringAsFixed(2)}',
                                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                                  ),
                                ),
                              ],
                            ),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
    );
  }
}
