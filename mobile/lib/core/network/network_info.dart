import 'dart:async';
import 'package:flutter/foundation.dart' show kIsWeb, debugPrint;
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:internet_connection_checker_plus/internet_connection_checker_plus.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:dio/dio.dart';

enum AppNetworkStatus {
  online,
  noInternet,
  serverUnreachable,
  timeout,
  pendingSync,
}

class NetworkInfo {
  static final NetworkInfo _instance = NetworkInfo._internal();
  factory NetworkInfo() => _instance;

  late InternetConnection _internetConnection;
  final Connectivity _connectivity = Connectivity();
  bool _isInitialized = false;

  // Cached internet access state (TTL 15 seconds)
  bool? _cachedHasInternet;
  DateTime? _lastInternetCheckTime;
  static const Duration _cacheTtl = Duration(seconds: 15);

  NetworkInfo._internal() {
    _initialize();
  }

  void _initialize() {
    if (_isInitialized) return;

    String healthUrl = AppConstants.baseUrl;
    if (healthUrl.endsWith('/')) {
      healthUrl += 'health';
    } else {
      healthUrl += '/health';
    }

    _internetConnection = InternetConnection.createInstance(
      customCheckOptions: [
        InternetCheckOption(
          uri: Uri.parse(healthUrl),
          timeout: const Duration(seconds: 8), // Increased from 3s to 8s
        ),
      ],
      useDefaultOptions: true, // Also checks Google/Cloudflare DNS as fallback
    );
    _isInitialized = true;
  }

  InternetConnection get connection => _internetConnection;

  /// Returns whether device has internet access (cached for 15s to prevent unnecessary latency spikes).
  Future<bool> get hasInternetAccess async {
    if (kIsWeb) {
      final results = await _connectivity.checkConnectivity();
      return !results.contains(ConnectivityResult.none);
    }

    final now = DateTime.now();
    if (_cachedHasInternet != null &&
        _lastInternetCheckTime != null &&
        now.difference(_lastInternetCheckTime!) < _cacheTtl) {
      return _cachedHasInternet!;
    }

    try {
      final result = await _internetConnection.hasInternetAccess;
      _cachedHasInternet = result;
      _lastInternetCheckTime = now;
      return result;
    } catch (e) {
      debugPrint('[NetworkInfo] Error checking internet access: $e');
      return _cachedHasInternet ?? false;
    }
  }

  /// Force an immediate un-cached check of internet access
  Future<bool> checkInternetNow() async {
    if (kIsWeb) {
      final results = await _connectivity.checkConnectivity();
      return !results.contains(ConnectivityResult.none);
    }
    final result = await _internetConnection.hasInternetAccess;
    _cachedHasInternet = result;
    _lastInternetCheckTime = DateTime.now();
    return result;
  }

  /// Detects current physical network type ('wifi', 'cellular', 'ethernet', 'none', 'unknown')
  Future<String> getNetworkType() async {
    if (kIsWeb) {
      final results = await _connectivity.checkConnectivity();
      if (results.contains(ConnectivityResult.none)) return 'none';
      return 'web';
    }
    try {
      final results = await _connectivity.checkConnectivity();
      if (results.contains(ConnectivityResult.wifi)) return 'wifi';
      if (results.contains(ConnectivityResult.mobile)) return 'cellular';
      if (results.contains(ConnectivityResult.ethernet)) return 'ethernet';
      if (results.contains(ConnectivityResult.none)) return 'none';
      return 'other';
    } catch (e) {
      return 'unknown';
    }
  }

  /// Measures health check round-trip latency in milliseconds.
  /// Returns null if server is unreachable or timed out.
  Future<int?> checkServerHealthLatency() async {
    final stopwatch = Stopwatch()..start();
    try {
      String healthUrl = AppConstants.baseUrl;
      if (healthUrl.endsWith('/')) {
        healthUrl += 'health';
      } else {
        healthUrl += '/health';
      }

      final dio = Dio(BaseOptions(
        connectTimeout: const Duration(seconds: 8),
        receiveTimeout: const Duration(seconds: 8),
      ));

      final res = await dio.get(healthUrl);
      stopwatch.stop();
      if (res.statusCode == 200) {
        return stopwatch.elapsedMilliseconds;
      }
      return null;
    } catch (e) {
      stopwatch.stop();
      return null;
    }
  }

  Stream<InternetStatus> get onStatusChange {
    if (kIsWeb) {
      return _connectivity.onConnectivityChanged.map((results) {
        return results.contains(ConnectivityResult.none)
            ? InternetStatus.disconnected
            : InternetStatus.connected;
      });
    }
    return _internetConnection.onStatusChange;
  }
}
