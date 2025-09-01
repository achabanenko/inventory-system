import 'package:flutter/material.dart';
import 'package:flutter_barcode_scanner/flutter_barcode_scanner.dart';
import 'package:provider/provider.dart';
import '../models/purchase_order.dart';
import '../models/item.dart';
import '../services/api_service.dart';

class GoodsReceiptScreen extends StatefulWidget {
  final PurchaseOrder purchaseOrder;

  const GoodsReceiptScreen({
    super.key,
    required this.purchaseOrder,
  });

  @override
  State<GoodsReceiptScreen> createState() => _GoodsReceiptScreenState();
}

class _GoodsReceiptScreenState extends State<GoodsReceiptScreen> {
  late List<ReceiptLine> _receiptLines;
  bool _isProcessing = false;
  final TextEditingController _searchController = TextEditingController();
  List<Item> _searchResults = [];
  bool _isSearching = false;

  @override
  void initState() {
    super.initState();
    _initializeReceiptLines();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _initializeReceiptLines() {
    _receiptLines = widget.purchaseOrder.lines.map((line) {
      return ReceiptLine(
        poLineId: line.id,
        itemId: line.itemId,
        itemName: line.itemName ?? 'Unknown Item',
        itemSku: line.itemSku ?? 'N/A',
        barcode: line.barcode,
        qtyOrdered: line.qtyOrdered,
        qtyReceived: 0,
        qtyPreviouslyReceived: line.qtyReceived,
        unitCost: line.unitCost,
      );
    }).toList();
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
      final items = await apiService.getItems(search: query, limit: 10);
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

      // Find the line with this barcode
      final lineIndex = _receiptLines.indexWhere((line) => line.barcode == barcode);

      if (lineIndex >= 0) {
        final line = _receiptLines[lineIndex];
        _showReceiveDialog(lineIndex, line);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Barcode $barcode not found in this purchase order'),
              backgroundColor: Colors.orange,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Scan failed: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  void _showReceiveDialog(int index, ReceiptLine line) {
    final TextEditingController qtyController = TextEditingController(
      text: '1'
    );

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Receive Item'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              line.itemName,
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
            Text(
              'SKU: ${line.itemSku}',
              style: TextStyle(color: Colors.grey[600], fontSize: 12),
            ),
            const SizedBox(height: 16),
            Text(
              'Ordered: ${line.qtyOrdered} • Previously Received: ${line.qtyPreviouslyReceived} • Remaining: ${line.qtyRemaining}',
              style: const TextStyle(fontSize: 12),
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: qtyController,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(
                labelText: 'Quantity to Receive',
                border: OutlineInputBorder(),
              ),
              validator: (value) {
                if (value == null || value.isEmpty) return 'Please enter quantity';
                final qty = int.tryParse(value);
                if (qty == null || qty <= 0) return 'Please enter valid quantity';
                if (qty > line.qtyRemaining) return 'Cannot receive more than remaining quantity';
                return null;
              },
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              final qty = int.tryParse(qtyController.text);
              if (qty != null && qty > 0 && qty <= line.qtyRemaining) {
                _updateReceivedQuantity(index, qty);
                Navigator.pop(context);
              }
            },
            child: const Text('Receive'),
          ),
        ],
      ),
    );
  }

  void _updateReceivedQuantity(int index, int quantity) {
    setState(() {
      _receiptLines[index] = _receiptLines[index].copyWith(
        qtyReceived: _receiptLines[index].qtyReceived + quantity,
      );
    });

    final line = _receiptLines[index];
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Received ${quantity}x ${line.itemName}'),
        action: SnackBarAction(
          label: 'Undo',
          onPressed: () {
            setState(() {
              _receiptLines[index] = _receiptLines[index].copyWith(
                qtyReceived: _receiptLines[index].qtyReceived - quantity,
              );
            });
          },
        ),
      ),
    );
  }

  void _quickReceive(int index, ReceiptLine line) {
    final remaining = line.qtyRemaining;
    if (remaining > 0) {
      _updateReceivedQuantity(index, remaining);
    }
  }

  Future<void> _submitReceipt() async {
    final hasReceipts = _receiptLines.any((line) => line.qtyReceived > 0);

    if (!hasReceipts) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please receive at least one item'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    setState(() => _isProcessing = true);

    try {
      final apiService = Provider.of<ApiService>(context, listen: false);

      // Prepare receipt data
      final receipts = _receiptLines
          .where((line) => line.qtyReceived > 0)
          .map((line) => {
                'line_id': line.poLineId,
                'qty_received': line.qtyReceived,
                'occurred_at': DateTime.now().toIso8601String(),
              })
          .toList();

      await apiService.receivePurchaseOrder(widget.purchaseOrder.id, receipts);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Items received successfully!'),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true); // Return success
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error receiving items: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      setState(() => _isProcessing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final totalItems = _receiptLines.length;
    final receivedItems = _receiptLines.where((line) => line.qtyReceived > 0).length;
    final totalReceived = _receiptLines.fold<int>(0, (sum, line) => sum + line.qtyReceived);

    return Scaffold(
      appBar: AppBar(
        title: Text('Receive PO ${widget.purchaseOrder.number}'),
        actions: [
          TextButton(
            onPressed: _isProcessing ? null : _submitReceipt,
            child: _isProcessing
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Submit', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
      body: Column(
        children: [
          // Header with PO info and progress
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
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'PO ${widget.purchaseOrder.number}',
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          Text(
                            widget.purchaseOrder.supplierName ?? 'Unknown Supplier',
                            style: TextStyle(
                              color: Colors.grey[600],
                              fontSize: 14,
                            ),
                          ),
                        ],
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: widget.purchaseOrder.status.color.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        widget.purchaseOrder.status.displayName,
                        style: TextStyle(
                          color: widget.purchaseOrder.status.color,
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        '$receivedItems/$totalItems items • $totalReceived received',
                        style: const TextStyle(fontWeight: FontWeight.w500),
                      ),
                    ),
                    ElevatedButton.icon(
                      onPressed: _scanBarcode,
                      icon: const Icon(Icons.qr_code_scanner, size: 16),
                      label: const Text('Scan'),
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Receipt Lines
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _receiptLines.length,
              itemBuilder: (context, index) {
                final line = _receiptLines[index];
                final progressValue = line.qtyOrdered > 0
                    ? (line.qtyPreviouslyReceived + line.qtyReceived) / line.qtyOrdered
                    : 0.0;

                return Card(
                  margin: const EdgeInsets.only(bottom: 8),
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Item info
                        Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    line.itemName,
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  Text(
                                    'SKU: ${line.itemSku}',
                                    style: TextStyle(
                                      color: Colors.grey[600],
                                      fontSize: 12,
                                    ),
                                  ),
                                  if (line.barcode != null)
                                    Text(
                                      'Barcode: ${line.barcode}',
                                      style: TextStyle(
                                        color: Colors.grey[600],
                                        fontSize: 12,
                                        fontFamily: 'monospace',
                                      ),
                                    ),
                                ],
                              ),
                            ),
                            if (line.barcode != null)
                              const Icon(Icons.qr_code, color: Colors.grey),
                          ],
                        ),

                        const SizedBox(height: 12),

                        // Progress and quantities
                        Column(
                          children: [
                            Row(
                              children: [
                                const Icon(Icons.inventory_2, size: 16, color: Colors.grey),
                                const SizedBox(width: 4),
                                Expanded(
                                  child: Text(
                                    'Ordered: ${line.qtyOrdered} • Previously: ${line.qtyPreviouslyReceived} • Remaining: ${line.qtyRemaining}',
                                    style: const TextStyle(fontSize: 12),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 6),
                            LinearProgressIndicator(
                              value: progressValue,
                              backgroundColor: Colors.grey[200],
                              valueColor: AlwaysStoppedAnimation<Color>(
                                progressValue >= 1.0 ? Colors.green : Colors.blue,
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 12),

                        // Actions
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                line.qtyReceived > 0
                                    ? 'Will receive: ${line.qtyReceived} items'
                                    : 'No items to receive',
                                style: TextStyle(
                                  color: line.qtyReceived > 0 ? Colors.green : Colors.grey,
                                  fontWeight: FontWeight.w500,
                                ),
                              ),
                            ),
                            if (line.qtyRemaining > 0) ...[
                              TextButton.icon(
                                onPressed: () => _showReceiveDialog(index, line),
                                icon: const Icon(Icons.add, size: 16),
                                label: const Text('Receive'),
                                style: TextButton.styleFrom(
                                  foregroundColor: const Color(0xFF059669),
                                ),
                              ),
                              const SizedBox(width: 8),
                              TextButton(
                                onPressed: () => _quickReceive(index, line),
                                child: const Text('Receive All'),
                                style: TextButton.styleFrom(
                                  foregroundColor: const Color(0xFF059669),
                                ),
                              ),
                            ],
                          ],
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),

          // Summary footer
          if (totalReceived > 0)
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
                      'Total: $totalReceived items from $receivedItems lines',
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 16,
                      ),
                    ),
                  ),
                  ElevatedButton(
                    onPressed: _isProcessing ? null : _submitReceipt,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF059669),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                    ),
                    child: _isProcessing
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('Submit Receipt'),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class ReceiptLine {
  final String poLineId;
  final String itemId;
  final String itemName;
  final String itemSku;
  final String? barcode;
  final int qtyOrdered;
  final int qtyPreviouslyReceived;
  int qtyReceived;
  final double unitCost;

  ReceiptLine({
    required this.poLineId,
    required this.itemId,
    required this.itemName,
    required this.itemSku,
    this.barcode,
    required this.qtyOrdered,
    required this.qtyPreviouslyReceived,
    required this.qtyReceived,
    required this.unitCost,
  });

  int get qtyRemaining => qtyOrdered - qtyPreviouslyReceived - qtyReceived;

  double get totalValue => qtyReceived * unitCost;

  ReceiptLine copyWith({
    int? qtyReceived,
  }) {
    return ReceiptLine(
      poLineId: poLineId,
      itemId: itemId,
      itemName: itemName,
      itemSku: itemSku,
      barcode: barcode,
      qtyOrdered: qtyOrdered,
      qtyPreviouslyReceived: qtyPreviouslyReceived,
      qtyReceived: qtyReceived ?? this.qtyReceived,
      unitCost: unitCost,
    );
  }
}
