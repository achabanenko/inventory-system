import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/secure_auth_service.dart';
import '../services/biometric_service.dart';
import '../screens/login_screen.dart';
import '../screens/home_screen.dart';

/// Authentication state machine states
enum AuthState {
  /// Initial state - checking authentication
  checking,
  /// Biometric authentication required
  biometricRequired,
  /// Biometric authentication in progress
  biometricAuthenticating,
  /// Token refresh in progress
  refreshingToken,
  /// User is authenticated
  authenticated,
  /// User needs to login
  needsLogin,
  /// Authentication failed
  authFailed,
}

/// Widget that manages the authentication flow with biometric support
class AuthWrapper extends StatelessWidget {
  const AuthWrapper({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<SecureAuthService>(
      builder: (context, authService, child) {
        return _AuthWrapperContent(authService: authService);
      },
    );
  }
}

class _AuthWrapperContent extends StatefulWidget {
  final SecureAuthService authService;

  const _AuthWrapperContent({required this.authService});

  @override
  State<_AuthWrapperContent> createState() => _AuthWrapperContentState();
}

class _AuthWrapperContentState extends State<_AuthWrapperContent> {
  AuthState _authState = AuthState.checking;
  String? _errorMessage;
  final BiometricService _biometricService = BiometricService();

  @override
  void initState() {
    super.initState();
    _initializeAuthentication();
  }

  @override
  void didUpdateWidget(_AuthWrapperContent oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Check if authentication state changed (e.g., user logged out)
    if (!widget.authService.isAuthenticated && _authState == AuthState.authenticated) {
      setState(() => _authState = AuthState.needsLogin);
    }
  }

  /// Initialize the authentication flow
  Future<void> _initializeAuthentication() async {
    // Wait a moment for the auth service to initialize
    await Future.delayed(const Duration(milliseconds: 500));

    // Check if user is already authenticated
    if (widget.authService.isAuthenticated) {
      setState(() => _authState = AuthState.authenticated);
      return;
    }

    // Check if biometric authentication is possible
    final canUseBiometric = await widget.authService.canUseBiometricAuth();
    final biometricAvailable = await _biometricService.isBiometricAvailable();

    if (canUseBiometric && biometricAvailable) {
      setState(() => _authState = AuthState.biometricRequired);
      await _attemptBiometricAuthentication();
    } else {
      setState(() => _authState = AuthState.needsLogin);
    }
  }

  /// Attempt biometric authentication
  Future<void> _attemptBiometricAuthentication() async {
    setState(() => _authState = AuthState.biometricAuthenticating);

    try {
      final biometricDescription = await _biometricService.getBiometricDescription();
      final authenticated = await _biometricService.authenticate(
        reason: 'Authenticate with $biometricDescription to access your account',
        title: 'Unlock Inventory Manager',
      );

      if (authenticated) {
        setState(() => _authState = AuthState.refreshingToken);
        await _refreshToken();
      } else {
        // Biometric authentication failed or was cancelled
        setState(() => _authState = AuthState.needsLogin);
      }
    } catch (e) {
      debugPrint('Biometric authentication error: $e');
      setState(() {
        _authState = AuthState.needsLogin;
        _errorMessage = 'Biometric authentication failed. Please login with your credentials.';
      });
    }
  }

  /// Refresh the access token using the stored refresh token
  Future<void> _refreshToken() async {
    try {
      final success = await widget.authService.refreshAuthToken();

      if (success) {
        setState(() => _authState = AuthState.authenticated);
      } else {
        setState(() {
          _authState = AuthState.needsLogin;
          _errorMessage = 'Session expired. Please login again.';
        });
      }
    } catch (e) {
      debugPrint('Token refresh error: $e');
      setState(() {
        _authState = AuthState.needsLogin;
        _errorMessage = 'Failed to refresh session. Please login again.';
      });
    }
  }



  /// Get the appropriate widget based on authentication state
  Widget _buildStateWidget() {
    switch (_authState) {
      case AuthState.checking:
        return _buildLoadingScreen('Checking authentication...');

      case AuthState.biometricRequired:
        return _buildBiometricPrompt();

      case AuthState.biometricAuthenticating:
        return _buildLoadingScreen('Authenticating...');

      case AuthState.refreshingToken:
        return _buildLoadingScreen('Refreshing session...');

      case AuthState.authenticated:
        return const HomeScreen();

      case AuthState.needsLogin:
        return LoginScreen(
          onLoginSuccess: () {
            setState(() => _authState = AuthState.authenticated);
          },
          onLoginFailure: () {
            setState(() {
              _authState = AuthState.needsLogin;
              _errorMessage = null;
            });
          },
          errorMessage: _errorMessage,
        );

      case AuthState.authFailed:
        return _buildErrorScreen();
    }
  }

  /// Build loading screen with message
  Widget _buildLoadingScreen(String message) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const CircularProgressIndicator(),
            const SizedBox(height: 24),
            Text(
              message,
              style: Theme.of(context).textTheme.bodyLarge,
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  /// Build biometric authentication prompt
  Widget _buildBiometricPrompt() {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              const Icon(
                Icons.fingerprint,
                size: 80,
                color: Color(0xFF2563EB),
              ),
              const SizedBox(height: 32),
              Text(
                'Unlock Inventory Manager',
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: const Color(0xFF2563EB),
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 16),
              const Text(
                'Use your biometric authentication to quickly access your account',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 16),
              ),
              const SizedBox(height: 48),
              ElevatedButton.icon(
                onPressed: _attemptBiometricAuthentication,
                icon: const Icon(Icons.fingerprint),
                label: const Text('Authenticate'),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
                ),
              ),
              const SizedBox(height: 24),
              TextButton(
                onPressed: () {
                  setState(() => _authState = AuthState.needsLogin);
                },
                child: const Text('Use password instead'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// Build error screen
  Widget _buildErrorScreen() {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              const Icon(
                Icons.error_outline,
                size: 80,
                color: Colors.red,
              ),
              const SizedBox(height: 24),
              Text(
                'Authentication Error',
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 16),
              Text(
                _errorMessage ?? 'An authentication error occurred.',
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 16),
              ),
              const SizedBox(height: 32),
              ElevatedButton(
                onPressed: () {
                  setState(() {
                    _authState = AuthState.needsLogin;
                    _errorMessage = null;
                  });
                },
                child: const Text('Try Again'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return _buildStateWidget();
  }
}
