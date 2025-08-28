import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/user.dart';
import 'api_service.dart';

class AuthService extends ChangeNotifier {
  User? _user;
  String? _token;
  String? _refreshToken;
  bool _isAuthenticated = false;
  ApiService? _apiService;

  User? get user => _user;
  bool get isAuthenticated => _isAuthenticated;
  String? get token => _token;

  AuthService() {
    _loadFromStorage();
  }
  
  void setApiService(ApiService apiService) {
    _apiService = apiService;
    if (_token != null) {
      _apiService!.setToken(_token!);
    }
    // Set up automatic token refresh
    _apiService!.setOnTokenExpired(() => refreshAuthToken());
  }

  Future<void> _loadFromStorage() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('auth_token');
    _refreshToken = prefs.getString('refresh_token');
    
    if (_token != null) {
      if (_apiService != null) {
        _apiService!.setToken(_token!);
      }
      _isAuthenticated = true;
      
      final userJson = prefs.getString('user');
      if (userJson != null) {
        _user = User.fromJson(_parseUserJson(userJson));
      }
      
      notifyListeners();
    }
  }

  Map<String, dynamic> _parseUserJson(String userJson) {
    final parts = userJson.split('|');
    return {
      'id': parts[0],
      'email': parts[1],
      'name': parts[2],
      'role': parts[3],
      'tenant_id': parts.length > 4 ? parts[4] : null,
    };
  }

  String _userToString(User user) {
    return '${user.id}|${user.email}|${user.name}|${user.role}|${user.tenantId ?? ''}';
  }

  Future<bool> login(String email, String password) async {
    try {
      print('AuthService: Starting login process');
      if (_apiService == null) {
        print('AuthService: ApiService not set');
        return false;
      }
      
      final response = await _apiService!.login(email, password);
      print('AuthService: Received response: $response');
      
      _token = response['access_token'];
      _refreshToken = response['refresh_token'];
      _user = User.fromJson(response['user']);
      _isAuthenticated = true;
      
      print('AuthService: Parsed user: ${_user?.email}');
      
      _apiService!.setToken(_token!);
      
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('auth_token', _token!);
      await prefs.setString('refresh_token', _refreshToken!);
      await prefs.setString('user', _userToString(_user!));
      
      print('AuthService: Login successful');
      notifyListeners();
      return true;
    } catch (e) {
      print('AuthService: Login error: $e');
      return false;
    }
  }

  Future<void> logout() async {
    try {
      if (_token != null && _apiService != null) {
        await _apiService!.logout();
      }
    } catch (e) {
      // Ignore errors during logout API call
    }
    
    _user = null;
    _token = null;
    _refreshToken = null;
    _isAuthenticated = false;
    
    if (_apiService != null) {
      _apiService!.setToken('');
    }
    
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('auth_token');
    await prefs.remove('refresh_token');
    await prefs.remove('user');
    
    notifyListeners();
  }

  Future<bool> refreshAuthToken() async {
    if (_refreshToken == null || _apiService == null) return false;
    
    try {
      final response = await _apiService!.refreshToken(_refreshToken!);
      
      _token = response['access_token'];
      _apiService!.setToken(_token!);
      
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('auth_token', _token!);
      
      notifyListeners();
      return true;
    } catch (e) {
      await logout();
      return false;
    }
  }
}