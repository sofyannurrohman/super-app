import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter/widgets.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/utils/constants.dart';

class HeartbeatService extends WidgetsBindingObserver {
  final SharedPreferences prefs;
  Timer? _heartbeatTimer;
  bool _isForeground = true;
  bool _isRunning = false; // Guard: prevents concurrent in-flight requests

  // Dedicated Dio with short timeouts — completely isolated from the main API
  // client so a slow heartbeat never blocks other requests.
  late final Dio _heartbeatDio;

  HeartbeatService({required ApiClient apiClient, required this.prefs}) {
    String baseUrl = AppConstants.baseUrl;
    if (!baseUrl.endsWith('/')) baseUrl += '/';

    _heartbeatDio = Dio(
      BaseOptions(
        baseUrl: baseUrl,
        connectTimeout: const Duration(seconds: 5),
        receiveTimeout: const Duration(seconds: 5),
        sendTimeout: const Duration(seconds: 5),
      ),
    );
  }

  void init() {
    WidgetsBinding.instance.addObserver(this);
    _startHeartbeat();
  }

  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _stopHeartbeat();
  }

  void forceHeartbeat() {
    _sendHeartbeat();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _isForeground = true;
      _startHeartbeat();
    } else if (state == AppLifecycleState.paused ||
        state == AppLifecycleState.detached) {
      _isForeground = false;
      _stopHeartbeat();
    }
  }

  void _startHeartbeat() {
    _stopHeartbeat();

    // Send immediately once on foreground
    _sendHeartbeat();

    // 30-second interval — safe gap so the previous request always finishes
    // (or times out at 5s) before the next one fires.
    _heartbeatTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      if (_isForeground) _sendHeartbeat();
    });
  }

  void _stopHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = null;
  }

  Future<void> _sendHeartbeat() async {
    // Skip if a previous request is still in-flight
    if (_isRunning) return;

    final token = prefs.getString(AppConstants.tokenKey);
    if (token == null || token.isEmpty) return;

    _isRunning = true;
    try {
      await _heartbeatDio.post(
        'users/heartbeat',
        options: Options(
          headers: {'Authorization': 'Bearer $token'},
        ),
      );
    } catch (_) {
      // Heartbeat is best-effort telemetry — silently ignore all failures
      // (including timeout). Never block the user experience.
      debugPrint('Heartbeat skipped (network slow or unavailable)');
    } finally {
      _isRunning = false; // Always reset, even on error
    }
  }
}
