import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/item.dart';
import '../models/purchase_order.dart';
import '../widgets/universal_item_search.dart';
import '../widgets/quantity_input_widget.dart';
import '../widgets/barcode_scanner_overlay.dart';

class PurchaseOrderAddItemScreen extends StatefulWidget {
  final PurchaseOrder purchaseOrder;
  final Function(PurchaseOrderLine) onItemAdded;
  
  const PurchaseOrderAddItemScreen({
    super.key,
    required this.purchaseOrder,
    required this.onItemAdded,
  });

  @override
  State<PurchaseOrderAddItemScreen> createState() => _PurchaseOrderAddItemScreenState();
}

class _PurchaseOrderAddItemScreenState extends State<PurchaseOrderAddItemScreen> {
  Item? _selectedItem;
  double _quantity = 1;
  double _unitPrice = 0;
  final PageController _pageController = PageController();
  int _currentPage = 0;
  
  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }
  
  void _onItemSelected(Item item) {
    setState(() {
      _selectedItem = item;
      _unitPrice = item.price ?? 0;
      _quantity = 1;
      _currentPage = 1;
    });
    
    // Animate to quantity page
    _pageController.animateToPage(
      1,
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeInOut,
    );
    
    // Haptic feedback
    HapticFeedback.lightImpact();
  }
  
  void _onQuantityChanged(double quantity) {
    setState(() {
      _quantity = quantity;
    });
  }
  
  void _addItemToOrder() {
    if (_selectedItem == null || _quantity <= 0) return;
    
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
    
    widget.onItemAdded(line);
    
    // Haptic feedback
    HapticFeedback.mediumImpact();
    
    // Show success and close
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.check_circle, color: Colors.white),
            const SizedBox(width: 8),
            Expanded(
              child: Text('${_selectedItem!.name} added to order'),
            ),
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
    
    Navigator.pop(context);
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
              const Expanded(
                child: Text(
                  'Add Item to Purchase Order',
                  style: TextStyle(
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
            hintText: 'Search items to add to PO',
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
              const Expanded(
                child: Text(
                  'Set Quantity',
                  style: TextStyle(
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
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Item header
                  Row(
                    children: [
                      Container(
                        width: 60,
                        height: 60,
                        decoration: BoxDecoration(
                          color: Theme.of(context).colorScheme.primaryContainer,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Icon(
                          Icons.inventory_2,
                          size: 32,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _selectedItem!.sku,
                              style: const TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.bold,
                                color: Colors.grey,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              _selectedItem!.name,
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                            if (_selectedItem!.barcode != null)
                              Text(
                                'Barcode: ${_selectedItem!.barcode}',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: Colors.grey[600],
                                ),
                              ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 24),
                  
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
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),
          
          // Quantity input
          Card(
            elevation: 2,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: QuantityInputWidget(
                initialQuantity: _quantity,
                onQuantityChanged: _onQuantityChanged,
                label: 'Quantity to Order',
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
          const SizedBox(height: 24),
          
          // Total price calculation
          Card(
            color: Theme.of(context).colorScheme.secondaryContainer,
            elevation: 0,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Total Price:',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      Text(
                        '\$${totalPrice.toStringAsFixed(2)}',
                        style: TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    '${_quantity.toInt()} ${_selectedItem!.uom ?? "units"} × \$${_unitPrice.toStringAsFixed(2)}',
                    style: TextStyle(
                      fontSize: 14,
                      color: Theme.of(context).colorScheme.onSecondaryContainer.withOpacity(0.7),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 32),
          
          // Add to order button
          FilledButton.icon(
            onPressed: _addItemToOrder,
            icon: const Icon(Icons.add_shopping_cart),
            label: const Text(
              'Add to Order',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            style: FilledButton.styleFrom(
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
      height: MediaQuery.of(context).size.height * 0.85,
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