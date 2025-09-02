import 'package:flutter/material.dart';
import '../models/item.dart';
import '../widgets/scanned_item_detail_card.dart';

class ScannedItemScreen extends StatefulWidget {
  final Item scannedItem;
  final String documentType; // 'PO', 'Receipt', 'Transfer', 'Count'
  final Map<String, dynamic>? contextData;
  final Function(Item, double) onAddToDocument;
  final VoidCallback? onScanAnother;
  
  const ScannedItemScreen({
    super.key,
    required this.scannedItem,
    required this.documentType,
    required this.onAddToDocument,
    this.contextData,
    this.onScanAnother,
  });

  @override
  State<ScannedItemScreen> createState() => _ScannedItemScreenState();
}

class _ScannedItemScreenState extends State<ScannedItemScreen> {
  double _quantity = 1; // Default quantity is always 1 for newly scanned items
  
  void _onQuantityChanged(double quantity) {
    setState(() {
      _quantity = quantity;
    });
  }
  
  void _addToDocument() {
    // Call the onAddToDocument callback which handles API call and navigation
    // The callback will handle showing success/error messages and closing the screen
    widget.onAddToDocument(widget.scannedItem, _quantity);
  }
  
  void _scanAnother() {
    if (widget.onScanAnother != null) {
      widget.onScanAnother!();
    } else {
      Navigator.pop(context);
    }
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade100,
      appBar: AppBar(
        title: Text('Scanned Item - ${widget.documentType}'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SingleChildScrollView(
        child: Column(
          children: [
            const SizedBox(height: 20),
            ScannedItemDetailCard(
              item: widget.scannedItem,
              initialQuantity: _quantity,
              onQuantityChanged: _onQuantityChanged,
              onAddToDocument: _addToDocument,
              onScanAnother: _scanAnother,
              documentType: widget.documentType,
              contextData: widget.contextData,
            ),
          ],
        ),
      ),
    );
  }
}