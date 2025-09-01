import 'dart:async';\nimport 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';\nimport 'package:flutter/material.dart';\nimport 'package:http/http.dart' as http;

/// Performance monitoring service to track metrics as specified in mobile_edit_batch.md:
/// - App bundle size under 25MB
/// - API response times under 500ms  
/// - Maintain 55+ fps during scrolling
/// - Time-to-interactive targets
class PerformanceService {
  static PerformanceService? _instance;
  static PerformanceService get instance => _instance ??= PerformanceService._();
  
  PerformanceService._();

  // Performance metrics
  final Map<String, List<Duration>> _apiResponseTimes = {};
  final List<double> _frameRenderTimes = [];
  final Map<String, DateTime> _navigationStartTimes = {};
  
  // Targets from mobile_edit_batch.md
  static const Duration maxApiResponseTime = Duration(milliseconds: 500);
  static const double targetFps = 55.0;
  static const Duration maxTimeToInteractive = Duration(milliseconds: 300);
  
  // Callbacks for performance alerts
  final List<void Function(PerformanceAlert)> _alertCallbacks = [];

  /// Start tracking API response time
  PerformanceTimer startApiTimer(String endpoint) {
    return PerformanceTimer._(endpoint, this);
  }

  /// Record API response time
  void _recordApiResponseTime(String endpoint, Duration duration) {
    if (!_apiResponseTimes.containsKey(endpoint)) {
      _apiResponseTimes[endpoint] = [];
    }
    
    _apiResponseTimes[endpoint]!.add(duration);
    
    // Keep only last 100 measurements per endpoint
    if (_apiResponseTimes[endpoint]!.length > 100) {
      _apiResponseTimes[endpoint]!.removeAt(0);
    }
    
    // Check if response time exceeds target
    if (duration > maxApiResponseTime) {
      _triggerAlert(PerformanceAlert(
        type: PerformanceAlertType.slowApiResponse,
        message: 'API $endpoint took ${duration.inMilliseconds}ms (target: ${maxApiResponseTime.inMilliseconds}ms)',
        value: duration.inMilliseconds.toDouble(),
        threshold: maxApiResponseTime.inMilliseconds.toDouble(),
      ));
    }
  }

  /// Record frame render time
  void recordFrameTime(Duration renderTime) {
    final fps = 1000 / renderTime.inMicroseconds * 1000;
    _frameRenderTimes.add(fps);
    
    // Keep only last 60 measurements (about 1 second at 60fps)
    if (_frameRenderTimes.length > 60) {
      _frameRenderTimes.removeAt(0);
    }
    
    // Check if FPS drops below target
    if (fps < targetFps) {
      _triggerAlert(PerformanceAlert(
        type: PerformanceAlertType.lowFps,
        message: 'Frame rate dropped to ${fps.toStringAsFixed(1)} fps (target: $targetFps fps)',
        value: fps,
        threshold: targetFps,
      ));
    }
  }

  /// Start navigation timing
  void startNavigation(String routeName) {
    _navigationStartTimes[routeName] = DateTime.now();
  }

  /// End navigation timing
  void endNavigation(String routeName) {
    final startTime = _navigationStartTimes[routeName];
    if (startTime != null) {
      final duration = DateTime.now().difference(startTime);
      _navigationStartTimes.remove(routeName);
      
      if (duration > maxTimeToInteractive) {
        _triggerAlert(PerformanceAlert(
          type: PerformanceAlertType.slowNavigation,
          message: 'Navigation to $routeName took ${duration.inMilliseconds}ms (target: ${maxTimeToInteractive.inMilliseconds}ms)',
          value: duration.inMilliseconds.toDouble(),
          threshold: maxTimeToInteractive.inMilliseconds.toDouble(),
        ));
      }
    }
  }

  /// Get performance statistics
  PerformanceStats getStats() {
    final avgApiTimes = <String, double>{};
    _apiResponseTimes.forEach((endpoint, times) {
      if (times.isNotEmpty) {
        final avg = times.map((t) => t.inMilliseconds).reduce((a, b) => a + b) / times.length;
        avgApiTimes[endpoint] = avg;
      }
    });

    final avgFps = _frameRenderTimes.isNotEmpty 
        ? _frameRenderTimes.reduce((a, b) => a + b) / _frameRenderTimes.length
        : 0.0;

    return PerformanceStats(
      averageApiResponseTimes: avgApiTimes,
      averageFps: avgFps,
      currentFrameRate: _frameRenderTimes.isNotEmpty ? _frameRenderTimes.last : 0.0,
    );
  }

  /// Get app bundle size (platform-specific)
  Future<double> getAppBundleSize() async {
    try {
      if (Platform.isAndroid) {
        // Android APK size check would go here
        // For now, return a placeholder
        return 0.0;
      } else if (Platform.isIOS) {
        // iOS IPA size check would go here
        return 0.0;
      }
    } catch (e) {
      debugPrint('Error getting app bundle size: $e');
    }
    return 0.0;
  }

  /// Monitor memory usage
  Future<MemoryInfo> getMemoryInfo() async {
    // This would use platform channels to get actual memory info
    // For now, return placeholder values
    return MemoryInfo(
      usedMemoryMB: 0.0,
      totalMemoryMB: 0.0,
      memoryPressure: MemoryPressure.normal,
    );
  }

  /// Add performance alert callback
  void addAlertCallback(void Function(PerformanceAlert) callback) {
    _alertCallbacks.add(callback);
  }

  /// Remove performance alert callback
  void removeAlertCallback(void Function(PerformanceAlert) callback) {
    _alertCallbacks.remove(callback);
  }

  /// Clear all performance data
  void clearMetrics() {
    _apiResponseTimes.clear();
    _frameRenderTimes.clear();
    _navigationStartTimes.clear();
  }

  /// Export performance data for analysis
  Map<String, dynamic> exportMetrics() {
    return {
      'timestamp': DateTime.now().toIso8601String(),
      'apiResponseTimes': _apiResponseTimes.map((key, value) => 
          MapEntry(key, value.map((d) => d.inMilliseconds).toList())),
      'frameRenderTimes': _frameRenderTimes,
      'stats': getStats().toJson(),
    };
  }

  void _triggerAlert(PerformanceAlert alert) {
    for (final callback in _alertCallbacks) {
      try {
        callback(alert);
      } catch (e) {
        debugPrint('Error in performance alert callback: $e');
      }
    }
  }
}

/// Timer for measuring API response times
class PerformanceTimer {
  final String _endpoint;
  final PerformanceService _service;
  final DateTime _startTime;

  PerformanceTimer._(this._endpoint, this._service) : _startTime = DateTime.now();

  /// Stop the timer and record the measurement
  void stop() {
    final duration = DateTime.now().difference(_startTime);
    _service._recordApiResponseTime(_endpoint, duration);
  }
}

/// Performance statistics
class PerformanceStats {
  final Map<String, double> averageApiResponseTimes;
  final double averageFps;
  final double currentFrameRate;

  const PerformanceStats({
    required this.averageApiResponseTimes,
    required this.averageFps,
    required this.currentFrameRate,
  });

  Map<String, dynamic> toJson() {
    return {
      'averageApiResponseTimes': averageApiResponseTimes,
      'averageFps': averageFps,
      'currentFrameRate': currentFrameRate,
    };
  }
}

/// Performance alert types
enum PerformanceAlertType {
  slowApiResponse,
  lowFps,
  slowNavigation,
  highMemoryUsage,
  largeBundleSize,
}

/// Performance alert
class PerformanceAlert {
  final PerformanceAlertType type;
  final String message;
  final double value;
  final double threshold;

  const PerformanceAlert({
    required this.type,
    required this.message,
    required this.value,
    required this.threshold,
  });

  @override
  String toString() => '$type: $message';
}

/// Memory information
class MemoryInfo {
  final double usedMemoryMB;
  final double totalMemoryMB;
  final MemoryPressure memoryPressure;

  const MemoryInfo({
    required this.usedMemoryMB,
    required this.totalMemoryMB,
    required this.memoryPressure,
  });

  double get usagePercentage => totalMemoryMB > 0 ? (usedMemoryMB / totalMemoryMB) * 100 : 0;
}

/// Memory pressure levels
enum MemoryPressure {
  normal,
  warning,
  critical,
}

/// Extension to add performance monitoring to HTTP requests
extension PerformanceHttpClient on http.Client {
  Future<http.Response> getWithTiming(Uri url, {Map<String, String>? headers}) async {
    final timer = PerformanceService.instance.startApiTimer(url.path);
    try {
      final response = await get(url, headers: headers);
      return response;
    } finally {
      timer.stop();
    }
  }

  Future<http.Response> postWithTiming(Uri url, {Map<String, String>? headers, Object? body, Encoding? encoding}) async {
    final timer = PerformanceService.instance.startApiTimer(url.path);
    try {
      final response = await post(url, headers: headers, body: body, encoding: encoding);
      return response;
    } finally {
      timer.stop();
    }
  }
}

/// Widget mixin for performance monitoring
mixin PerformanceMonitorMixin<T extends StatefulWidget> on State<T> {
  late DateTime _buildStartTime;

  @override
  void initState() {
    super.initState();
    PerformanceService.instance.startNavigation(widget.runtimeType.toString());
  }

  @override
  Widget build(BuildContext context) {
    _buildStartTime = DateTime.now();
    final widget = buildWithTiming(context);
    
    // Schedule post-frame callback to measure build time
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final buildTime = DateTime.now().difference(_buildStartTime);
      if (buildTime.inMilliseconds > 16) { // More than one frame at 60fps
        debugPrint('Slow build detected: ${widget.runtimeType} took ${buildTime.inMilliseconds}ms');
      }
      
      PerformanceService.instance.endNavigation(widget.runtimeType.toString());
    });
    
    return widget;
  }

  /// Override this instead of build()
  Widget buildWithTiming(BuildContext context);
}