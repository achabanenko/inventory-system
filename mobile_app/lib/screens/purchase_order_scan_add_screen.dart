import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/item.dart';
import '../models/purchase_order.dart';
import '../utils/scanner_utils.dart';

class PurchaseOrderScanAddScreen extends StatefulWidget {
  final PurchaseOrder purchaseOrder;
  final Function(PurchaseOrderLine) onItemAdded;
  
  const PurchaseOrderScanAddScreen({
    super.key,
    required this.purchaseOrder,
    required this.onItemAdded,
  });

  @override
  State<PurchaseOrderScanAddScreen> createState() => _PurchaseOrderScanAddScreenState();
}

class _PurchaseOrderScanAddScreenState extends State<PurchaseOrderScanAddScreen> {
  
  void _onAddItemToOrder(Item item, double quantity) {
    // Create purchase order line
    final line = PurchaseOrderLine(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      poId: widget.purchaseOrder.id,
      itemId: item.id,
      itemName: item.name,
      itemSku: item.sku,
      barcode: item.barcode,
      qtyOrdered: quantity.toInt(),
      qtyReceived: 0,
      unitCost: item.price,
    );
    
    // Add to purchase order
    widget.onItemAdded(line);
    
    // Provide haptic feedback
    HapticFeedback.mediumImpact();
    
    // Show success message
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.check_circle, color: Colors.white),
            const SizedBox(width: 8),
            Expanded(
              child: Text('${item.name} added to purchase order'),
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
  }
  
  void _startScanning() {
    ScannerUtils.showScannerWithItemDetail(
      context: context,
      documentType: 'Purchase Order',
      onAddToDocument: _onAddItemToOrder,
      contextData: {
        'po_number': widget.purchaseOrder.number,
        'supplier': widget.purchaseOrder.supplierName ?? 'Unknown Supplier',
      },
      headerText: 'Scan Item for PO ${widget.purchaseOrder.number}',
      contextWidget: Container(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.shopping_cart, color: Colors.white, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'PO #${widget.purchaseOrder.number}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            if (widget.purchaseOrder.supplierName != null)
              Padding(
                padding: const EdgeInsets.only(top: 4),
                child: Text(
                  widget.purchaseOrder.supplierName!,
                  style: const TextStyle(
                    color: Colors.white70,
                    fontSize: 12,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade50,
      appBar: AppBar(
        title: Text('Add Items - PO ${widget.purchaseOrder.number}'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black,
        elevation: 0,
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Large scan button
            Container(
              width: 200,
              height: 200,
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primary,
                borderRadius: BorderRadius.circular(100),
                boxShadow: [
                  BoxShadow(
                    color: Theme.of(context).colorScheme.primary.withOpacity(0.3),
                    blurRadius: 30,
                    offset: const Offset(0, 10),
                  ),
                ],
              ),
              child: Material(
                color: Colors.transparent,
                child: InkWell(
                  onTap: _startScanning,
                  borderRadius: BorderRadius.circular(100),
                  child: const Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.qr_code_scanner,
                        size: 80,
                        color: Colors.white,
                      ),
                      SizedBox(height: 12),
                      Text(
                        'SCAN',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          letterSpacing: 2,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            
            const SizedBox(height: 40),
            
            // Instruction text
            Text(
              'Tap to scan item barcodes',
              style: TextStyle(
                fontSize: 18,
                color: Colors.grey.shade600,
                fontWeight: FontWeight.w500,
              ),
            ),
            
            const SizedBox(height: 8),
            
            Text(
              'Each scanned item will show details before adding',
              style: TextStyle(
                fontSize: 14,
                color: Colors.grey.shade500,
              ),
              textAlign: TextAlign.center,
            ),
            
            const SizedBox(height: 60),
            
            // Purchase order info card
            Container(
              margin: const EdgeInsets.symmetric(horizontal: 32),
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.05),
                    blurRadius: 10,
                    offset: const Offset(0, 5),
                  ),
                ],
              ),
              child: Column(
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: Theme.of(context).colorScheme.primaryContainer,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Icon(
                          Icons.shopping_cart,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Purchase Order',
                              style: TextStyle(
                                fontSize: 12,
                                color: Colors.grey.shade600,
                              ),
                            ),
                            Text(
                              widget.purchaseOrder.number,
                              style: const TextStyle(
                                fontSize: 18,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  
                  if (widget.purchaseOrder.supplierName != null) ...[
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Icon(
                          Icons.business,
                          size: 16,
                          color: Colors.grey.shade600,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            widget.purchaseOrder.supplierName!,
                            style: TextStyle(
                              color: Colors.grey.shade700,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                  
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Icon(
                        Icons.list_alt,
                        size: 16,
                        color: Colors.grey.shade600,
                      ),
                      const SizedBox(width: 8),
                      Text(
                        '${widget.purchaseOrder.lines.length} items already added',
                        style: TextStyle(
                          color: Colors.grey.shade700,
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
  }
}