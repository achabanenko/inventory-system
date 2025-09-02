import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_barcode_scanner/flutter_barcode_scanner.dart';
import 'package:provider/provider.dart';
import '../models/purchase_order.dart';
import '../models/supplier.dart';
import '../models/item.dart';
import '../services/api_service.dart';
import '../widgets/expandable_fab.dart';
import '../widgets/barcode_scanner_overlay.dart';
import '../screens/scanned_item_screen.dart';

class PurchaseOrderEditScreen extends StatefulWidget {
  final PurchaseOrder purchaseOrder;

  const PurchaseOrderEditScreen({
    super.key,
    required this.purchaseOrder,
  });

  @override
  State<PurchaseOrderEditScreen> createState() => _PurchaseOrderEditScreenState();
}

class _PurchaseOrderEditScreenState extends State<PurchaseOrderEditScreen> {
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
    _initializeFromPO();
    _loadSuppliers();
    // Refresh the purchase order to get the latest data on screen load
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _refreshPurchaseOrder();
    });
  }

  void _initializeFromPO() {
    final po = widget.purchaseOrder;
    _notesController.text = po.notes ?? '';
    _expectedDate = po.expectedAt;
    _lines = List.from(po.lines);
    
    // Find selected supplier
    _selectedSupplier = _suppliers.isNotEmpty 
        ? _suppliers.firstWhere((s) => s.id == po.supplierId, orElse: () => _suppliers.first)
        : null;
  }

  @override
  void dispose() {
    _notesController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _refreshPurchaseOrder() async {
    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      final updatedPO = await apiService.getPurchaseOrder(widget.purchaseOrder.id);
      setState(() {
        _lines = List.from(updatedPO.lines);
      });
    } catch (e) {
      // Handle error silently or show a message
      print('Failed to refresh purchase order: $e');
    }
  }

  Future<void> _loadSuppliers() async {
    setState(() => _isLoading = true);
    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      final suppliers = await apiService.getSuppliers(limit: 100);
      setState(() {
        _suppliers = suppliers;
        // Set selected supplier if not already set
        if (_selectedSupplier == null && suppliers.isNotEmpty) {
          _selectedSupplier = suppliers.firstWhere(
            (s) => s.id == widget.purchaseOrder.supplierId,
            orElse: () => suppliers.first,
          );
        }
      });
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
      print('Starting barcode scan for PO edit...');
      
      // Use the new barcode scanner overlay with purchase order context
      final barcode = await Navigator.of(context).push<String>(
        MaterialPageRoute(
          fullscreenDialog: true,
          builder: (context) => BarcodeScannerOverlay(
            onBarcodeScanned: (code) {
              Navigator.of(context).pop(code);
            },
            onCancel: () {
              Navigator.of(context).pop();
            },
            headerText: 'Scan Item for PO ${widget.purchaseOrder.number}',
            contextWidget: Container(
              child: Row(
                children: [
                  const Icon(Icons.receipt_long, color: Colors.white, size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'PO ${widget.purchaseOrder.number}',
                          style: const TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                        Text(
                          widget.purchaseOrder.supplierName ?? 'Unknown Supplier',
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            mode: ScannerMode.single,
            playSound: true,
          ),
        ),
      );

      print('Scan result: $barcode');
      if (barcode == null) return; // User cancelled

      // Play success sound
      SystemSound.play(SystemSoundType.click);

      // Search for item by barcode using the new search method
      final apiService = Provider.of<ApiService>(context, listen: false);
      final items = await apiService.searchItems(barcode, 'barcode');

      if (items.isNotEmpty && mounted) {
        // Found item - show scanned item detail screen
        _addItemToOrder(items.first);
      } else if (mounted) {
        // Item not found - show option to create or scan another
        _showItemNotFoundDialog(barcode);
      }
    } catch (e) {
      print('Barcode scan error: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Scan failed: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _addItemToOrder(Item item) async {
    // Show scanned item detail screen and wait for result
    final result = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (context) => ScannedItemScreen(
          scannedItem: item,
          documentType: 'Purchase Order',
          contextData: {
            'po_number': widget.purchaseOrder.number,
            'supplier': widget.purchaseOrder.supplierName ?? 'Unknown Supplier',
          },
          onAddToDocument: (item, quantity) async {
            try {
              // Call API to add item with specified quantity
              final apiService = Provider.of<ApiService>(context, listen: false);
              await apiService.addItemToPurchaseOrder(
                widget.purchaseOrder.id,
                item.id,
                quantity.toInt(),
                item.price,
              );

              // Refresh the purchase order to get updated data
              await _refreshPurchaseOrder();

              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('${item.name} added to purchase order'),
                    backgroundColor: Colors.green,
                  ),
                );
              }
              
              // Return success result
              Navigator.pop(context, true);
            } catch (e) {
              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Failed to add item: $e'),
                    backgroundColor: Colors.red,
                  ),
                );
              }
              // Return failure result
              Navigator.pop(context, false);
            }
          },
        ),
      ),
    );

    // Only clear search if item was successfully added or user cancelled
    if (result == true || result == null) {
      _searchController.clear();
      setState(() => _searchResults = []);
    }
  }
  
  void _showItemNotFoundDialog(String barcode) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Row(
          children: [
            Icon(Icons.warning, color: Colors.orange),
            SizedBox(width: 8),
            Text('Item Not Found'),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('No item found with barcode:'),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.grey.shade100,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                barcode,
                style: const TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
            const SizedBox(height: 16),
            const Text('What would you like to do?'),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              _scanBarcode(); // Scan another barcode
            },
            child: const Text('Scan Another'),
          ),
          FilledButton(
            onPressed: () {
              Navigator.pop(context);
              _showCreateItemDialog(barcode);
            },
            child: const Text('Create Item'),
          ),
        ],
      ),
    );
  }
  
  void _showCreateItemDialog(String barcode) {
    final nameController = TextEditingController();
    final skuController = TextEditingController();
    final priceController = TextEditingController(text: '0.00');
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Create New Item'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: skuController,
                decoration: const InputDecoration(
                  labelText: 'Item SKU/Code *',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: nameController,
                decoration: const InputDecoration(
                  labelText: 'Item Name *',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: priceController,
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
                decoration: const InputDecoration(
                  labelText: 'Unit Price',
                  border: OutlineInputBorder(),
                  prefixText: '\$',
                ),
              ),
              const SizedBox(height: 12),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.grey.shade100,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    const Text('Barcode: '),
                    Text(
                      barcode,
                      style: const TextStyle(
                        fontFamily: 'monospace',
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () {
              if (nameController.text.trim().isEmpty || skuController.text.trim().isEmpty) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Please fill in required fields'),
                    backgroundColor: Colors.red,
                  ),
                );
                return;
              }
              
              // Create temporary item for the scanned item flow
              final newItem = Item(
                id: DateTime.now().millisecondsSinceEpoch.toString(),
                sku: skuController.text.trim(),
                name: nameController.text.trim(),
                barcode: barcode,
                price: double.tryParse(priceController.text) ?? 0.0,
                costPrice: double.tryParse(priceController.text) ?? 0.0,
                uom: 'PCS',
                isActive: true,
              );
              
              Navigator.pop(context);
              
              // Show scanned item detail for the new item
              _addItemToOrder(newItem);
            },
            child: const Text('Create & Add'),
          ),
        ],
      ),
    );
  }

  void _showLineActionsMenu(BuildContext context, int index) {
    final line = _lines[index];
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.symmetric(vertical: 20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.edit, color: Colors.blue),
              title: const Text('Edit Quantity'),
              onTap: () {
                Navigator.pop(context);
                _editQuantity(index);
              },
            ),
            ListTile(
              leading: const Icon(Icons.copy, color: Colors.green),
              title: const Text('Duplicate Item'),
              onTap: () {
                Navigator.pop(context);
                _duplicateLine(index);
              },
            ),
            ListTile(
              leading: const Icon(Icons.info, color: Colors.orange),
              title: const Text('View Item Details'),
              onTap: () {
                Navigator.pop(context);
                _showItemDetails(line);
              },
            ),
            const Divider(),
            ListTile(
              leading: const Icon(Icons.delete, color: Colors.red),
              title: const Text('Remove from Order'),
              onTap: () {
                Navigator.pop(context);
                _removeLine(index);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _duplicateLine(int index) {
    final line = _lines[index];
    final duplicatedLine = PurchaseOrderLine(
      id: DateTime.now().millisecondsSinceEpoch.toString(), // Temporary ID
      poId: line.poId,
      itemId: line.itemId,
      itemName: line.itemName,
      itemSku: line.itemSku,
      barcode: line.barcode,
      qtyOrdered: line.qtyOrdered,
      qtyReceived: 0, // Reset received quantity for duplicated line
      unitCost: line.unitCost,
      tax: line.tax,
    );
    
    setState(() {
      _lines.insert(index + 1, duplicatedLine);
    });
    
    HapticFeedback.lightImpact();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Duplicated ${line.itemName}'),
        backgroundColor: Colors.green,
      ),
    );
  }

  void _showItemDetails(PurchaseOrderLine line) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(line.itemName ?? 'Item Details'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildDetailRow('SKU', line.itemSku ?? 'N/A'),
            _buildDetailRow('Barcode', line.barcode ?? 'N/A'),
            _buildDetailRow('Unit Cost', '\$${line.unitCost.toStringAsFixed(2)}'),
            _buildDetailRow('Quantity Ordered', '${line.qtyOrdered}'),
            _buildDetailRow('Quantity Received', '${line.qtyReceived}'),
            _buildDetailRow('Line Total', '\$${line.lineTotal.toStringAsFixed(2)}'),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Close'),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(
              '$label:',
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            child: Text(value),
          ),
        ],
      ),
    );
  }

  void _updateLineQuantity(int index, int quantity) async {
    if (quantity <= 0) {
      // Remove the line entirely
      final lineToRemove = _lines[index];
      try {
        final apiService = Provider.of<ApiService>(context, listen: false);
        await apiService.removeItemFromPurchaseOrder(widget.purchaseOrder.id, lineToRemove.id);
        
        // Refresh the purchase order data to get updated lines
        await _refreshPurchaseOrder();
        
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Item removed from purchase order')),
          );
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Failed to remove item: $e')),
          );
        }
      }
    } else {
      final line = _lines[index];
      try {
        final apiService = Provider.of<ApiService>(context, listen: false);
        await apiService.updatePurchaseOrderLineQuantity(
          widget.purchaseOrder.id, 
          line.id, 
          quantity
        );
        
        // Refresh the purchase order data to get updated line IDs
        await _refreshPurchaseOrder();
        
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Quantity updated')),
          );
        }
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Failed to update quantity: $e')),
          );
        }
      }
    }
  }

  void _removeLine(int index) async {
    final line = _lines[index];
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Remove Item'),
        content: Text('Remove "${line.itemName}" from this purchase order?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              await _removeItemFromPurchaseOrder(index, line);
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
  }

  Future<void> _removeItemFromPurchaseOrder(int index, PurchaseOrderLine line) async {
    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      
      // Show loading indicator
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Row(
            children: [
              SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
              ),
              SizedBox(width: 12),
              Text('Removing item...'),
            ],
          ),
          duration: Duration(seconds: 30),
        ),
      );

      // Call API to delete specific line item
      await apiService.removeItemFromPurchaseOrder(widget.purchaseOrder.id, line.id);
      
      // Update local state on successful API call
      setState(() {
        _lines.removeAt(index);
        });

      // Hide loading indicator and show success message
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Removed ${line.itemName} from order'),
          backgroundColor: Colors.green,
          action: SnackBarAction(
            label: 'Undo',
            textColor: Colors.white,
            onPressed: () async {
              try {
                // Re-add item to purchase order
                await apiService.addItemToPurchaseOrder(
                  widget.purchaseOrder.id, 
                  line.itemId, 
                  line.qtyOrdered, 
                  line.unitCost
                );
                
                // Refresh the purchase order to get updated data
                await _refreshPurchaseOrder();
                
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('Restored ${line.itemName} to order')),
                );
              } catch (e) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Failed to restore item: $e'),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
          ),
        ),
      );

    } catch (e) {
      // Hide loading indicator and show error
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to remove item: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  void _editQuantity(int index) {
    final line = _lines[index];
    final controller = TextEditingController(text: line.qtyOrdered.toString());
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Edit Quantity - ${line.itemName}'),
        content: TextField(
          controller: controller,
          keyboardType: TextInputType.number,
          autofocus: true,
          decoration: const InputDecoration(
            labelText: 'Quantity',
            border: OutlineInputBorder(),
          ),
          onSubmitted: (value) {
            final quantity = int.tryParse(value) ?? line.qtyOrdered;
            Navigator.pop(context);
            _updateLineQuantity(index, quantity);
          },
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              final quantity = int.tryParse(controller.text) ?? line.qtyOrdered;
              Navigator.pop(context);
              _updateLineQuantity(index, quantity);
            },
            child: const Text('Update'),
          ),
        ],
      ),
    );
  }

  void _showAddItemBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.7,
        maxChildSize: 0.9,
        minChildSize: 0.5,
        builder: (context, scrollController) {
          return Container(
            decoration: const BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
            ),
            child: Column(
              children: [
                // Handle bar
                Container(
                  width: 40,
                  height: 4,
                  margin: const EdgeInsets.symmetric(vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.grey[300],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                // Title
                const Padding(
                  padding: EdgeInsets.all(16),
                  child: Text(
                    'Add Item to Order',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  ),
                ),
                // Search bar
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: TextField(
                    controller: _searchController,
                    onChanged: (_) => _searchItems(),
                    autofocus: true,
                    decoration: InputDecoration(
                      hintText: 'Search items by name or SKU...',
                      prefixIcon: const Icon(Icons.search),
                      border: const OutlineInputBorder(),
                      suffixIcon: _isSearching
                          ? const Padding(
                              padding: EdgeInsets.all(12),
                              child: SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              ),
                            )
                          : null,
                    ),
                  ),
                ),
                // Search results
                Expanded(
                  child: _searchResults.isEmpty
                      ? const Center(
                          child: Text(
                            'Type to search for items',
                            style: TextStyle(color: Colors.grey),
                          ),
                        )
                      : ListView.builder(
                          controller: scrollController,
                          padding: const EdgeInsets.all(16),
                          itemCount: _searchResults.length,
                          itemBuilder: (context, index) {
                            final item = _searchResults[index];
                            return Card(
                              child: ListTile(
                                leading: Container(
                                  width: 40,
                                  height: 40,
                                  decoration: BoxDecoration(
                                    color: Colors.blue[50],
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: const Icon(
                                    Icons.inventory_2,
                                    color: Colors.blue,
                                  ),
                                ),
                                title: Text(item.name),
                                subtitle: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text('SKU: ${item.sku}'),
                                    Text(
                                      '\$${item.price.toStringAsFixed(2)}',
                                      style: const TextStyle(
                                        color: Colors.green,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ],
                                ),
                                trailing: IconButton(
                                  icon: const Icon(Icons.add_circle, color: Colors.blue),
                                  onPressed: () {
                                    Navigator.pop(context);
                                    _addItemToOrder(item);
                                  },
                                ),
                                onTap: () {
                                  Navigator.pop(context);
                                  _addItemToOrder(item);
                                },
                              ),
                            );
                          },
                        ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Future<void> _savePurchaseOrder() async {
    if (!_formKey.currentState!.validate() || _lines.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please add at least one item to the order'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    setState(() => _isSaving = true);

    try {
      final apiService = Provider.of<ApiService>(context, listen: false);

      // Update purchase order using individual parameters
      await apiService.updatePurchaseOrder(
        widget.purchaseOrder.id,
        supplierId: _selectedSupplier?.id ?? widget.purchaseOrder.supplierId,
        expectedAt: _expectedDate,
        notes: _notesController.text.trim().isEmpty ? null : _notesController.text.trim(),
        lines: _lines.map((line) => {
          'id': line.id, // Include the line ID for existing lines
          'item_id': line.itemId,
          'qty_ordered': line.qtyOrdered,
          'unit_cost': line.unitCost.toStringAsFixed(2),
        }).toList(),
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Purchase order updated successfully')),
        );
        Navigator.pop(context, true); // Return true to indicate changes were saved
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error updating purchase order: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      setState(() => _isSaving = false);
    }
  }


  @override
  Widget build(BuildContext context) {
    final po = widget.purchaseOrder;
    final canEdit = po.status == PurchaseOrderStatus.draft || po.status == PurchaseOrderStatus.approved;

    return Scaffold(
        appBar: AppBar(
          title: Text('Edit PO ${po.number}'),
        ),
        body: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : Column(
                children: [
                  // Status Banner
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(16),
                    color: po.status.color.withOpacity(0.1),
                    child: Row(
                      children: [
                        Icon(Icons.info, color: po.status.color),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'Status: ${po.status.displayName}',
                                style: TextStyle(
                                  color: po.status.color,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              if (!canEdit)
                                const Text(
                                  'This purchase order cannot be edited in its current status',
                                  style: TextStyle(fontSize: 12),
                                ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),

                  // PO General Information Header
                  Container(
                    margin: const EdgeInsets.all(16),
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(12),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.1),
                          blurRadius: 4,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'PO ${po.number}',
                                    style: const TextStyle(
                                      fontSize: 20,
                                      fontWeight: FontWeight.bold,
                                      color: Color(0xFF2563EB),
                                    ),
                                  ),
                                  Text(
                                    po.supplierName ?? 'Unknown Supplier',
                                    style: TextStyle(
                                      fontSize: 16,
                                      color: Colors.grey[600],
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                Text(
                                  '\$${_lines.fold<double>(0, (sum, line) => sum + line.lineTotal).toStringAsFixed(2)}',
                                  style: const TextStyle(
                                    fontSize: 20,
                                    fontWeight: FontWeight.bold,
                                    color: Color(0xFF059669),
                                  ),
                                ),
                                Text(
                                  'Total Amount',
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: Colors.grey[600],
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        Row(
                          children: [
                            Expanded(
                              child: Container(
                                padding: const EdgeInsets.all(12),
                                decoration: BoxDecoration(
                                  color: Colors.blue[50],
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Column(
                                  children: [
                                    Text(
                                      '${_lines.length}',
                                      style: const TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                        color: Color(0xFF2563EB),
                                      ),
                                    ),
                                    Text(
                                      'Items',
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: Colors.grey[600],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Container(
                                padding: const EdgeInsets.all(12),
                                decoration: BoxDecoration(
                                  color: Colors.green[50],
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Column(
                                  children: [
                                    Text(
                                      '${_lines.fold<int>(0, (sum, line) => sum + line.qtyOrdered)}',
                                      style: const TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                        color: Color(0xFF059669),
                                      ),
                                    ),
                                    Text(
                                      'Total Qty',
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: Colors.grey[600],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Container(
                                padding: const EdgeInsets.all(12),
                                decoration: BoxDecoration(
                                  color: Colors.orange[50],
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Column(
                                  children: [
                                    Text(
                                      '${_lines.fold<int>(0, (sum, line) => sum + line.qtyReceived)}',
                                      style: const TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                        color: Color(0xFFEA580C),
                                      ),
                                    ),
                                    Text(
                                      'Received',
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: Colors.grey[600],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),

                  Expanded(
                    child: Form(
                      key: _formKey,
                      child: Column(
                        children: [

                          // Order Lines
                          Expanded(
                            child: _lines.isEmpty
                                ? Center(
                                    child: Column(
                                      mainAxisAlignment: MainAxisAlignment.center,
                                      children: [
                                        Icon(
                                          Icons.qr_code_scanner,
                                          size: 64,
                                          color: Colors.blue[400],
                                        ),
                                        const SizedBox(height: 16),
                                        Text(
                                          'Ready to Add Items!',
                                          style: TextStyle(
                                            fontSize: 20,
                                            fontWeight: FontWeight.bold,
                                            color: Colors.grey[700],
                                          ),
                                        ),
                                        if (canEdit) ...[
                                          const SizedBox(height: 12),
                                          const Text(
                                            'Scan barcodes or search for items to build your purchase order',
                                            textAlign: TextAlign.center,
                                            style: TextStyle(color: Colors.grey),
                                          ),
                                        ],
                                      ],
                                    ),
                                  )
                                : ListView.builder(
                                    padding: const EdgeInsets.all(16),
                                    itemCount: _lines.length,
                                    itemBuilder: (context, index) {
                                      final line = _lines[index];
                                      return Dismissible(
                                        key: Key('line_${line.id}'),
                                        direction: canEdit ? DismissDirection.horizontal : DismissDirection.none,
                                        // Left swipe (end to start) - Delete action
                                        background: canEdit ? Container(
                                          alignment: Alignment.centerLeft,
                                          padding: const EdgeInsets.only(left: 20),
                                          decoration: BoxDecoration(
                                            color: Colors.blue[600],
                                            borderRadius: BorderRadius.circular(8),
                                          ),
                                          child: const Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Icon(Icons.edit, color: Colors.white, size: 24),
                                              SizedBox(width: 8),
                                              Text(
                                                'Edit',
                                                style: TextStyle(
                                                  color: Colors.white,
                                                  fontWeight: FontWeight.bold,
                                                  fontSize: 16,
                                                ),
                                              ),
                                            ],
                                          ),
                                        ) : null,
                                        // Right swipe (start to end) - Edit/Duplicate actions  
                                        secondaryBackground: canEdit ? Container(
                                          alignment: Alignment.centerRight,
                                          padding: const EdgeInsets.only(right: 20),
                                          decoration: BoxDecoration(
                                            color: Colors.red[600],
                                            borderRadius: BorderRadius.circular(8),
                                          ),
                                          child: const Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Text(
                                                'Delete',
                                                style: TextStyle(
                                                  color: Colors.white,
                                                  fontWeight: FontWeight.bold,
                                                  fontSize: 16,
                                                ),
                                              ),
                                              SizedBox(width: 8),
                                              Icon(Icons.delete, color: Colors.white, size: 24),
                                            ],
                                          ),
                                        ) : null,
                                        confirmDismiss: canEdit ? (direction) async {
                                          if (direction == DismissDirection.startToEnd) {
                                            // Left swipe - Edit action (no dismissal)
                                            _editQuantity(index);
                                            return false;
                                          } else {
                                            // Right swipe - Delete action
                                            HapticFeedback.mediumImpact();
                                            return await showDialog<bool>(
                                              context: context,
                                              builder: (context) => AlertDialog(
                                                title: const Text('Remove Item'),
                                                content: Text('Remove "${line.itemName}" from this purchase order?'),
                                                actions: [
                                                  TextButton(
                                                    onPressed: () => Navigator.pop(context, false),
                                                    child: const Text('Cancel'),
                                                  ),
                                                  TextButton(
                                                    onPressed: () => Navigator.pop(context, true),
                                                    style: TextButton.styleFrom(foregroundColor: Colors.red),
                                                    child: const Text('Remove'),
                                                  ),
                                                ],
                                              ),
                                            );
                                          }
                                        } : null,
                                        onDismissed: canEdit ? (direction) async {
                                          if (direction == DismissDirection.endToStart) {
                                            // Only handle delete dismissal - use API integration
                                            await _removeItemFromPurchaseOrder(index, line);
                                            HapticFeedback.lightImpact();
                                          }
                                        } : null,
                                        child: Card(
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
                                                        '\$${line.unitCost.toStringAsFixed(2)} each • Total: \$${line.lineTotal.toStringAsFixed(2)}',
                                                        style: const TextStyle(color: Colors.green),
                                                      ),
                                                      if (line.qtyReceived > 0)
                                                        Text(
                                                          'Received: ${line.qtyReceived} of ${line.qtyOrdered}',
                                                          style: TextStyle(color: Colors.blue[600], fontSize: 12),
                                                        ),
                                                    ],
                                                  ),
                                                ),
                                                if (canEdit)
                                                  Row(
                                                    mainAxisSize: MainAxisSize.min,
                                                    children: [
                                                      // Stepper controls for quantity
                                                      Container(
                                                        decoration: BoxDecoration(
                                                          color: Colors.grey[100],
                                                          borderRadius: BorderRadius.circular(20),
                                                          border: Border.all(color: Colors.grey[300]!),
                                                        ),
                                                        child: Row(
                                                          mainAxisSize: MainAxisSize.min,
                                                          children: [
                                                            InkWell(
                                                              onTap: () {
                                                                HapticFeedback.lightImpact();
                                                                _updateLineQuantity(index, line.qtyOrdered - 1);
                                                              },
                                                              borderRadius: BorderRadius.circular(20),
                                                              child: Container(
                                                                width: 32,
                                                                height: 32,
                                                                decoration: BoxDecoration(
                                                                  color: line.qtyOrdered > 1 ? Colors.red[50] : Colors.grey[200],
                                                                  borderRadius: BorderRadius.circular(20),
                                                                ),
                                                                child: Icon(
                                                                  Icons.remove,
                                                                  size: 18,
                                                                  color: line.qtyOrdered > 1 ? Colors.red[600] : Colors.grey,
                                                                ),
                                                              ),
                                                            ),
                                                            GestureDetector(
                                                              onTap: () => _editQuantity(index),
                                                              child: Container(
                                                                constraints: const BoxConstraints(minWidth: 50),
                                                                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                                                child: Text(
                                                                  '${line.qtyOrdered}',
                                                                  textAlign: TextAlign.center,
                                                                  style: const TextStyle(
                                                                    fontSize: 16,
                                                                    fontWeight: FontWeight.bold,
                                                                  ),
                                                                ),
                                                              ),
                                                            ),
                                                            InkWell(
                                                              onTap: () {
                                                                HapticFeedback.lightImpact();
                                                                _updateLineQuantity(index, line.qtyOrdered + 1);
                                                              },
                                                              borderRadius: BorderRadius.circular(20),
                                                              child: Container(
                                                                width: 32,
                                                                height: 32,
                                                                decoration: BoxDecoration(
                                                                  color: Colors.green[50],
                                                                  borderRadius: BorderRadius.circular(20),
                                                                ),
                                                                child: Icon(
                                                                  Icons.add,
                                                                  size: 18,
                                                                  color: Colors.green[600],
                                                                ),
                                                              ),
                                                            ),
                                                          ],
                                                        ),
                                                      ),
                                                      const SizedBox(width: 8),
                                                      // More actions button
                                                      InkWell(
                                                        onTap: () => _showLineActionsMenu(context, index),
                                                        borderRadius: BorderRadius.circular(20),
                                                        child: Container(
                                                          width: 32,
                                                          height: 32,
                                                          decoration: BoxDecoration(
                                                            color: Colors.grey[100],
                                                            borderRadius: BorderRadius.circular(20),
                                                          ),
                                                          child: Icon(
                                                            Icons.more_vert,
                                                            size: 16,
                                                            color: Colors.grey[600],
                                                          ),
                                                        ),
                                                      ),
                                                    ],
                                                  )
                                                else
                                                  Container(
                                                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                                    decoration: BoxDecoration(
                                                      color: Colors.grey[200],
                                                      borderRadius: BorderRadius.circular(8),
                                                    ),
                                                    child: Text(
                                                      'Qty: ${line.qtyOrdered}',
                                                      style: const TextStyle(fontWeight: FontWeight.bold),
                                                    ),
                                                  ),
                                              ],
                                            ),
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
                  ),
                ],
              ),
        floatingActionButton: canEdit ? ExpandableFab(
          actions: [
            FabAction(
              icon: const Icon(Icons.qr_code_scanner),
              tooltip: 'Scan Barcode',
              backgroundColor: Colors.blue,
              heroTag: 'scan_barcode',
              onPressed: _scanBarcode,
            ),
            FabAction(
              icon: const Icon(Icons.search),
              tooltip: 'Browse Catalog',
              backgroundColor: Colors.green,
              heroTag: 'browse_catalog',
              onPressed: _showAddItemBottomSheet,
            ),
            FabAction(
              icon: const Icon(Icons.add_box),
              tooltip: 'Add Manual Item',
              backgroundColor: Colors.orange,
              heroTag: 'add_manual',
              onPressed: () {
                // TODO: Implement manual item creation
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Manual item creation - Coming Soon!')),
                );
              },
            ),
          ],
        ) : null,
    );
  }

  /// Build highlighted text for search results
  TextSpan _buildHighlightedText(String text, String query) {
    if (query.isEmpty) {
      return TextSpan(text: text);
    }

    final children = <TextSpan>[];
    final lowerText = text.toLowerCase();
    final lowerQuery = query.toLowerCase();
    
    int start = 0;
    int index = lowerText.indexOf(lowerQuery);
    
    while (index != -1) {
      // Add text before the match
      if (index > start) {
        children.add(TextSpan(
          text: text.substring(start, index),
          style: const TextStyle(color: Colors.black87),
        ));
      }
      
      // Add highlighted match
      children.add(TextSpan(
        text: text.substring(index, index + query.length),
        style: TextStyle(
          backgroundColor: Colors.yellow[200],
          color: Colors.black,
          fontWeight: FontWeight.bold,
        ),
      ));
      
      start = index + query.length;
      index = lowerText.indexOf(lowerQuery, start);
    }
    
    // Add remaining text
    if (start < text.length) {
      children.add(TextSpan(
        text: text.substring(start),
        style: const TextStyle(color: Colors.black87),
      ));
    }
    
    return TextSpan(children: children);
  }
}
