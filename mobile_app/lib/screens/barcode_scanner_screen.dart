import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_barcode_scanner/flutter_barcode_scanner.dart';
import 'package:provider/provider.dart';
import '../models/item.dart';
import '../models/purchase_order.dart';
import '../services/api_service.dart';
import 'purchase_order_create_screen.dart';
import 'goods_receipt_screen.dart';

enum BarcodeScanMode {
  lookup,      // Just find and display item info
  addToOrder,  // Add item to a new or existing purchase order
  receive,     // Receive items for goods receipt
}

class BarcodeScannerScreen extends StatefulWidget {
  const BarcodeScannerScreen({super.key});

  @override
  State<BarcodeScannerScreen> createState() => _BarcodeScannerScreenState();
}

class _BarcodeScannerScreenState extends State<BarcodeScannerScreen> {
  String _lastScannedCode = '';
  Item? _scannedItem;
  bool _isLoading = false;
  BarcodeScanMode _scanMode = BarcodeScanMode.lookup;

  Future<void> _scanBarcode() async {
    print('_scanBarcode called - starting scan');
    
    try {
      print('Calling FlutterBarcodeScanner.scanBarcode...');
      final barcode = await FlutterBarcodeScanner.scanBarcode(
        '#ff6666',
        'Cancel',
        true,
        ScanMode.BARCODE,
      );

      print('================================');
      print('BARCODE SCAN RESULT:');
      print('Raw result: "$barcode"');
      print('Length: ${barcode.length}');
      print('Type: ${barcode.runtimeType}');
      print('Is empty: ${barcode.isEmpty}');
      print('Is -1 (cancelled): ${barcode == '-1'}');
      print('================================');
      
      if (barcode == '-1') {
        print('User cancelled scan');
        return; // User cancelled
      }

      // Play success sound
      SystemSound.play(SystemSoundType.click);

      setState(() {
        _lastScannedCode = barcode;
        _isLoading = true;
        _scannedItem = null;
      });

      final apiService = Provider.of<ApiService>(context, listen: false);
      final item = await apiService.getItemByBarcode(barcode);

      setState(() {
        _scannedItem = item;
        _isLoading = false;
      });

      if (item == null && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('No item found for barcode: $barcode'),
            backgroundColor: Colors.orange,
          ),
        );
      } else if (item != null && mounted) {
        await _handleScanResult(item);
      }
    } catch (e) {
      print('Scan error: $e');
      print('Error type: ${e.runtimeType}');
      
      setState(() {
        _isLoading = false;
      });
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Scan failed: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _handleScanResult(Item item) async {
    switch (_scanMode) {
      case BarcodeScanMode.lookup:
        // Just show the item info (already handled in the scan method)
        break;
      case BarcodeScanMode.addToOrder:
        await _handleAddToOrder(item);
        break;
      case BarcodeScanMode.receive:
        await _handleReceiveItem(item);
        break;
    }
  }

  Future<void> _handleAddToOrder(Item item) async {
    // Show options: Create new PO or add to existing PO
    final result = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add to Purchase Order'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.add, color: Color(0xFF2563EB)),
              title: const Text('Create New PO'),
              subtitle: const Text('Start a new purchase order with this item'),
              onTap: () => Navigator.pop(context, 'new'),
            ),
            const Divider(),
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 8),
              child: Text(
                'Add to existing PO feature coming soon!',
                style: TextStyle(color: Colors.grey, fontSize: 12),
                textAlign: TextAlign.center,
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );

    if (result == 'new' && mounted) {
      // Navigate to PO creation screen with pre-selected item
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => PurchaseOrderCreateScreen(
            preSelectedItem: item,
          ),
        ),
      );
    }
  }

  Future<void> _handleReceiveItem(Item item) async {
    // Get approved POs that contain this item
    setState(() => _isLoading = true);

    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      final purchaseOrders = await apiService.getPurchaseOrders(
        status: 'APPROVED',
      );

      // Filter POs that contain this item
      final relevantPOs = purchaseOrders.where((po) =>
        po.lines.any((line) => line.itemId == item.id && line.qtyRemaining > 0)
      ).toList();

      setState(() => _isLoading = false);

      if (relevantPOs.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('No approved purchase orders found for this item'),
              backgroundColor: Colors.orange,
            ),
          );
        }
        return;
      }

      if (relevantPOs.length == 1) {
        // Only one PO, go directly to receipt
        if (mounted) {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => GoodsReceiptScreen(purchaseOrder: relevantPOs[0]),
            ),
          );
        }
      } else {
        // Multiple POs, let user choose
        await _showPOSelectionDialog(relevantPOs, item);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _showPOSelectionDialog(List<PurchaseOrder> pos, Item item) async {
    final selectedPO = await showDialog<PurchaseOrder>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Select Purchase Order for ${item.name}'),
        content: SizedBox(
          width: double.maxFinite,
          child: ListView.builder(
            shrinkWrap: true,
            itemCount: pos.length,
            itemBuilder: (context, index) {
              final po = pos[index];
              final line = po.lines.firstWhere((line) => line.itemId == item.id);
              return ListTile(
                title: Text('PO ${po.number}'),
                subtitle: Text(
                  '${po.supplierName ?? 'Unknown'} • ${line.qtyRemaining} remaining'
                ),
                onTap: () => Navigator.pop(context, po),
              );
            },
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );

    if (selectedPO != null && mounted) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => GoodsReceiptScreen(purchaseOrder: selectedPO),
        ),
      );
    }
  }

  void _showModeSelectionDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Scan Mode'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            RadioListTile<BarcodeScanMode>(
              title: const Text('Lookup Item'),
              subtitle: const Text('Find and display item information'),
              value: BarcodeScanMode.lookup,
              groupValue: _scanMode,
              onChanged: (value) {
                setState(() => _scanMode = value!);
                Navigator.pop(context);
              },
            ),
            RadioListTile<BarcodeScanMode>(
              title: const Text('Add to Purchase Order'),
              subtitle: const Text('Add scanned item to a new or existing PO'),
              value: BarcodeScanMode.addToOrder,
              groupValue: _scanMode,
              onChanged: (value) {
                setState(() => _scanMode = value!);
                Navigator.pop(context);
              },
            ),
            RadioListTile<BarcodeScanMode>(
              title: const Text('Receive Items'),
              subtitle: const Text('Receive items for goods receipt'),
              value: BarcodeScanMode.receive,
              groupValue: _scanMode,
              onChanged: (value) {
                setState(() => _scanMode = value!);
                Navigator.pop(context);
              },
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );
  }

  String _getModeDescription() {
    switch (_scanMode) {
      case BarcodeScanMode.lookup:
        return 'Find and display item information';
      case BarcodeScanMode.addToOrder:
        return 'Add scanned item to purchase order';
      case BarcodeScanMode.receive:
        return 'Receive items for goods receipt';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const SizedBox(height: 40), // Add top spacing
            const Icon(
              Icons.qr_code_scanner,
              size: 120,
              color: Color(0xFF2563EB),
            ),
            const SizedBox(height: 32),
            Text(
              'Barcode Scanner',
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.bold,
                color: const Color(0xFF2563EB),
              ),
            ),
            const SizedBox(height: 16),

            // Mode selection
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.grey[50],
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.grey[200]!),
              ),
              child: Column(
                children: [
                  Row(
                    children: [
                      const Icon(Icons.settings, color: Color(0xFF2563EB)),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              'Scan Mode',
                              style: TextStyle(fontWeight: FontWeight.bold),
                            ),
                            Text(
                              _getModeDescription(),
                              style: TextStyle(color: Colors.grey[600], fontSize: 12),
                            ),
                          ],
                        ),
                      ),
                      TextButton.icon(
                        onPressed: _showModeSelectionDialog,
                        icon: const Icon(Icons.edit, size: 16),
                        label: const Text('Change'),
                        style: TextButton.styleFrom(
                          foregroundColor: const Color(0xFF2563EB),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),

            const SizedBox(height: 24),
            Text(
              'Tap the button below to scan a barcode.',
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                color: Colors.grey[600],
              ),
            ),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              height: 72,
              child: ElevatedButton.icon(
                onPressed: _isLoading ? null : _scanBarcode,
                icon: _isLoading
                    ? const SizedBox(
                        width: 24,
                        height: 24,
                        child: CircularProgressIndicator(
                          strokeWidth: 3,
                          valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                        ),
                      )
                    : const Icon(Icons.qr_code_scanner, size: 28),
                label: Text(
                  _isLoading ? 'Processing...' : 'Scan Barcode',
                  style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            const SizedBox(height: 32),
            if (_lastScannedCode.isNotEmpty) ...[
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: Column(
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.qr_code, color: Colors.grey),
                          const SizedBox(width: 8),
                          const Text(
                            'Last Scanned:',
                            style: TextStyle(fontWeight: FontWeight.bold),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              _lastScannedCode,
                              style: const TextStyle(fontFamily: 'monospace'),
                            ),
                          ),
                        ],
                      ),
                      if (_scannedItem != null) ...[
                        const Divider(),
                        _ItemResult(item: _scannedItem!),
                      ],
                    ],
                  ),
                ),
              ),
            ],
              const SizedBox(height: 40), // Add bottom spacing
            ],
          ),
        ),
      ),
    );
  }
}

class _ItemResult extends StatelessWidget {
  final Item item;

  const _ItemResult({required this.item});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: const Color(0xFF059669).withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(
                Icons.check_circle,
                color: Color(0xFF059669),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Item Found!',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: Color(0xFF059669),
                    ),
                  ),
                  Text(
                    item.name,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'SKU',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(item.sku),
                ],
              ),
            ),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Price',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    '\$${item.price.toStringAsFixed(2)}',
                    style: const TextStyle(
                      color: Color(0xFF059669),
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        if (item.description != null) ...[
          const SizedBox(height: 12),
          const Text(
            'Description',
            style: TextStyle(
              fontSize: 12,
              color: Colors.grey,
              fontWeight: FontWeight.bold,
            ),
          ),
          Text(item.description!),
        ],
      ],
    );
  }
}