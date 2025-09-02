import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/item.dart';
import '../models/purchase_order.dart';

class ScannedItemDetailCard extends StatefulWidget {
  final Item item;
  final double initialQuantity;
  final Function(double) onQuantityChanged;
  final VoidCallback onAddToDocument;
  final VoidCallback onScanAnother;
  final String documentType; // 'PO', 'Receipt', 'Transfer', 'Count'
  final Map<String, dynamic>? contextData; // PO info, stock levels, etc.
  
  const ScannedItemDetailCard({
    super.key,
    required this.item,
    this.initialQuantity = 1,
    required this.onQuantityChanged,
    required this.onAddToDocument,
    required this.onScanAnother,
    this.documentType = 'PO',
    this.contextData,
  });

  @override
  State<ScannedItemDetailCard> createState() => _ScannedItemDetailCardState();
}

class _ScannedItemDetailCardState extends State<ScannedItemDetailCard>
    with TickerProviderStateMixin {
  late double _quantity;
  late AnimationController _slideController;
  late Animation<Offset> _slideAnimation;
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;
  
  @override
  void initState() {
    super.initState();
    _quantity = widget.initialQuantity;
    
    // Slide-in animation
    _slideController = AnimationController(
      duration: const Duration(milliseconds: 400),
      vsync: this,
    );
    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, 1),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _slideController,
      curve: Curves.easeOutCubic,
    ));
    
    // Pulse animation for scan success
    _pulseController = AnimationController(
      duration: const Duration(milliseconds: 600),
      vsync: this,
    );
    _pulseAnimation = Tween<double>(
      begin: 1.0,
      end: 1.1,
    ).animate(CurvedAnimation(
      parent: _pulseController,
      curve: Curves.elasticOut,
    ));
    
    // Start animations
    _slideController.forward();
    _pulseController.forward().then((_) {
      _pulseController.reverse();
    });
  }
  
  @override
  void dispose() {
    _slideController.dispose();
    _pulseController.dispose();
    super.dispose();
  }
  
  void _updateQuantity(double newQuantity) {
    setState(() {
      _quantity = newQuantity;
    });
    widget.onQuantityChanged(newQuantity);
    HapticFeedback.lightImpact();
  }
  
  void _incrementQuantity() {
    _updateQuantity(_quantity + 1);
  }
  
  void _decrementQuantity() {
    if (_quantity > 1) {
      _updateQuantity(_quantity - 1);
    }
  }
  
  Widget _buildContextInfo() {
    if (widget.contextData == null) return const SizedBox();
    
    switch (widget.documentType) {
      case 'Receipt':
        final poQty = widget.contextData!['po_quantity'] as int? ?? 0;
        final received = widget.contextData!['already_received'] as int? ?? 0;
        final remaining = poQty - received;
        
        return Container(
          padding: const EdgeInsets.all(12),
          margin: const EdgeInsets.only(bottom: 16),
          decoration: BoxDecoration(
            color: Colors.blue.shade50,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.blue.shade200),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Purchase Order Context',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: Colors.blue.shade800,
                ),
              ),
              const SizedBox(height: 4),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Ordered: $poQty'),
                  Text('Received: $received'),
                  Text(
                    'Remaining: $remaining',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: remaining > 0 ? Colors.green : Colors.orange,
                    ),
                  ),
                ],
              ),
            ],
          ),
        );
        
      case 'Transfer':
        final availableStock = widget.contextData!['available_stock'] as int? ?? 0;
        
        return Container(
          padding: const EdgeInsets.all(12),
          margin: const EdgeInsets.only(bottom: 16),
          decoration: BoxDecoration(
            color: Colors.orange.shade50,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.orange.shade200),
          ),
          child: Row(
            children: [
              Icon(Icons.inventory, color: Colors.orange.shade700),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  'Available Stock: $availableStock units',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: Colors.orange.shade800,
                  ),
                ),
              ),
            ],
          ),
        );
        
      case 'Count':
        final systemQty = widget.contextData!['system_quantity'] as int? ?? 0;
        final variance = _quantity.toInt() - systemQty;
        
        return Container(
          padding: const EdgeInsets.all(12),
          margin: const EdgeInsets.only(bottom: 16),
          decoration: BoxDecoration(
            color: variance == 0 
              ? Colors.green.shade50 
              : (variance > 0 ? Colors.blue.shade50 : Colors.red.shade50),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: variance == 0 
                ? Colors.green.shade200 
                : (variance > 0 ? Colors.blue.shade200 : Colors.red.shade200),
            ),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('System: $systemQty'),
                  Text('Counted: ${_quantity.toInt()}'),
                  Text(
                    'Variance: ${variance >= 0 ? '+' : ''}$variance',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: variance == 0 
                        ? Colors.green.shade700
                        : (variance > 0 ? Colors.blue.shade700 : Colors.red.shade700),
                    ),
                  ),
                ],
              ),
            ],
          ),
        );
        
      default:
        return const SizedBox();
    }
  }
  
  @override
  Widget build(BuildContext context) {
    final totalPrice = widget.item.price * _quantity;
    
    return SlideTransition(
      position: _slideAnimation,
      child: AnimatedBuilder(
        animation: _pulseAnimation,
        builder: (context, child) {
          return Transform.scale(
            scale: _pulseAnimation.value,
            child: Container(
              margin: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    blurRadius: 20,
                    offset: const Offset(0, 10),
                  ),
                ],
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Success header
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: Colors.green.shade50,
                      borderRadius: const BorderRadius.only(
                        topLeft: Radius.circular(16),
                        topRight: Radius.circular(16),
                      ),
                      border: Border(
                        bottom: BorderSide(color: Colors.green.shade200),
                      ),
                    ),
                    child: Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.green,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: const Icon(
                            Icons.qr_code_scanner,
                            color: Colors.white,
                            size: 24,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'Item Scanned Successfully!',
                                style: TextStyle(
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600,
                                  color: Colors.green,
                                ),
                              ),
                              Text(
                                'Ready to add to ${widget.documentType}',
                                style: TextStyle(
                                  fontSize: 12,
                                  color: Colors.green.shade700,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                  
                  // Item details
                  Padding(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Item info
                        Row(
                          children: [
                            Container(
                              width: 60,
                              height: 60,
                              decoration: BoxDecoration(
                                color: Theme.of(context).colorScheme.primaryContainer,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Icon(
                                Icons.inventory_2,
                                size: 32,
                                color: Theme.of(context).colorScheme.primary,
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    widget.item.sku,
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                      color: Colors.grey,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    widget.item.name,
                                    style: const TextStyle(
                                      fontSize: 18,
                                      fontWeight: FontWeight.bold,
                                    ),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  if (widget.item.barcode != null)
                                    Text(
                                      'Barcode: ${widget.item.barcode}',
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
                        const SizedBox(height: 20),
                        
                        // Context information
                        _buildContextInfo(),
                        
                        // Quantity section
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: Colors.grey.shade50,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'Quantity',
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.grey.shade700,
                                ),
                              ),
                              const SizedBox(height: 12),
                              
                              // Quantity controls
                              Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  // Minus button
                                  Material(
                                    color: _quantity > 1 
                                      ? Theme.of(context).colorScheme.primary
                                      : Colors.grey.shade300,
                                    borderRadius: BorderRadius.circular(12),
                                    child: InkWell(
                                      onTap: _quantity > 1 ? _decrementQuantity : null,
                                      borderRadius: BorderRadius.circular(12),
                                      child: Container(
                                        width: 50,
                                        height: 50,
                                        alignment: Alignment.center,
                                        child: Icon(
                                          Icons.remove,
                                          size: 24,
                                          color: _quantity > 1 ? Colors.white : Colors.grey,
                                        ),
                                      ),
                                    ),
                                  ),
                                  
                                  // Quantity display
                                  Container(
                                    width: 100,
                                    height: 50,
                                    margin: const EdgeInsets.symmetric(horizontal: 16),
                                    alignment: Alignment.center,
                                    decoration: BoxDecoration(
                                      border: Border.all(color: Colors.grey.shade300),
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                    child: Text(
                                      '${_quantity.toInt()}',
                                      style: const TextStyle(
                                        fontSize: 24,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ),
                                  
                                  // Plus button
                                  Material(
                                    color: Theme.of(context).colorScheme.primary,
                                    borderRadius: BorderRadius.circular(12),
                                    child: InkWell(
                                      onTap: _incrementQuantity,
                                      borderRadius: BorderRadius.circular(12),
                                      child: Container(
                                        width: 50,
                                        height: 50,
                                        alignment: Alignment.center,
                                        child: const Icon(
                                          Icons.add,
                                          size: 24,
                                          color: Colors.white,
                                        ),
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                              
                              // Unit and price info
                              const SizedBox(height: 12),
                              Row(
                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                children: [
                                  Text(
                                    'Unit: ${widget.item.uom ?? "PCS"}',
                                    style: TextStyle(
                                      color: Colors.grey.shade600,
                                    ),
                                  ),
                                  Text(
                                    'Price: \$${widget.item.price.toStringAsFixed(2)}',
                                    style: TextStyle(
                                      color: Colors.grey.shade600,
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                        
                        // Big Confirmation Button - moved above Total
                        const SizedBox(height: 20),
                        Container(
                          width: double.infinity,
                          height: 60,
                          child: FilledButton.icon(
                            onPressed: widget.onAddToDocument,
                            icon: const Icon(Icons.check_circle, size: 24),
                            label: Text(
                              'CONFIRM - Add to ${widget.documentType}',
                              style: const TextStyle(
                                fontSize: 18,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            style: FilledButton.styleFrom(
                              backgroundColor: Colors.green,
                              foregroundColor: Colors.white,
                              padding: const EdgeInsets.all(16),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(16),
                              ),
                            ),
                          ),
                        ),
                        
                        // Total price
                        const SizedBox(height: 16),
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: Theme.of(context).colorScheme.primaryContainer,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              const Text(
                                'Total:',
                                style: TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
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
                        ),
                        
                        // Secondary Actions
                        const SizedBox(height: 12),
                        Row(
                          children: [
                            Expanded(
                              child: OutlinedButton.icon(
                                onPressed: widget.onScanAnother,
                                icon: const Icon(Icons.qr_code_scanner, size: 18),
                                label: const Text('Scan Another'),
                                style: OutlinedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: TextButton.icon(
                                onPressed: () {
                                  Navigator.of(context).pop();
                                },
                                icon: const Icon(Icons.close, size: 18),
                                label: const Text('Cancel'),
                                style: TextButton.styleFrom(
                                  foregroundColor: Colors.grey[600],
                                  padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}