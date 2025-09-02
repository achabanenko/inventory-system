import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/purchase_order.dart';
import '../services/api_service.dart';
import '../screens/purchase_order_item_workflow_screen.dart';

class EnhancedPurchaseOrderEditScreen extends StatefulWidget {
  final PurchaseOrder purchaseOrder;
  
  const EnhancedPurchaseOrderEditScreen({
    super.key,
    required this.purchaseOrder,
  });

  @override
  State<EnhancedPurchaseOrderEditScreen> createState() => _EnhancedPurchaseOrderEditScreenState();
}

class _EnhancedPurchaseOrderEditScreenState extends State<EnhancedPurchaseOrderEditScreen> {
  late PurchaseOrder _currentPO;
  final ApiService _apiService = ApiService();
  bool _isLoading = false;
  
  @override
  void initState() {
    super.initState();
    _currentPO = widget.purchaseOrder;
  }
  
  Future<void> _refreshPurchaseOrder() async {
    setState(() {
      _isLoading = true;
    });
    
    try {
      final updatedPO = await _apiService.getPurchaseOrder(widget.purchaseOrder.id);
      setState(() {
        _currentPO = updatedPO;
      });
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error refreshing purchase order: $e'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }
  
  void _showItemWorkflow(WorkflowType workflowType, [PurchaseOrderLine? existingLine]) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => PurchaseOrderItemWorkflowScreen(
        purchaseOrder: _currentPO,
        workflowType: workflowType,
        existingLine: existingLine,
        onItemAdded: (line) {
          setState(() {
            _currentPO = PurchaseOrder(
              id: _currentPO.id,
              number: _currentPO.number,
              supplierId: _currentPO.supplierId,
              supplierName: _currentPO.supplierName,
              status: _currentPO.status,
              expectedAt: _currentPO.expectedAt,
              createdBy: _currentPO.createdBy,
              approvedBy: _currentPO.approvedBy,
              notes: _currentPO.notes,
              createdAt: _currentPO.createdAt,
              updatedAt: _currentPO.updatedAt,
              lines: [..._currentPO.lines, line],
            );
          });
          
          // Refresh from server to get accurate data
          Future.delayed(const Duration(milliseconds: 500), _refreshPurchaseOrder);
        },
        onItemUpdated: (oldLine, newLine) {
          setState(() {
            final lineIndex = _currentPO.lines.indexWhere((l) => l.id == oldLine.id);
            if (lineIndex != -1) {
              final updatedLines = List<PurchaseOrderLine>.from(_currentPO.lines);
              updatedLines[lineIndex] = newLine;
              
              _currentPO = PurchaseOrder(
                id: _currentPO.id,
                number: _currentPO.number,
                supplierId: _currentPO.supplierId,
                supplierName: _currentPO.supplierName,
                status: _currentPO.status,
                expectedAt: _currentPO.expectedAt,
                createdBy: _currentPO.createdBy,
                approvedBy: _currentPO.approvedBy,
                notes: _currentPO.notes,
                createdAt: _currentPO.createdAt,
                updatedAt: _currentPO.updatedAt,
                lines: updatedLines,
              );
            }
          });
          
          // Refresh from server to get accurate data
          Future.delayed(const Duration(milliseconds: 500), _refreshPurchaseOrder);
        },
        onItemDeleted: (deletedLine) {
          setState(() {
            final updatedLines = _currentPO.lines.where((l) => l.id != deletedLine.id).toList();
            
            _currentPO = PurchaseOrder(
              id: _currentPO.id,
              number: _currentPO.number,
              supplierId: _currentPO.supplierId,
              supplierName: _currentPO.supplierName,
              status: _currentPO.status,
              expectedAt: _currentPO.expectedAt,
              createdBy: _currentPO.createdBy,
              approvedBy: _currentPO.approvedBy,
              notes: _currentPO.notes,
              createdAt: _currentPO.createdAt,
              updatedAt: _currentPO.updatedAt,
              lines: updatedLines,
            );
          });
          
          // Refresh from server to get accurate data
          Future.delayed(const Duration(milliseconds: 500), _refreshPurchaseOrder);
        },
      ),
    );
  }
  
  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [Colors.blue.shade600, Colors.blue.shade800],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
      ),
      child: SafeArea(
        bottom: false,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                IconButton(
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(Icons.arrow_back, color: Colors.white),
                ),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Purchase Order',
                        style: TextStyle(
                          color: Colors.white70,
                          fontSize: 14,
                        ),
                      ),
                      Text(
                        _currentPO.number,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: _currentPO.status.color.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: _currentPO.status.color),
                  ),
                  child: Text(
                    _currentPO.status.displayName,
                    style: TextStyle(
                      color: _currentPO.status.color,
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  ),
                ),
              ],
            ),
            
            if (_currentPO.supplierName != null)
              Padding(
                padding: const EdgeInsets.only(top: 8, left: 48),
                child: Text(
                  _currentPO.supplierName!,
                  style: const TextStyle(
                    color: Colors.white70,
                    fontSize: 16,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
  
  Widget _buildStats() {
    final totalItems = _currentPO.totalItems;
    final totalValue = _currentPO.totalValue;
    
    return Container(
      margin: const EdgeInsets.all(16),
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
      child: Row(
        children: [
          Expanded(
            child: Column(
              children: [
                Text(
                  '$totalItems',
                  style: const TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.blue,
                  ),
                ),
                const Text(
                  'Items',
                  style: TextStyle(
                    color: Colors.grey,
                    fontSize: 14,
                  ),
                ),
              ],
            ),
          ),
          Container(
            width: 1,
            height: 40,
            color: Colors.grey.shade300,
          ),
          Expanded(
            child: Column(
              children: [
                Text(
                  '\$${totalValue.toStringAsFixed(2)}',
                  style: const TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.green,
                  ),
                ),
                const Text(
                  'Total Value',
                  style: TextStyle(
                    color: Colors.grey,
                    fontSize: 14,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
  
  Widget _buildActionButtons() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: FilledButton.icon(
              onPressed: () => _showItemWorkflow(WorkflowType.add),
              icon: const Icon(Icons.add),
              label: const Text('Add Item'),
              style: FilledButton.styleFrom(
                backgroundColor: Colors.green,
                padding: const EdgeInsets.all(16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: OutlinedButton.icon(
              onPressed: () => _showItemWorkflow(WorkflowType.edit),
              icon: const Icon(Icons.edit),
              label: const Text('Edit Item'),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.all(16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: OutlinedButton.icon(
              onPressed: () => _showItemWorkflow(WorkflowType.delete),
              icon: const Icon(Icons.delete),
              label: const Text('Delete Item'),
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.red,
                side: const BorderSide(color: Colors.red),
                padding: const EdgeInsets.all(16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
  
  Widget _buildItemsList() {
    if (_currentPO.lines.isEmpty) {
      return Expanded(
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.shopping_cart_outlined,
                size: 80,
                color: Colors.grey.shade400,
              ),
              const SizedBox(height: 16),
              Text(
                'No items in this purchase order',
                style: TextStyle(
                  fontSize: 18,
                  color: Colors.grey.shade600,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Tap "Add Item" to start adding items',
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey.shade500,
                ),
              ),
            ],
          ),
        ),
      );
    }
    
    return Expanded(
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: _currentPO.lines.length,
        itemBuilder: (context, index) {
          final line = _currentPO.lines[index];
          
          return Card(
            margin: const EdgeInsets.only(bottom: 8),
            child: InkWell(
              onTap: () => _showLineActionsMenu(context, line),
              borderRadius: BorderRadius.circular(12),
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Container(
                      width: 50,
                      height: 50,
                      decoration: BoxDecoration(
                        color: Colors.blue.shade50,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Icon(
                        Icons.inventory_2,
                        color: Colors.blue.shade700,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            line.itemName ?? 'Unknown Item',
                            style: const TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 16,
                            ),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'SKU: ${line.itemSku ?? 'N/A'}',
                            style: TextStyle(
                              color: Colors.grey.shade600,
                              fontSize: 14,
                            ),
                          ),
                          if (line.barcode != null)
                            Text(
                              'Barcode: ${line.barcode}',
                              style: TextStyle(
                                color: Colors.grey.shade600,
                                fontSize: 12,
                              ),
                            ),
                        ],
                      ),
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          '${line.qtyOrdered} units',
                          style: const TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 16,
                          ),
                        ),
                        Text(
                          '\$${(line.qtyOrdered * line.unitCost).toStringAsFixed(2)}',
                          style: TextStyle(
                            color: Colors.green.shade700,
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(width: 8),
                    const Icon(
                      Icons.arrow_forward_ios,
                      size: 16,
                      color: Colors.grey,
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }
  
  void _showLineActionsMenu(BuildContext context, PurchaseOrderLine line) {
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.symmetric(vertical: 20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 40,
              height: 4,
              margin: const EdgeInsets.only(bottom: 20),
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            
            // Item info header
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
              child: Row(
                children: [
                  Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: Colors.blue.shade50,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(
                      Icons.inventory_2,
                      color: Colors.blue.shade700,
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          line.itemName ?? 'Unknown Item',
                          style: const TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 16,
                          ),
                        ),
                        Text(
                          'SKU: ${line.itemSku ?? 'N/A'}',
                          style: TextStyle(
                            color: Colors.grey.shade600,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            
            const Divider(),
            
            ListTile(
              leading: const Icon(Icons.edit, color: Colors.blue),
              title: const Text('Edit Item'),
              subtitle: const Text('Update quantity or details'),
              onTap: () {
                Navigator.pop(context);
                _showItemWorkflow(WorkflowType.edit, line);
              },
            ),
            
            ListTile(
              leading: const Icon(Icons.delete, color: Colors.red),
              title: const Text('Remove Item'),
              subtitle: const Text('Delete from purchase order'),
              onTap: () {
                Navigator.pop(context);
                _showItemWorkflow(WorkflowType.delete, line);
              },
            ),
            
            ListTile(
              leading: const Icon(Icons.info, color: Colors.orange),
              title: const Text('View Details'),
              subtitle: const Text('Show item information'),
              onTap: () {
                Navigator.pop(context);
                _showItemDetails(line);
              },
            ),
          ],
        ),
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
            _buildDetailRow('Line Total', '\$${(line.qtyOrdered * line.unitCost).toStringAsFixed(2)}'),
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
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: const TextStyle(fontWeight: FontWeight.w500),
          ),
          Text(
            value,
            style: const TextStyle(color: Colors.grey),
          ),
        ],
      ),
    );
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade50,
      body: Column(
        children: [
          _buildHeader(),
          _buildStats(),
          _buildActionButtons(),
          const SizedBox(height: 8),
          _buildItemsList(),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showItemWorkflow(WorkflowType.add),
        backgroundColor: Colors.green,
        child: const Icon(Icons.add),
      ),
    );
  }
}