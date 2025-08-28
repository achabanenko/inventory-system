class User {
  final String id;
  final String email;
  final String name;
  final String role;
  final String? tenantId;

  User({
    required this.id,
    required this.email,
    required this.name,
    required this.role,
    this.tenantId,
  });

  String get fullName => name;
  String get firstName => name.split(' ').first;
  String get lastName => name.split(' ').length > 1 ? name.split(' ').last : '';

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'],
      email: json['email'],
      name: json['name'],
      role: json['role'],
      tenantId: json['tenant_id'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'email': email,
      'name': name,
      'role': role,
      'tenant_id': tenantId,
    };
  }
}