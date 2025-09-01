import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/purchase_order.dart';
import '../services/api_service.dart';
import 'purchase_order_create_screen.dart';
import 'purchase_order_edit_screen.dart';
import 'goods_receipt_screen.dart';

class PurchaseOrdersScreen extends StatefulWidget {
  const PurchaseOrdersScreen({super.key});

  @override
  State<PurchaseOrdersScreen> createState() => _PurchaseOrdersScreenState();
}

class _PurchaseOrdersScreenState extends State<PurchaseOrdersScreen> {
  final TextEditingController _searchController = TextEditingController();
  List<PurchaseOrder> _purchaseOrders = [];
  bool _isLoading = false;
  String? _selectedStatus;
  String _searchQuery = '';

  final List<String> _statusOptions = [
    'All',
    'DRAFT',
    'APPROVED',
    'PARTIAL',
    'RECEIVED',
    'CLOSED',
    'CANCELED',
  ];

  @override
  void initState() {
    super.initState();
    _loadPurchaseOrders();
    _searchController.addListener(() {
      setState(() {});
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadPurchaseOrders() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final apiService = Provider.of<ApiService>(context, listen: false);
      final status = _selectedStatus == 'All' || _selectedStatus == null ? null : _selectedStatus;

      final purchaseOrders = await apiService.getPurchaseOrders(
        status: status,
        page: 1,
        limit: 50,
      );

      setState(() {
        _purchaseOrders = purchaseOrders;
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error loading purchase orders: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  void _onSearchPressed() {
    setState(() {
      _searchQuery = _searchController.text.trim();
    });
    _loadPurchaseOrders();
  }

  void _onClearSearch() {
    _searchController.clear();
    setState(() {
      _searchQuery = '';
    });
    _loadPurchaseOrders();
  }

  List<PurchaseOrder> get _filteredPurchaseOrders {
    if (_searchQuery.isEmpty) return _purchaseOrders;

    return _purchaseOrders.where((po) {
      final query = _searchQuery.toLowerCase();
      return po.number.toLowerCase().contains(query) ||
             (po.supplierName?.toLowerCase().contains(query) ?? false) ||
             po.lines.any((line) =>
               (line.itemName?.toLowerCase().contains(query) ?? false) ||
               (line.itemSku?.toLowerCase().contains(query) ?? false)
             );
    }).toList();
  }

  void _showStatusFilter() {
    showModalBottomSheet(
      context: context,
      builder: (context) => Container(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Filter by Status',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: _statusOptions.map((status) {
                final isSelected = _selectedStatus == status || (_selectedStatus == null && status == 'All');
                return FilterChip(
                  label: Text(status == 'All' ? 'All Statuses' : PurchaseOrderStatus.fromString(status).displayName),
                  selected: isSelected,
                  onSelected: (selected) {
                    setState(() {
                      _selectedStatus = selected ? (status == 'All' ? null : status) : null;
                    });
                    Navigator.pop(context);
                    _loadPurchaseOrders();
                  },
                );
              }).toList(),
            ),
          ],
        ),
      ),
    );
  }

  void _navigateToCreatePO() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const PurchaseOrderCreateScreen(),
      ),
    ).then((_) => _loadPurchaseOrders()); // Refresh list when returning
  }

  void _navigateToPODetail(PurchaseOrder po) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => PurchaseOrderEditScreen(purchaseOrder: po),
      ),
    ).then((result) {
      // Refresh list when returning if changes were made
      if (result == true) {
        _loadPurchaseOrders();
      }
    });
  }

  void _navigateToGoodsReceipt(PurchaseOrder po) {
    if (!po.canReceive) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('This purchase order cannot receive items at this time'),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => GoodsReceiptScreen(purchaseOrder: po),
      ),
    ).then((_) => _loadPurchaseOrders()); // Refresh list when returning
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          // Search and Filter Bar
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
                      child: TextField(
                        controller: _searchController,
                        onSubmitted: (_) => _onSearchPressed(),
                        decoration: InputDecoration(
                          hintText: 'Search POs, suppliers, items...',
                          prefixIcon: const Icon(Icons.search),
                          border: const OutlineInputBorder(),
                          suffixIcon: _searchController.text.isNotEmpty
                              ? IconButton(
                                  icon: const Icon(Icons.clear),
                                  onPressed: _onClearSearch,
                                )
                              : null,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    ElevatedButton.icon(
                      onPressed: _onSearchPressed,
                      icon: const Icon(Icons.search, size: 18),
                      label: const Text('Search'),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        'Status: ${_selectedStatus == null ? 'All' : PurchaseOrderStatus.fromString(_selectedStatus!).displayName}',
                        style: const TextStyle(fontWeight: FontWeight.w500),
                      ),
                    ),
                    TextButton.icon(
                      onPressed: _showStatusFilter,
                      icon: const Icon(Icons.filter_list, size: 18),
                      label: const Text('Filter'),
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Purchase Orders List
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _filteredPurchaseOrders.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.inventory_2,
                              size: 64,
                              color: Colors.grey[400],
                            ),
                            const SizedBox(height: 16),
                            Text(
                              _purchaseOrders.isEmpty ? 'No purchase orders found' : 'No matching purchase orders',
                              style: TextStyle(
                                fontSize: 18,
                                color: Colors.grey[600],
                              ),
                            ),
                          ],
                        ),
                      )
                    : RefreshIndicator(
                        onRefresh: _loadPurchaseOrders,
                        child: ListView.builder(
                          padding: const EdgeInsets.all(16),
                          itemCount: _filteredPurchaseOrders.length,
                          itemBuilder: (context, index) {
                            final po = _filteredPurchaseOrders[index];
                            return _PurchaseOrderCard(
                              purchaseOrder: po,
                              onTap: () => _navigateToPODetail(po),
                              onReceive: po.canReceive ? () => _navigateToGoodsReceipt(po) : null,
                            );
                          },
                        ),
                      ),
          ),
        ],
      ),

      floatingActionButton: FloatingActionButton.extended(
        onPressed: _navigateToCreatePO,
        icon: const Icon(Icons.add),
        label: const Text('New PO'),
      ),
    );
  }
}

class _PurchaseOrderCard extends StatelessWidget {
  final PurchaseOrder purchaseOrder;
  final VoidCallback onTap;
  final VoidCallback? onReceive;

  const _PurchaseOrderCard({
    required this.purchaseOrder,
    required this.onTap,
    this.onReceive,
  });

  @override
  Widget build(BuildContext context) {
    final progressValue = purchaseOrder.totalItems > 0
        ? purchaseOrder.totalReceived / purchaseOrder.totalItems
        : 0.0;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header with PO Number and Status
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'PO ${purchaseOrder.number}',
                          style: const TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Text(
                          purchaseOrder.supplierName ?? 'Unknown Supplier',
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
                      color: purchaseOrder.status.color.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      purchaseOrder.status.displayName,
                      style: TextStyle(
                        color: purchaseOrder.status.color,
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 12),

              // Progress Bar for Received Items
              if (purchaseOrder.lines.isNotEmpty) ...[
                Row(
                  children: [
                    const Icon(Icons.inventory_2, size: 16, color: Colors.grey),
                    const SizedBox(width: 4),
                    Text(
                      '${purchaseOrder.totalReceived}/${purchaseOrder.totalItems} items received',
                      style: const TextStyle(fontSize: 12),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                LinearProgressIndicator(
                  value: progressValue,
                  backgroundColor: Colors.grey[200],
                  valueColor: AlwaysStoppedAnimation<Color>(
                    purchaseOrder.status == PurchaseOrderStatus.received
                        ? Colors.green
                        : Colors.blue,
                  ),
                ),
              ],

              const SizedBox(height: 12),

              // Footer with dates and value
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Created: ${_formatDate(purchaseOrder.createdAt)}',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[600],
                          ),
                        ),
                        if (purchaseOrder.expectedAt != null)
                          Text(
                            'Expected: ${_formatDate(purchaseOrder.expectedAt!)}',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey[600],
                            ),
                          ),
                      ],
                    ),
                  ),
                  Text(
                    '\$${purchaseOrder.totalValue.toStringAsFixed(2)}',
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Color(0xFF059669),
                    ),
                  ),
                ],
              ),

              // Action buttons
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: onTap,
                      icon: const Icon(Icons.edit, size: 16),
                      label: Text(purchaseOrder.status == PurchaseOrderStatus.draft || purchaseOrder.status == PurchaseOrderStatus.approved ? 'Edit PO' : 'View PO'),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: const Color(0xFF2563EB),
                        side: const BorderSide(color: Color(0xFF2563EB)),
                      ),
                    ),
                  ),
                  if (onReceive != null) ...[
                    const SizedBox(width: 8),
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: onReceive,
                        icon: const Icon(Icons.inventory, size: 16),
                        label: const Text('Receive'),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: const Color(0xFF059669),
                          side: const BorderSide(color: Color(0xFF059669)),
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final difference = now.difference(date).inDays;

    if (difference == 0) {
      return 'Today';
    } else if (difference == 1) {
      return 'Yesterday';
    } else if (difference < 7) {
      return '$difference days ago';
    } else {
      return '${date.month}/${date.day}/${date.year}';
    }
  }
}
