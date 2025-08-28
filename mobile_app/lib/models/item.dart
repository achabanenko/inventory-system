class Item {
  final String id;
  final String sku;
  final String name;
  final String? description;
  final String? barcode;
  final double price;
  final double costPrice;
  final String? uom;
  final bool isActive;

  Item({
    required this.id,
    required this.sku,
    required this.name,
    this.description,
    this.barcode,
    required this.price,
    required this.costPrice,
    this.uom,
    required this.isActive,
  });

  factory Item.fromJson(Map<String, dynamic> json) {
    return Item(
      id: json['id'],
      sku: json['sku'].toString(),
      name: json['name'],
      description: json['description'],
      barcode: json['barcode']?.toString(),
      price: _parseDouble(json['price']),
      costPrice: _parseDouble(json['cost']),
      uom: json['uom'],
      isActive: json['is_active'] ?? true,
    );
  }

  static double _parseDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'sku': sku,
      'name': name,
      'description': description,
      'barcode': barcode,
      'price': price,
      'cost': costPrice,
      'uom': uom,
      'is_active': isActive,
    };
  }
}