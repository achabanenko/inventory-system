import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/user.dart';
import '../models/item.dart';
import '../models/supplier.dart';
import '../models/location.dart';
import '../models/purchase_order.dart';
import 'cache_service.dart';

class ApiService {
  static const String baseUrl = 'http://192.168.8.135:8080/api/v1';
  // static const String baseUrl = 'http://192.168.0.44:8080/api/v1';
  // static const String baseUrl = 'http://localhost:8080/api/v1';
  String? _token;
  Function? _onTokenExpired;

  void setToken(String token) {
    _token = token.isEmpty ? null : token;
    print('ApiService: Token set to: ${_token != null ? "***${_token!.substring(_token!.length - 10)}" : "null"}');
  }

  void setOnTokenExpired(Function callback) {
    _onTokenExpired = callback;
  }

  Map<String, String> get _headers {
    final headers = {
      'Content-Type': 'application/json',
    };
    if (_token != null) {
      headers['Authorization'] = 'Bearer $_token';
      print('ApiService: Using auth header: Bearer ***${_token!.substring(_token!.length - 10)}');
    } else {
      print('ApiService: No token available for request');
    }
    return headers;
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    print('Attempting login to: $baseUrl/auth/login');
    print('Email: $email');
    
    final response = await http.post(
      Uri.parse('$baseUrl/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'email': email,
        'password': password,
      }),
    );

    print('Response status: ${response.statusCode}');
    print('Response body: ${response.body}');

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      throw Exception('Login failed: ${response.body}');
    }
  }

  Future<void> logout() async {
    await http.post(
      Uri.parse('$baseUrl/auth/logout'),
      headers: _headers,
    );
  }

  Future<Map<String, dynamic>> refreshToken(String refreshToken) async {
    final response = await http.post(
      Uri.parse('$baseUrl/auth/refresh'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'refresh_token': refreshToken,
      }),
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      throw Exception('Token refresh failed');
    }
  }

  Future<Map<String, dynamic>> googleOAuth(String code, String redirectUri) async {
    print('Attempting Google OAuth to: $baseUrl/auth/google');
    print('Code: $code');
    print('Redirect URI: $redirectUri');
    
    final response = await http.post(
      Uri.parse('$baseUrl/auth/google'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'code': code,
        'redirect_uri': redirectUri,
      }),
    );

    print('OAuth Response status: ${response.statusCode}');
    print('OAuth Response body: ${response.body}');

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      throw Exception('Google OAuth failed: ${response.body}');
    }
  }

  Future<List<Item>> getItems({String? search, int page = 1, int limit = 20, bool useCache = true}) async {
    // Use cached version if enabled
    if (useCache) {
      final cachedItems = await CacheService.instance.getCachedItems(search: search);
      if (cachedItems.isNotEmpty) {
        return cachedItems;
      }
    }
    return _fetchItemsFromAPI(search: search, page: page, limit: limit);
  }

  Future<List<Item>> _fetchItemsFromAPI({String? search, int page = 1, int limit = 20}) async {
    final queryParams = <String, String>{
      'page': page.toString(),
      'limit': limit.toString(),
    };
    if (search != null && search.isNotEmpty) {
      queryParams['q'] = search;
    }

    final uri = Uri.parse('$baseUrl/items').replace(queryParameters: queryParams);
    print('ApiService: Requesting URL: $uri');
    print('ApiService: Query params: $queryParams');
    final response = await http.get(uri, headers: _headers);

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      print('Items response: $data');
      
      final itemsList = data['data'] as List?;
      if (itemsList == null) {
        print('Items list is null, returning empty list');
        return <Item>[];
      }
      
      final items = itemsList
          .map((item) => Item.fromJson(item))
          .toList();
      
      // Cache the items if caching is enabled
      for (final item in items) {
        await CacheService.instance.cacheItem(item);
      }
      
      return items;
    } else if (response.statusCode == 401) {
      print('ApiService: Token expired, attempting refresh');
      if (_onTokenExpired != null) {
        final refreshed = await _onTokenExpired!();
        if (refreshed) {
          // Retry the request with new token
          return getItems(search: search, page: page, limit: limit);
        }
      }
      throw Exception('Authentication failed - please login again');
    } else {
      throw Exception('Failed to fetch items: ${response.statusCode} - ${response.body}');
    }
  }

  Future<Item?> _fetchItemFromAPI(String itemId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/items/$itemId'),
      headers: _headers,
    );

    if (response.statusCode == 200) {
      return Item.fromJson(jsonDecode(response.body));
    } else if (response.statusCode == 404) {
      return null;
    } else {
      throw Exception('Failed to fetch item');
    }
  }

  Future<Item?> getItemByBarcode(String barcode, {bool useCache = true}) async {
    final response = await http.get(
      Uri.parse('$baseUrl/items/barcode/$barcode'),
      headers: _headers,
    );

    if (response.statusCode == 200) {
      final item = Item.fromJson(jsonDecode(response.body));
      // Cache the item if caching is enabled
      if (useCache) {
        await CacheService.instance.cacheItem(item);
      }
      return item;
    } else if (response.statusCode == 404) {
      return null;
    } else {
      throw Exception('Failed to fetch item by barcode');
    }
  }

  Future<Item> createItem(Item item) async {
    final response = await http.post(
      Uri.parse('$baseUrl/items'),
      headers: _headers,
      body: jsonEncode(item.toJson()),
    );

    if (response.statusCode == 201) {
      return Item.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to create item');
    }
  }

  Future<Item> updateItem(Item item) async {
    final response = await http.put(
      Uri.parse('$baseUrl/items/${item.id}'),
      headers: _headers,
      body: jsonEncode(item.toJson()),
    );

    if (response.statusCode == 200) {
      return Item.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to update item');
    }
  }

  Future<void> deleteItem(int id) async {
    final response = await http.delete(
      Uri.parse('$baseUrl/items/$id'),
      headers: _headers,
    );

    if (response.statusCode != 204) {
      throw Exception('Failed to delete item');
    }
  }

  Future<String> chatWithAI(String message) async {
    final response = await http.post(
      Uri.parse('$baseUrl/chat'),
      headers: _headers,
      body: jsonEncode({
        'message': message,
      }),
    );

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return data['response'];
    } else {
      throw Exception('AI chat failed');
    }
  }

  Future<Map<String, dynamic>> getInventoryLevel(int itemId) async {
    final response = await http.get(
      Uri.parse('$baseUrl/inventory/$itemId/locations'),
      headers: _headers,
    );

    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      throw Exception('Failed to fetch inventory level');
    }
  }

  // ============================================================================
  // DEBUGGING HELPERS
  // ============================================================================

  /// Test method to verify authentication is working
  Future<bool> testAuthentication() async {
    try {
      print('ApiService: Testing authentication...');
      print('ApiService: Current token: ${_token != null ? "***${_token!.substring(_token!.length - 10)}" : "null"}');

      // Try to get current user info
      final response = await http.get(
        Uri.parse('$baseUrl/auth/me'),
        headers: _headers,
      );

      print('ApiService: Auth test response status: ${response.statusCode}');
      print('ApiService: Auth test response body: ${response.body}');

      return response.statusCode == 200;
    } catch (e) {
      print('ApiService: Auth test error: $e');
      return false;
    }
  }

  // ============================================================================
  // SUPPLIERS API METHODS
  // ============================================================================

  Future<List<Supplier>> getSuppliers({String? search, int page = 1, int limit = 20}) async {
    final queryParams = <String, String>{
      'page': page.toString(),
      'limit': limit.toString(),
    };
    if (search != null && search.isNotEmpty) {
      queryParams['q'] = search;
    }

    final uri = Uri.parse('$baseUrl/suppliers').replace(queryParameters: queryParams);
    print('ApiService: Making request to suppliers: ${uri.toString()}');
    final response = await http.get(uri, headers: _headers);
    print('ApiService: Suppliers response status: ${response.statusCode}');

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final suppliersList = data['data'] as List?;
      if (suppliersList == null) {
        return <Supplier>[];
      }

      return suppliersList
          .map((supplier) => Supplier.fromJson(supplier))
          .toList();
    } else if (response.statusCode == 401) {
      if (_onTokenExpired != null) {
        final refreshed = await _onTokenExpired!();
        if (refreshed) {
          return getSuppliers(search: search, page: page, limit: limit);
        }
      }
      throw Exception('Authentication failed - please login again');
    } else {
      throw Exception('Failed to fetch suppliers: ${response.statusCode} - ${response.body}');
    }
  }

  Future<Supplier> getSupplier(String id) async {
    final response = await http.get(
      Uri.parse('$baseUrl/suppliers/$id'),
      headers: _headers,
    );

    if (response.statusCode == 200) {
      return Supplier.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to fetch supplier');
    }
  }

  // ============================================================================
  // LOCATIONS API METHODS
  // ============================================================================

  Future<List<Location>> getLocations({String? search, int page = 1, int limit = 20}) async {
    final queryParams = <String, String>{
      'page': page.toString(),
      'limit': limit.toString(),
    };
    if (search != null && search.isNotEmpty) {
      queryParams['q'] = search;
    }

    final uri = Uri.parse('$baseUrl/locations').replace(queryParameters: queryParams);
    final response = await http.get(uri, headers: _headers);

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final locationsList = data['data'] as List?;
      if (locationsList == null) {
        return <Location>[];
      }

      return locationsList
          .map((location) => Location.fromJson(location))
          .toList();
    } else if (response.statusCode == 401) {
      if (_onTokenExpired != null) {
        final refreshed = await _onTokenExpired!();
        if (refreshed) {
          return getLocations(search: search, page: page, limit: limit);
        }
      }
      throw Exception('Authentication failed - please login again');
    } else {
      throw Exception('Failed to fetch locations: ${response.statusCode} - ${response.body}');
    }
  }

  Future<Location> getLocation(String id) async {
    final response = await http.get(
      Uri.parse('$baseUrl/locations/$id'),
      headers: _headers,
    );

    if (response.statusCode == 200) {
      return Location.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to fetch location');
    }
  }

  // ============================================================================
  // PURCHASE ORDERS API METHODS
  // ============================================================================

  Future<List<PurchaseOrder>> getPurchaseOrders({
    String? status,
    String? supplierId,
    int page = 1,
    int limit = 20
  }) async {
    final queryParams = <String, String>{
      'page': page.toString(),
      'limit': limit.toString(),
    };
    if (status != null && status.isNotEmpty) {
      queryParams['status'] = status;
    }
    if (supplierId != null && supplierId.isNotEmpty) {
      queryParams['supplier_id'] = supplierId;
    }

    final uri = Uri.parse('$baseUrl/purchase-orders').replace(queryParameters: queryParams);
    print('ApiService: Making request to purchase orders: ${uri.toString()}');
    final response = await http.get(uri, headers: _headers);
    print('ApiService: Purchase orders response status: ${response.statusCode}');

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      final poList = data['data'] as List?;
      if (poList == null) {
        return <PurchaseOrder>[];
      }

      return poList
          .map((po) => PurchaseOrder.fromJson(po))
          .toList();
    } else if (response.statusCode == 401) {
      if (_onTokenExpired != null) {
        final refreshed = await _onTokenExpired!();
        if (refreshed) {
          return getPurchaseOrders(
            status: status,
            supplierId: supplierId,
            page: page,
            limit: limit,
          );
        }
      }
      throw Exception('Authentication failed - please login again');
    } else {
      throw Exception('Failed to fetch purchase orders: ${response.statusCode} - ${response.body}');
    }
  }

  Future<PurchaseOrder> getPurchaseOrder(String id) async {
    final response = await http.get(
      Uri.parse('$baseUrl/purchase-orders/$id'),
      headers: _headers,
    );

    if (response.statusCode == 200) {
      return PurchaseOrder.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to fetch purchase order');
    }
  }

  Future<PurchaseOrder> createPurchaseOrder({
    required String supplierId,
    DateTime? expectedAt,
    String? notes,
    List<Map<String, dynamic>>? lines,
  }) async {
    final requestBody = {
      'supplier_id': supplierId,
      'expected_at': expectedAt?.toIso8601String(),
      'notes': notes,
      'lines': lines ?? [],
    };

    final response = await http.post(
      Uri.parse('$baseUrl/purchase-orders'),
      headers: _headers,
      body: jsonEncode(requestBody),
    );

    if (response.statusCode == 201) {
      return PurchaseOrder.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to create purchase order: ${response.body}');
    }
  }

  Future<PurchaseOrder> updatePurchaseOrder(
    String id, {
    String? supplierId,
    DateTime? expectedAt,
    String? notes,
    List<Map<String, dynamic>>? lines,
  }) async {
    final requestBody = <String, dynamic>{};
    if (supplierId != null) {
      requestBody['supplier_id'] = supplierId;
    }
    if (expectedAt != null) {
      requestBody['expected_at'] = expectedAt.toIso8601String();
    }
    if (notes != null) {
      requestBody['notes'] = notes;
    }
    if (lines != null) {
      requestBody['lines'] = lines;
    }

    final response = await http.put(
      Uri.parse('$baseUrl/purchase-orders/$id'),
      headers: _headers,
      body: jsonEncode(requestBody),
    );

    if (response.statusCode == 200) {
      return PurchaseOrder.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to update purchase order: ${response.body}');
    }
  }

  Future<void> approvePurchaseOrder(String id) async {
    final response = await http.post(
      Uri.parse('$baseUrl/purchase-orders/$id/approve'),
      headers: _headers,
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to approve purchase order: ${response.body}');
    }
  }

  Future<void> receivePurchaseOrder(String id, List<Map<String, dynamic>> receipts) async {
    final response = await http.post(
      Uri.parse('$baseUrl/purchase-orders/$id/receive'),
      headers: _headers,
      body: jsonEncode(receipts),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to receive purchase order: ${response.body}');
    }
  }

  Future<void> closePurchaseOrder(String id) async {
    final response = await http.post(
      Uri.parse('$baseUrl/purchase-orders/$id/close'),
      headers: _headers,
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to close purchase order: ${response.body}');
    }
  }

  Future<void> addItemToPurchaseOrder(String poId, String itemId, int quantity, double unitCost) async {
    final response = await http.post(
      Uri.parse('$baseUrl/purchase-orders/$poId/items'),
      headers: _headers,
      body: jsonEncode({
        'item_id': itemId,
        'qty_ordered': quantity,
        'unit_cost': unitCost.toStringAsFixed(2),
      }),
    );

    if (response.statusCode != 200) {
      throw Exception('Failed to add item to purchase order: ${response.body}');
    }
  }
}