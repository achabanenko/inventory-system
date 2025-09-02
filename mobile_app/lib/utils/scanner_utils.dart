import 'package:flutter/material.dart';
import '../models/item.dart';
import '../services/api_service.dart';
import '../screens/scanned_item_screen.dart';
import '../widgets/barcode_scanner_overlay.dart';

class ScannerUtils {
  static final ApiService _apiService = ApiService();
  
  /// Show barcode scanner and handle scanned item details
  static Future<void> showScannerWithItemDetail({
    required BuildContext context,
    required String documentType,
    required Function(Item, double) onAddToDocument,
    Map<String, dynamic>? contextData,
    String? headerText,
    Widget? contextWidget,
  }) async {
    
    // Show scanner overlay
    final scannedBarcode = await showModalBottomSheet<String>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => BarcodeScannerOverlay(
        mode: ScannerMode.showDetail,
        headerText: headerText ?? 'Scan Item for $documentType',
        contextWidget: contextWidget,
        documentType: documentType,
        contextData: contextData,
        onBarcodeScanned: (barcode) {
          Navigator.pop(context, barcode);
        },
        onCancel: () {
          Navigator.pop(context);
        },
      ),
    );
    
    if (scannedBarcode == null) return;
    
    // Show loading while looking up item
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(
        child: CircularProgressIndicator(),
      ),
    );
    
    try {
      // Look up item by barcode
      final items = await _apiService.searchItems(scannedBarcode, 'barcode');
      
      // Dismiss loading
      Navigator.pop(context);
      
      if (items.isEmpty) {
        // Show item not found dialog
        _showItemNotFoundDialog(context, scannedBarcode, documentType, onAddToDocument, contextData);
        return;
      }
      
      final scannedItem = items.first;
      
      // Show scanned item detail screen
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => ScannedItemScreen(
            scannedItem: scannedItem,
            documentType: documentType,
            contextData: contextData,
            onAddToDocument: onAddToDocument,
            onScanAnother: () {
              // Navigate back and scan another item
              Navigator.pop(context);
              showScannerWithItemDetail(
                context: context,
                documentType: documentType,
                onAddToDocument: onAddToDocument,
                contextData: contextData,
                headerText: headerText,
                contextWidget: contextWidget,
              );
            },
          ),
        ),
      );
      
    } catch (e) {
      // Dismiss loading
      Navigator.pop(context);
      
      // Show error
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error looking up item: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }
  
  static void _showItemNotFoundDialog(
    BuildContext context,
    String barcode,
    String documentType,
    Function(Item, double) onAddToDocument,
    Map<String, dynamic>? contextData,
  ) {
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
            Text('No item found with barcode:'),
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
            onPressed: () {
              Navigator.pop(context);
            },
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              // Scan another item
              showScannerWithItemDetail(
                context: context,
                documentType: documentType,
                onAddToDocument: onAddToDocument,
                contextData: contextData,
              );
            },
            child: const Text('Scan Another'),
          ),
          FilledButton(
            onPressed: () {
              Navigator.pop(context);
              _showCreateItemDialog(context, barcode, documentType, onAddToDocument, contextData);
            },
            child: const Text('Create Item'),
          ),
        ],
      ),
    );
  }
  
  static void _showCreateItemDialog(
    BuildContext context,
    String barcode,
    String documentType,
    Function(Item, double) onAddToDocument,
    Map<String, dynamic>? contextData,
  ) {
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
                keyboardType: TextInputType.numberWithOptions(decimal: true),
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
              
              // Create temporary item (in real app, would create via API)
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
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => ScannedItemScreen(
                    scannedItem: newItem,
                    documentType: documentType,
                    contextData: contextData,
                    onAddToDocument: onAddToDocument,
                  ),
                ),
              );
            },
            child: const Text('Create & Add'),
          ),
        ],
      ),
    );
  }
}