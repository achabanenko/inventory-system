import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
// import 'package:google_sign_in/google_sign_in.dart'; // Temporarily commented
import 'dart:convert';
import 'dart:async';
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
  // final GoogleSignIn _googleSignIn = GoogleSignIn(
  //   scopes: ['email', 'profile'],
  //   serverClientId: '778454408539-06ffskhonrbilehq1q1e241rpffn6let.apps.googleusercontent.com',
  // ); // Temporarily commented
  Timer? _refreshTimer;

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

      print('SecureAuthService: Loading from secure storage...');
      print('SecureAuthService: Access token exists: ${_token != null}');
      print('SecureAuthService: Refresh token exists: ${_refreshToken != null}');
      print('SecureAuthService: User data exists: ${userJson != null}');

      if (_token != null && userJson != null) {
        // Check if token is expired and refresh if needed
        if (_isTokenExpired(_token!) && _refreshToken != null) {
          print('SecureAuthService: Access token expired, attempting refresh...');
          final refreshSuccess = await refreshAuthToken();
          if (!refreshSuccess) {
            print('SecureAuthService: Token refresh failed, clearing storage');
            await _clearSecureStorage();
            return;
          }
        }

        _user = User.fromJson(_parseUserJson(userJson));
        _isAuthenticated = true;

        print('SecureAuthService: User loaded: ${_user!.name} (${_user!.email})');
        print('SecureAuthService: Access token: ***${_token!.substring(_token!.length - 10)}');

        if (_apiService != null) {
          print('SecureAuthService: Setting token in API service');
          _apiService!.setToken(_token!);
        }

        // Set up automatic token refresh
        _scheduleTokenRefresh();

        notifyListeners();
      } else {
        print('SecureAuthService: Missing required data for authentication');
        _isAuthenticated = false;
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
      print('SecureAuthService: Saving to secure storage...');
      print('SecureAuthService: Saving access token: ***${_token!.substring(_token!.length - 10)}');
      print('SecureAuthService: Saving refresh token: ***${_refreshToken?.substring(_refreshToken!.length - 10) ?? "null"}');
      print('SecureAuthService: Saving user: ${_user!.name} (${_user!.email})');

      await _secureStorage.write(key: _accessTokenKey, value: _token);
      if (_refreshToken != null) {
        await _secureStorage.write(key: _refreshTokenKey, value: _refreshToken);
      }
      await _secureStorage.write(key: _userDataKey, value: _userToString(_user!));

      print('SecureAuthService: Successfully saved to secure storage');
    } else {
      print('SecureAuthService: Cannot save - token or user is null');
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

      // Set up automatic token refresh
      _scheduleTokenRefresh();

      debugPrint('SecureAuthService: Login successful');
      notifyListeners();
      return true;
    } catch (e) {
      debugPrint('SecureAuthService: Login error: $e');
      return false;
    }
  }

  /// Login with Google OAuth - Temporarily disabled
  Future<bool> loginWithGoogle() async {
    debugPrint('SecureAuthService: Google OAuth temporarily disabled');
    return false;
    /*
    try {
      debugPrint('SecureAuthService: Starting Google OAuth login');
      if (_apiService == null) {
        debugPrint('SecureAuthService: ApiService not set');
        return false;
      }

      // Sign out first to ensure fresh authentication
      // await _googleSignIn.signOut(); // Temporarily disabled

      // Initiate Google Sign-In
      final GoogleSignInAccount? googleAccount = await _googleSignIn.signIn();
      if (googleAccount == null) {
        debugPrint('SecureAuthService: Google sign-in cancelled by user');
        return false;
      }

      // Get authentication details
      final GoogleSignInAuthentication googleAuth = await googleAccount.authentication;
      final String? authCode = googleAuth.serverAuthCode;
      
      if (authCode == null) {
        debugPrint('SecureAuthService: No server auth code received');
        return false;
      }

      debugPrint('SecureAuthService: Received auth code from Google');

      // Exchange auth code for tokens via backend
      const redirectUri = 'com.inventory.app:/oauth'; // Custom scheme for mobile
      final response = await _apiService!.googleOAuth(authCode, redirectUri);
      debugPrint('SecureAuthService: Received backend response');

      _token = response['access_token'];
      _refreshToken = response['refresh_token'];
      _user = User.fromJson(response['user']);
      _isAuthenticated = true;

      debugPrint('SecureAuthService: Parsed user: ${_user?.email}');

      _apiService!.setToken(_token!);

      // Save to secure storage
      await _saveToSecureStorage();

      // Set up automatic token refresh
      _scheduleTokenRefresh();

      debugPrint('SecureAuthService: Google OAuth login successful');
      notifyListeners();
      return true;
    } catch (e) {
      debugPrint('SecureAuthService: Google OAuth login error: $e');
      // Sign out from Google on error
      // await _googleSignIn.signOut(); // Temporarily disabled
      return false;
    }
    */
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

    // Sign out from Google
    try {
      // await _googleSignIn.signOut(); // Temporarily disabled
    } catch (e) {
      debugPrint('Error during Google sign-out: $e');
    }

    // Cancel automatic refresh timer
    _refreshTimer?.cancel();

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
        
        // Schedule next automatic refresh
        _scheduleTokenRefresh();
        
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

  /// Check if a JWT token is expired
  bool _isTokenExpired(String token) {
    try {
      final parts = token.split('.');
      if (parts.length != 3) return true;

      // Decode payload (base64)
      final payload = parts[1];
      final decodedBytes = base64Url.decode(base64Url.normalize(payload));
      final decodedPayload = json.decode(utf8.decode(decodedBytes));

      final exp = decodedPayload['exp'] as int?;
      if (exp == null) return true;

      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      return now >= exp;
    } catch (e) {
      debugPrint('Error checking token expiration: $e');
      return true;
    }
  }

  /// Check if a JWT token will expire within the next N seconds
  bool _isTokenExpiringSoon(String token, {int seconds = 60}) {
    try {
      final parts = token.split('.');
      if (parts.length != 3) return true;

      // Decode payload (base64)
      final payload = parts[1];
      final decodedBytes = base64Url.decode(base64Url.normalize(payload));
      final decodedPayload = json.decode(utf8.decode(decodedBytes));

      final exp = decodedPayload['exp'] as int?;
      if (exp == null) return true;

      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      return (exp - now) <= seconds;
    } catch (e) {
      debugPrint('Error checking token expiration: $e');
      return true;
    }
  }

  /// Schedule automatic token refresh
  void _scheduleTokenRefresh() {
    _refreshTimer?.cancel(); // Cancel existing timer

    if (_token == null) return;

    try {
      final parts = _token!.split('.');
      if (parts.length != 3) return;

      // Decode payload (base64)
      final payload = parts[1];
      final decodedBytes = base64Url.decode(base64Url.normalize(payload));
      final decodedPayload = json.decode(utf8.decode(decodedBytes));

      final exp = decodedPayload['exp'] as int?;
      if (exp == null) return;

      final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      final expiryTime = exp;
      final refreshTime = expiryTime - 60; // Refresh 60 seconds before expiry
      final delay = Duration(seconds: refreshTime - now);

      if (delay.inSeconds > 0) {
        debugPrint('SecureAuthService: Scheduling token refresh in ${delay.inSeconds} seconds');
        
        _refreshTimer = Timer(delay, () async {
          debugPrint('SecureAuthService: Auto-refreshing token');
          await refreshAuthToken();
        });
      } else {
        debugPrint('SecureAuthService: Token expires soon, refreshing immediately');
        // Token expires very soon, refresh immediately
        Future.delayed(const Duration(seconds: 1), () async {
          await refreshAuthToken();
        });
      }
    } catch (e) {
      debugPrint('Error scheduling token refresh: $e');
    }
  }

  /// Dispose method to clean up resources
  @override
  void dispose() {
    _refreshTimer?.cancel();
    super.dispose();
  }
}
