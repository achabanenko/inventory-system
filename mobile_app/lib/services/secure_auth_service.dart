import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/user.dart';
import 'api_service.dart';

// Secure storage keys
const String _accessTokenKey = 'secure_access_token';
const String _refreshTokenKey = 'secure_refresh_token';
const String _userDataKey = 'secure_user_data';
const String _biometricEnabledKey = 'biometric_enabled';

class SecureAuthService extends ChangeNotifier {
  User? _user;
  String? _token;
  String? _refreshToken;
  bool _isAuthenticated = false;
  bool _biometricEnabled = false;
  ApiService? _apiService;

  final FlutterSecureStorage _secureStorage = const FlutterSecureStorage();

  User? get user => _user;
  bool get isAuthenticated => _isAuthenticated;
  String? get token => _token;
  bool get biometricEnabled => _biometricEnabled;

  SecureAuthService() {
    _initialize();
  }

  /// Initialize the service by loading stored data
  Future<void> _initialize() async {
    await _loadBiometricPreference();
    await _loadFromSecureStorage();
  }

  /// Load biometric preference from shared preferences
  Future<void> _loadBiometricPreference() async {
    final prefs = await SharedPreferences.getInstance();
    _biometricEnabled = prefs.getBool(_biometricEnabledKey) ?? false;
    notifyListeners();
  }

  /// Enable or disable biometric authentication
  Future<void> setBiometricEnabled(bool enabled) async {
    _biometricEnabled = enabled;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_biometricEnabledKey, enabled);
    notifyListeners();

    if (!enabled) {
      // Clear tokens when disabling biometric auth
      await _clearSecureStorage();
    }
  }

  /// Load authentication data from secure storage
  Future<void> _loadFromSecureStorage() async {
    try {
      _token = await _secureStorage.read(key: _accessTokenKey);
      _refreshToken = await _secureStorage.read(key: _refreshTokenKey);
      final userJson = await _secureStorage.read(key: _userDataKey);

      if (_token != null && userJson != null) {
        _user = User.fromJson(_parseUserJson(userJson));
        _isAuthenticated = true;

        if (_apiService != null) {
          _apiService!.setToken(_token!);
        }

        notifyListeners();
      }
    } catch (e) {
      debugPrint('Error loading from secure storage: $e');
      // If secure storage fails, clear everything
      await _clearSecureStorage();
    }
  }

  /// Parse user JSON string into map
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

  /// Convert user object to string for storage
  String _userToString(User user) {
    return '${user.id}|${user.email}|${user.name}|${user.role}|${user.tenantId ?? ''}';
  }

  /// Clear all data from secure storage
  Future<void> _clearSecureStorage() async {
    await _secureStorage.deleteAll();
    _token = null;
    _refreshToken = null;
    _user = null;
    _isAuthenticated = false;
    notifyListeners();
  }

  /// Save authentication data to secure storage
  Future<void> _saveToSecureStorage() async {
    if (_token != null && _user != null) {
      await _secureStorage.write(key: _accessTokenKey, value: _token);
      if (_refreshToken != null) {
        await _secureStorage.write(key: _refreshTokenKey, value: _refreshToken);
      }
      await _secureStorage.write(key: _userDataKey, value: _userToString(_user!));
    }
  }

  void setApiService(ApiService apiService) {
    _apiService = apiService;
    if (_token != null) {
      _apiService!.setToken(_token!);
    }
    // Set up automatic token refresh
    _apiService!.setOnTokenExpired(() => refreshAuthToken());
  }

  /// Login with email and password
  Future<bool> login(String email, String password) async {
    try {
      debugPrint('SecureAuthService: Starting login process');
      if (_apiService == null) {
        debugPrint('SecureAuthService: ApiService not set');
        return false;
      }

      final response = await _apiService!.login(email, password);
      debugPrint('SecureAuthService: Received response');

      _token = response['access_token'];
      _refreshToken = response['refresh_token'];
      _user = User.fromJson(response['user']);
      _isAuthenticated = true;

      debugPrint('SecureAuthService: Parsed user: ${_user?.email}');

      _apiService!.setToken(_token!);

      // Save to secure storage
      await _saveToSecureStorage();

      debugPrint('SecureAuthService: Login successful');
      notifyListeners();
      return true;
    } catch (e) {
      debugPrint('SecureAuthService: Login error: $e');
      return false;
    }
  }

  /// Logout and clear all stored data
  Future<void> logout() async {
    try {
      if (_token != null && _apiService != null) {
        await _apiService!.logout();
      }
    } catch (e) {
      // Ignore errors during logout API call
      debugPrint('Error during API logout: $e');
    }

    // Clear local data
    await _clearSecureStorage();

    if (_apiService != null) {
      _apiService!.setToken('');
    }

    notifyListeners();
  }

  /// Refresh access token using stored refresh token
  Future<bool> refreshAuthToken() async {
    if (_refreshToken == null || _apiService == null) {
      debugPrint('No refresh token or API service available');
      return false;
    }

    try {
      debugPrint('Refreshing access token...');
      final response = await _apiService!.refreshToken(_refreshToken!);

      _token = response['access_token'];

      if (_token != null) {
        _apiService!.setToken(_token!);
        // Update stored token
        await _secureStorage.write(key: _accessTokenKey, value: _token);
        debugPrint('Token refreshed successfully');
        notifyListeners();
        return true;
      } else {
        debugPrint('Failed to get new access token');
        await logout();
        return false;
      }
    } catch (e) {
      debugPrint('Token refresh failed: $e');
      await logout();
      return false;
    }
  }

  /// Check if biometric authentication is available and tokens exist
  Future<bool> canUseBiometricAuth() async {
    return _biometricEnabled && _refreshToken != null && _refreshToken!.isNotEmpty;
  }
}
