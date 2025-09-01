import 'package:flutter/material.dart';

enum PurchaseOrderStatus {
  draft('DRAFT'),
  approved('APPROVED'),
  partial('PARTIAL'),
  received('RECEIVED'),
  closed('CLOSED'),
  canceled('CANCELED');

  const PurchaseOrderStatus(this.value);
  final String value;

  static PurchaseOrderStatus fromString(String value) {
    return PurchaseOrderStatus.values.firstWhere(
      (status) => status.value == value,
      orElse: () => PurchaseOrderStatus.draft,
    );
  }

  String get displayName {
    switch (this) {
      case PurchaseOrderStatus.draft:
        return 'Draft';
      case PurchaseOrderStatus.approved:
        return 'Approved';
      case PurchaseOrderStatus.partial:
        return 'Partially Received';
      case PurchaseOrderStatus.received:
        return 'Received';
      case PurchaseOrderStatus.closed:
        return 'Closed';
      case PurchaseOrderStatus.canceled:
        return 'Canceled';
    }
  }

  Color get color {
    switch (this) {
      case PurchaseOrderStatus.draft:
        return Colors.grey;
      case PurchaseOrderStatus.approved:
        return Colors.blue;
      case PurchaseOrderStatus.partial:
        return Colors.orange;
      case PurchaseOrderStatus.received:
        return Colors.green;
      case PurchaseOrderStatus.closed:
        return Colors.purple;
      case PurchaseOrderStatus.canceled:
        return Colors.red;
    }
  }
}

class PurchaseOrder {
  final String id;
  final String number;
  final String supplierId;
  final String? supplierName;
  final PurchaseOrderStatus status;
  final DateTime? expectedAt;
  final String? createdBy;
  final String? approvedBy;
  final String? notes;
  final DateTime createdAt;
  final DateTime updatedAt;
  final List<PurchaseOrderLine> lines;

  PurchaseOrder({
    required this.id,
    required this.number,
    required this.supplierId,
    this.supplierName,
    required this.status,
    this.expectedAt,
    this.createdBy,
    this.approvedBy,
    this.notes,
    required this.createdAt,
    required this.updatedAt,
    required this.lines,
  });

  factory PurchaseOrder.fromJson(Map<String, dynamic> json) {
    return PurchaseOrder(
      id: json['id'],
      number: json['number'].toString(),
      supplierId: json['supplier_id'],
      supplierName: json['supplier_name'] ?? json['supplier']?['name'],
      status: PurchaseOrderStatus.fromString(json['status']),
      expectedAt: json['expected_at'] != null
          ? DateTime.parse(json['expected_at'])
          : null,
      createdBy: json['created_by'],
      approvedBy: json['approved_by'],
      notes: json['notes'],
      createdAt: DateTime.parse(json['created_at']),
      updatedAt: DateTime.parse(json['updated_at']),
      lines: (json['lines'] as List<dynamic>?)
          ?.map((line) => PurchaseOrderLine.fromJson(line))
          .toList() ?? [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'number': number,
      'supplier_id': supplierId,
      'supplier_name': supplierName,
      'status': status.value,
      'expected_at': expectedAt?.toIso8601String(),
      'created_by': createdBy,
      'approved_by': approvedBy,
      'notes': notes,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
      'lines': lines.map((line) => line.toJson()).toList(),
    };
  }

  // Helper methods
  int get totalItems => lines.fold(0, (sum, line) => sum + line.qtyOrdered);
  int get totalReceived => lines.fold(0, (sum, line) => sum + line.qtyReceived);
  double get totalValue => lines.fold(0, (sum, line) => sum + (line.qtyOrdered * line.unitCost));

  bool get isEditable => status == PurchaseOrderStatus.draft;
  bool get canReceive => status == PurchaseOrderStatus.approved || status == PurchaseOrderStatus.partial;
  bool get isCompleted => status == PurchaseOrderStatus.received || status == PurchaseOrderStatus.closed;
}

class PurchaseOrderLine {
  final String id;
  final String poId;
  final String itemId;
  final String? itemName;
  final String? itemSku;
  final String? barcode;
  final int qtyOrdered;
  final int qtyReceived;
  final double unitCost;
  final Map<String, dynamic>? tax;

  PurchaseOrderLine({
    required this.id,
    required this.poId,
    required this.itemId,
    this.itemName,
    this.itemSku,
    this.barcode,
    required this.qtyOrdered,
    required this.qtyReceived,
    required this.unitCost,
    this.tax,
  });

  factory PurchaseOrderLine.fromJson(Map<String, dynamic> json) {
    return PurchaseOrderLine(
      id: json['id'],
      poId: json['po_id'] ?? json['purchase_order_id'] ?? '',
      itemId: json['item_id'],
      itemName: json['item_name'] ?? json['item']?['name'],
      itemSku: json['item_sku'] ?? json['item']?['sku'],
      barcode: json['barcode']?.toString() ?? json['item']?['barcode']?.toString(),
      qtyOrdered: json['qty_ordered'] ?? 0,
      qtyReceived: json['qty_received'] ?? 0,
      unitCost: _parseDouble(json['unit_cost']),
      tax: _parseTax(json['tax']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'po_id': poId,
      'item_id': itemId,
      'item_name': itemName,
      'item_sku': itemSku,
      'barcode': barcode,
      'qty_ordered': qtyOrdered,
      'qty_received': qtyReceived,
      'unit_cost': unitCost,
      'tax': tax,
    };
  }

  static double _parseDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  static Map<String, dynamic>? _parseTax(dynamic value) {
    if (value == null) return null;
    if (value is Map<String, dynamic>) return value;
    if (value is String) {
      try {
        // Handle base64 encoded JSON or plain JSON string
        if (value.isEmpty || value == 'e30=') {
          return {}; // Empty object for empty base64 ({})
        }
        // Try to parse as JSON string
        return {};
      } catch (e) {
        return {};
      }
    }
    return {};
  }

  // Helper methods
  int get qtyRemaining => qtyOrdered - qtyReceived;
  double get lineTotal => qtyOrdered * unitCost;
  double get receivedTotal => qtyReceived * unitCost;
  bool get isFullyReceived => qtyReceived >= qtyOrdered;
  bool get hasPartialReceipt => qtyReceived > 0 && qtyReceived < qtyOrdered;
}
