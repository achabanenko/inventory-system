import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/item.dart';
import '../models/purchase_order.dart';
import '../widgets/universal_item_search.dart';
import '../widgets/quantity_input_widget.dart';
import '../services/api_service.dart';

enum WorkflowType { add, edit, delete }

class PurchaseOrderItemWorkflowScreen extends StatefulWidget {
  final PurchaseOrder purchaseOrder;
  final WorkflowType workflowType;
  final PurchaseOrderLine? existingLine; // For edit/delete operations
  final Function(PurchaseOrderLine) onItemAdded;
  final Function(PurchaseOrderLine, PurchaseOrderLine) onItemUpdated; // (old, new)
  final Function(PurchaseOrderLine) onItemDeleted;
  
  const PurchaseOrderItemWorkflowScreen({
    super.key,
    required this.purchaseOrder,
    required this.workflowType,
    required this.onItemAdded,
    required this.onItemUpdated,
    required this.onItemDeleted,
    this.existingLine,
  });

  @override
  State<PurchaseOrderItemWorkflowScreen> createState() => _PurchaseOrderItemWorkflowScreenState();
}

class _PurchaseOrderItemWorkflowScreenState extends State<PurchaseOrderItemWorkflowScreen> {
  final PageController _pageController = PageController();
  final ApiService _apiService = ApiService();
  
  Item? _selectedItem;
  double _quantity = 1;
  double _unitPrice = 0;
  int _currentPage = 0;
  bool _isLoading = false;
  
  @override
  void initState() {
    super.initState();
    
    // Pre-populate for edit/delete workflows
    if (widget.workflowType != WorkflowType.add && widget.existingLine != null) {
      _quantity = widget.existingLine!.qtyOrdered.toDouble();
      _unitPrice = widget.existingLine!.unitCost;
      
      // If we have item info, create item object
      if (widget.existingLine!.itemSku != null) {
        _selectedItem = Item(
          id: widget.existingLine!.itemId,
          sku: widget.existingLine!.itemSku!,
          name: widget.existingLine!.itemName ?? 'Unknown Item',
          barcode: widget.existingLine!.barcode,
          price: widget.existingLine!.unitCost,
          costPrice: widget.existingLine!.unitCost,
          uom: 'PCS',
          isActive: true,
        );
        
        // For edit/delete, skip to quantity page
        if (widget.workflowType == WorkflowType.edit || widget.workflowType == WorkflowType.delete) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            setState(() {
              _currentPage = 1;
            });
            _pageController.animateToPage(
              1,
              duration: const Duration(milliseconds: 300),
              curve: Curves.easeInOut,
            );
          });
        }
      }
    }
  }
  
  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }
  
  String get _workflowTitle {
    switch (widget.workflowType) {
      case WorkflowType.add:
        return 'Add Item to Purchase Order';
      case WorkflowType.edit:
        return 'Edit Purchase Order Item';
      case WorkflowType.delete:
        return 'Remove Purchase Order Item';
    }
  }
  
  String get _actionButtonText {
    switch (widget.workflowType) {
      case WorkflowType.add:
        return 'Add to Order';
      case WorkflowType.edit:
        return 'Update Item';
      case WorkflowType.delete:
        return 'Remove Item';
    }
  }
  
  Color get _actionButtonColor {
    switch (widget.workflowType) {
      case WorkflowType.add:
        return Colors.green;
      case WorkflowType.edit:
        return Colors.blue;
      case WorkflowType.delete:
        return Colors.red;
    }
  }
  
  IconData get _actionButtonIcon {
    switch (widget.workflowType) {
      case WorkflowType.add:
        return Icons.add_shopping_cart;
      case WorkflowType.edit:
        return Icons.edit;
      case WorkflowType.delete:
        return Icons.delete;
    }
  }
  
  void _onItemSelected(Item item) async {
    setState(() {
      _isLoading = true;
    });
    
    try {
      // For add workflow, get fresh item details
      if (widget.workflowType == WorkflowType.add) {
        final items = await _apiService.searchItems(item.sku, 'code');
        if (items.isNotEmpty) {
          item = items.first;
        }
      }
      
      setState(() {
        _selectedItem = item;
        _unitPrice = item.price;
        if (widget.workflowType == WorkflowType.add) {
          _quantity = 1; // Reset to 1 for new items
        }
        _currentPage = 1;
        _isLoading = false;
      });
      
      // Animate to quantity page
      _pageController.animateToPage(
        1,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeInOut,
      );
      
      // Haptic feedback
      HapticFeedback.lightImpact();
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error loading item details: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }
  
  void _onQuantityChanged(double quantity) {
    setState(() {
      _quantity = quantity;
    });
  }
  
  void _executeAction() async {
    if (_selectedItem == null || (widget.workflowType != WorkflowType.delete && _quantity <= 0)) {
      return;
    }
    
    setState(() {
      _isLoading = true;
    });
    
    try {
      switch (widget.workflowType) {
        case WorkflowType.add:
          await _addItemToOrder();
          break;
        case WorkflowType.edit:
          await _editItemInOrder();
          break;
        case WorkflowType.delete:
          await _deleteItemFromOrder();
          break;
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error: $e'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }
  
  Future<void> _addItemToOrder() async {
    final line = PurchaseOrderLine(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      poId: widget.purchaseOrder.id,
      itemId: _selectedItem!.id,
      itemName: _selectedItem!.name,
      itemSku: _selectedItem!.sku,
      barcode: _selectedItem!.barcode,
      qtyOrdered: _quantity.toInt(),
      qtyReceived: 0,
      unitCost: _unitPrice,
    );
    
    // Call API to add item
    await _apiService.addItemToPurchaseOrder(
      widget.purchaseOrder.id,
      _selectedItem!.id,
      _quantity.toInt(),
      _unitPrice,
    );
    
    widget.onItemAdded(line);
    
    // Provide haptic feedback and success message
    HapticFeedback.mediumImpact();
    _showSuccessMessage('${_selectedItem!.name} added to purchase order');
    
    Navigator.pop(context);
  }
  
  Future<void> _editItemInOrder() async {
    if (widget.existingLine == null) return;
    
    final updatedLine = PurchaseOrderLine(
      id: widget.existingLine!.id,
      poId: widget.purchaseOrder.id,
      itemId: _selectedItem!.id,
      itemName: _selectedItem!.name,
      itemSku: _selectedItem!.sku,
      barcode: _selectedItem!.barcode,
      qtyOrdered: _quantity.toInt(),
      qtyReceived: widget.existingLine!.qtyReceived,
      unitCost: _unitPrice,
    );
    
    // Call API to update item (if quantity changed)
    if (_quantity.toInt() != widget.existingLine!.qtyOrdered) {
      await _apiService.updatePurchaseOrderLineQuantity(
        widget.purchaseOrder.id,
        widget.existingLine!.id,
        _quantity.toInt(),
      );
    }
    
    widget.onItemUpdated(widget.existingLine!, updatedLine);
    
    // Provide haptic feedback and success message
    HapticFeedback.mediumImpact();
    _showSuccessMessage('${_selectedItem!.name} updated in purchase order');
    
    Navigator.pop(context);
  }
  
  Future<void> _deleteItemFromOrder() async {
    if (widget.existingLine == null) return;
    
    // Show confirmation dialog
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Row(
          children: [
            Icon(Icons.warning, color: Colors.orange),
            SizedBox(width: 8),
            Text('Confirm Removal'),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Are you sure you want to remove this item from the purchase order?'),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.grey.shade100,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _selectedItem!.name,
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                  Text('SKU: ${_selectedItem!.sku}'),
                  Text('Quantity: ${_quantity.toInt()}'),
                  Text('Total: \$${(_unitPrice * _quantity).toStringAsFixed(2)}'),
                ],
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            style: FilledButton.styleFrom(
              backgroundColor: Colors.red,
            ),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
    
    if (confirmed != true) return;
    
    // Call API to remove item
    await _apiService.removeItemFromPurchaseOrder(
      widget.purchaseOrder.id,
      widget.existingLine!.id,
    );
    
    widget.onItemDeleted(widget.existingLine!);
    
    // Provide haptic feedback and success message
    HapticFeedback.mediumImpact();
    _showSuccessMessage('${_selectedItem!.name} removed from purchase order');
    
    Navigator.pop(context);
  }
  
  void _showSuccessMessage(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.check_circle, color: Colors.white),
            const SizedBox(width: 8),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: Colors.green,
        behavior: SnackBarBehavior.floating,
        margin: const EdgeInsets.all(16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }
  
  Widget _buildSearchPhase() {
    return Column(
      children: [
        // Header
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.primaryContainer,
            borderRadius: const BorderRadius.only(
              topLeft: Radius.circular(20),
              topRight: Radius.circular(20),
            ),
          ),
          child: Row(
            children: [
              IconButton(
                onPressed: () => Navigator.pop(context),
                icon: const Icon(Icons.close),
              ),
              Expanded(
                child: Text(
                  _workflowTitle,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
              const SizedBox(width: 48), // Balance the close button
            ],
          ),
        ),
        
        // Purchase Order Context
        Container(
          margin: const EdgeInsets.all(16),
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.blue.shade50,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.blue.shade200),
          ),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.blue,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.shopping_cart,
                  color: Colors.white,
                  size: 20,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'PO #${widget.purchaseOrder.number}',
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 16,
                      ),
                    ),
                    if (widget.purchaseOrder.supplierName != null)
                      Text(
                        widget.purchaseOrder.supplierName!,
                        style: TextStyle(
                          color: Colors.blue.shade700,
                          fontSize: 14,
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
        
        // Search component
        Expanded(
          child: UniversalItemSearch(
            onItemSelected: _onItemSelected,
            availableModes: const [
              SearchMode.text,
              SearchMode.barcode,
              SearchMode.description,
            ],
            showStockInfo: false,
            hintText: widget.workflowType == WorkflowType.add 
              ? 'Search items to add to PO'
              : 'Search for the item to ${widget.workflowType.name}',
          ),
        ),
      ],
    );
  }
  
  Widget _buildQuantityPhase() {
    if (_selectedItem == null) {
      return const Center(
        child: Text('No item selected'),
      );
    }
    
    final totalPrice = _unitPrice * _quantity;
    final isDeleteMode = widget.workflowType == WorkflowType.delete;
    
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Header with back button
          Row(
            children: [
              IconButton(
                onPressed: () {
                  _pageController.animateToPage(
                    0,
                    duration: const Duration(milliseconds: 300),
                    curve: Curves.easeInOut,
                  );
                },
                icon: const Icon(Icons.arrow_back),
              ),
              Expanded(
                child: Text(
                  _workflowTitle,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
              IconButton(
                onPressed: () => Navigator.pop(context),
                icon: const Icon(Icons.close),
              ),
            ],
          ),
          const SizedBox(height: 24),
          
          // Selected item card
          Card(
            elevation: 2,
            child: Container(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                border: isDeleteMode 
                  ? Border.all(color: Colors.red.shade300, width: 2)
                  : null,
                color: isDeleteMode ? Colors.red.shade50 : null,
              ),
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Warning banner for delete mode
                  if (isDeleteMode)
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(12),
                      margin: const EdgeInsets.only(bottom: 16),
                      decoration: BoxDecoration(
                        color: Colors.red.shade100,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.red.shade300),
                      ),
                      child: const Row(
                        children: [
                          Icon(Icons.warning, color: Colors.red),
                          SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              'You are about to remove this item from the purchase order',
                              style: TextStyle(
                                color: Colors.red,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  
                  // Item header
                  Row(
                    children: [
                      Container(
                        width: 60,
                        height: 60,
                        decoration: BoxDecoration(
                          color: isDeleteMode 
                            ? Colors.red.shade100
                            : Theme.of(context).colorScheme.primaryContainer,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Icon(
                          Icons.inventory_2,
                          size: 32,
                          color: isDeleteMode 
                            ? Colors.red
                            : Theme.of(context).colorScheme.primary,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _selectedItem!.sku,
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.bold,
                                color: isDeleteMode ? Colors.red.shade700 : Colors.grey,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              _selectedItem!.name,
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                color: isDeleteMode ? Colors.red.shade800 : null,
                              ),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                            if (_selectedItem!.barcode != null)
                              Text(
                                'Barcode: ${_selectedItem!.barcode}',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: isDeleteMode ? Colors.red.shade600 : Colors.grey[600],
                                ),
                              ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 24),
                  
                  // Current quantity info for edit/delete
                  if (widget.workflowType != WorkflowType.add && widget.existingLine != null)
                    Container(
                      padding: const EdgeInsets.all(12),
                      margin: const EdgeInsets.only(bottom: 16),
                      decoration: BoxDecoration(
                        color: Colors.blue.shade50,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text(
                            'Current Quantity:',
                            style: TextStyle(fontWeight: FontWeight.bold),
                          ),
                          Text(
                            '${widget.existingLine!.qtyOrdered} units',
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                    ),
                  
                  // Unit price
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Unit Price:',
                        style: TextStyle(fontSize: 16),
                      ),
                      Text(
                        '\$${_unitPrice.toStringAsFixed(2)}',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: isDeleteMode 
                            ? Colors.red.shade700
                            : Theme.of(context).colorScheme.primary,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),
          
          // Quantity input (disabled for delete mode)
          if (!isDeleteMode)
            Card(
              elevation: 2,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: QuantityInputWidget(
                  initialQuantity: _quantity,
                  onQuantityChanged: _onQuantityChanged,
                  label: widget.workflowType == WorkflowType.edit 
                    ? 'Update Quantity'
                    : 'Quantity to Order',
                  unit: _selectedItem!.uom ?? 'PCS',
                  showQuickButtons: true,
                  quickButtonValues: const [1, 5, 10, 25, 50, 100],
                  decimalPlaces: 0,
                  showSteppers: true,
                  stepAmount: 1,
                  minQuantity: 1,
                  isLarge: true,
                  primaryColor: Theme.of(context).colorScheme.primary,
                ),
              ),
            ),
          
          if (!isDeleteMode) const SizedBox(height: 24),
          
          // Total price calculation (show even in delete mode)
          Card(
            color: isDeleteMode 
              ? Colors.red.shade50
              : Theme.of(context).colorScheme.secondaryContainer,
            elevation: 0,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        isDeleteMode ? 'Value to Remove:' : 'Total Price:',
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      Text(
                        '\$${totalPrice.toStringAsFixed(2)}',
                        style: TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          color: isDeleteMode 
                            ? Colors.red.shade700
                            : Theme.of(context).colorScheme.primary,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    '${_quantity.toInt()} ${_selectedItem!.uom ?? "units"} × \$${_unitPrice.toStringAsFixed(2)}',
                    style: TextStyle(
                      fontSize: 14,
                      color: isDeleteMode 
                        ? Colors.red.shade600
                        : Theme.of(context).colorScheme.onSecondaryContainer.withOpacity(0.7),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 32),
          
          // Action button
          FilledButton.icon(
            onPressed: _isLoading ? null : _executeAction,
            icon: _isLoading 
              ? const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                )
              : Icon(_actionButtonIcon),
            label: Text(
              _isLoading ? 'Processing...' : _actionButtonText,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            style: FilledButton.styleFrom(
              backgroundColor: _actionButtonColor,
              padding: const EdgeInsets.all(16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
          ),
        ],
      ),
    );
  }
  
  @override
  Widget build(BuildContext context) {
    return Container(
      height: MediaQuery.of(context).size.height * 0.9,
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.only(
          topLeft: Radius.circular(20),
          topRight: Radius.circular(20),
        ),
      ),
      child: PageView(
        controller: _pageController,
        physics: const NeverScrollableScrollPhysics(),
        onPageChanged: (page) {
          setState(() {
            _currentPage = page;
          });
        },
        children: [
          _buildSearchPhase(),
          _buildQuantityPhase(),
        ],
      ),
    );
  }
}