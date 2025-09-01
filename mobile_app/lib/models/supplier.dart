class Supplier {
  final String id;
  final String code;
  final String name;
  final Map<String, dynamic>? contact;
  final bool isActive;

  Supplier({
    required this.id,
    required this.code,
    required this.name,
    this.contact,
    required this.isActive,
  });

  factory Supplier.fromJson(Map<String, dynamic> json) {
    return Supplier(
      id: json['id'],
      code: json['code'].toString(),
      name: json['name'],
      contact: json['contact'] as Map<String, dynamic>?,
      isActive: json['is_active'] ?? true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'code': code,
      'name': name,
      'contact': contact,
      'is_active': isActive,
    };
  }

  String get displayName => '$code - $name';

  String? get email => contact?['email'];
  String? get phone => contact?['phone'];
  String? get address => contact?['address'];
}
