import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/user.dart';
import '../models/item.dart';

class ApiService {
  static const String baseUrl = 'http://192.168.8.135:8080/api/v1';
  // static const String baseUrl = 'http://192.168.0.44:8080/api/v1';
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

  Future<List<Item>> getItems({String? search, int page = 1, int limit = 20}) async {
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
      
      return itemsList
          .map((item) => Item.fromJson(item))
          .toList();
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

  Future<Item?> getItemByBarcode(String barcode) async {
    final response = await http.get(
      Uri.parse('$baseUrl/items/barcode/$barcode'),
      headers: _headers,
    );

    if (response.statusCode == 200) {
      return Item.fromJson(jsonDecode(response.body));
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
}