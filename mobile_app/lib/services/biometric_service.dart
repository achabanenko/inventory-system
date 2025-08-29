import 'package:flutter/foundation.dart';
import 'package:local_auth/local_auth.dart';
import 'package:local_auth_android/local_auth_android.dart';
import 'package:local_auth_darwin/local_auth_darwin.dart';
import 'package:local_auth_platform_interface/types/auth_messages.dart';

/// Service for handling biometric authentication (Face ID, Touch ID, fingerprint)
class BiometricService {
  final LocalAuthentication _localAuth = LocalAuthentication();

  /// Check if biometric authentication is available on the device
  Future<bool> isBiometricAvailable() async {
    try {
      final isDeviceSupported = await _localAuth.isDeviceSupported();
      final availableBiometrics = await _localAuth.getAvailableBiometrics();

      debugPrint('Device supported: $isDeviceSupported');
      debugPrint('Available biometrics: $availableBiometrics');

      return isDeviceSupported &&
             availableBiometrics.isNotEmpty &&
             (availableBiometrics.contains(BiometricType.face) ||
              availableBiometrics.contains(BiometricType.fingerprint) ||
              availableBiometrics.contains(BiometricType.strong));
    } catch (e) {
      debugPrint('Error checking biometric availability: $e');
      return false;
    }
  }

  /// Get list of available biometric types
  Future<List<BiometricType>> getAvailableBiometrics() async {
    try {
      return await _localAuth.getAvailableBiometrics();
    } catch (e) {
      debugPrint('Error getting available biometrics: $e');
      return [];
    }
  }

  /// Authenticate using biometrics
  Future<bool> authenticate({
    String reason = 'Please authenticate to continue',
    String? title,
    bool stickyAuth = true,
    bool useErrorDialogs = true,
  }) async {
    try {
      final authenticated = await _localAuth.authenticate(
        localizedReason: reason,
        authMessages: title != null ? <AuthMessages>[
          AndroidAuthMessages(
            signInTitle: title,
            biometricHint: '',
            cancelButton: 'Cancel',
          ),
          IOSAuthMessages(
            cancelButton: 'Cancel',
          ),
        ] : <AuthMessages>[],
        options: const AuthenticationOptions(
          biometricOnly: true,
        ),
      );

      debugPrint('Biometric authentication result: $authenticated');
      return authenticated;
    } catch (e) {
      debugPrint('Biometric authentication error: $e');
      return false;
    }
  }

  /// Get a user-friendly description of available biometric methods
  Future<String> getBiometricDescription() async {
    try {
      final biometrics = await getAvailableBiometrics();

      if (biometrics.isEmpty) {
        return 'Biometric authentication';
      }

      final descriptions = <String>[];

      for (final biometric in biometrics) {
        switch (biometric) {
          case BiometricType.face:
            descriptions.add('Face ID');
            break;
          case BiometricType.fingerprint:
            descriptions.add('Fingerprint');
            break;
          case BiometricType.iris:
            descriptions.add('Iris scan');
            break;
          case BiometricType.strong:
            descriptions.add('Device authentication');
            break;
          case BiometricType.weak:
            descriptions.add('Weak authentication');
            break;
        }
      }

      return descriptions.join(', ');
    } catch (e) {
      debugPrint('Error getting biometric description: $e');
      return 'Biometric authentication';
    }
  }

  /// Check if the device supports biometric authentication
  Future<bool> canAuthenticate() async {
    try {
      return await _localAuth.isDeviceSupported();
    } catch (e) {
      debugPrint('Error checking authentication capability: $e');
      return false;
    }
  }
}
