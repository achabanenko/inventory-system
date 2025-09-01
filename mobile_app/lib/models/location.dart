class Location {
  final String id;
  final String code;
  final String name;
  final Map<String, dynamic>? address;
  final bool isActive;

  Location({
    required this.id,
    required this.code,
    required this.name,
    this.address,
    required this.isActive,
  });

  factory Location.fromJson(Map<String, dynamic> json) {
    return Location(
      id: json['id'],
      code: json['code'].toString(),
      name: json['name'],
      address: json['address'] as Map<String, dynamic>?,
      isActive: json['is_active'] ?? true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'code': code,
      'name': name,
      'address': address,
      'is_active': isActive,
    };
  }

  String get displayName => '$code - $name';

  String? get street => address?['street'];
  String? get city => address?['city'];
  String? get state => address?['state'];
  String? get zipCode => address?['zip_code'];
  String? get country => address?['country'];

  String get fullAddress {
    if (address == null) return '';
    final parts = [
      street,
      city,
      state,
      zipCode,
      country,
    ].where((part) => part != null && part.isNotEmpty).toList();
    return parts.join(', ');
  }
}
