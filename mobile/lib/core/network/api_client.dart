import 'dart:async';
import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/network/network_info.dart';

class ApiClient {
  late Dio dio;
  final SharedPreferences prefs;

  // Timeout constants – tuned for responsive mobile & field networks (fast-fail)
  static const Duration _connectTimeout = Duration(seconds: 8);
  static const Duration _receiveTimeout = Duration(seconds: 10);
  static const Duration _sendTimeout = Duration(seconds: 10);

  // Max automatic retries on timeout (GET only) – 1 retry with short delay to avoid long freezes
  static const int _maxRetries = 1;

  // StreamController to broadcast 401 Unauthorized events with optional message
  final StreamController<String?> _unauthorizedController =
      StreamController<String?>.broadcast();
  Stream<String?> get onUnauthorized => _unauthorizedController.stream;

  ApiClient({required this.prefs}) {
    dio = Dio(
      BaseOptions(
        baseUrl: AppConstants.baseUrl,
        connectTimeout: _connectTimeout,
        receiveTimeout: _receiveTimeout,
        sendTimeout: _sendTimeout,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ),
    );

    // ── Interceptor 1: Diagnostics & Performance Logging
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          options.extra['_requestStartTime'] = DateTime.now().millisecondsSinceEpoch;
          
          // Estimate payload size
          int payloadSize = 0;
          if (options.data is String) {
            payloadSize = utf8.encode(options.data as String).length;
          } else if (options.data is Map || options.data is List) {
            try {
              payloadSize = utf8.encode(jsonEncode(options.data)).length;
            } catch (_) {}
          } else if (options.data is FormData) {
            final fd = options.data as FormData;
            payloadSize = fd.length;
          }
          options.extra['_payloadSize'] = payloadSize;

          // Extract client_id if available in data
          String? clientId;
          if (options.data is Map) {
            clientId = options.data['client_id']?.toString();
          }
          if (clientId != null) {
            options.extra['_clientId'] = clientId;
          }

          return handler.next(options);
        },
        onResponse: (response, handler) async {
          final start = response.requestOptions.extra['_requestStartTime'] as int?;
          final latency = start != null ? (DateTime.now().millisecondsSinceEpoch - start) : -1;
          final netType = await NetworkInfo().getNetworkType();
          final payloadSize = response.requestOptions.extra['_payloadSize'] ?? 0;
          final clientId = response.requestOptions.extra['_clientId'] ?? '-';

          debugPrint(
            '[API-LOG] [NET:$netType] [${response.requestOptions.method}] ${response.requestOptions.path} '
            '[HTTP:${response.statusCode}] [LATENCY:${latency}ms] [REQ_SIZE:${payloadSize}B] [CLIENT_ID:$clientId]',
          );
          return handler.next(response);
        },
        onError: (DioException e, handler) async {
          final start = e.requestOptions.extra['_requestStartTime'] as int?;
          final latency = start != null ? (DateTime.now().millisecondsSinceEpoch - start) : -1;
          final netType = await NetworkInfo().getNetworkType();
          final payloadSize = e.requestOptions.extra['_payloadSize'] ?? 0;
          final clientId = e.requestOptions.extra['_clientId'] ?? '-';
          final statusCode = e.response?.statusCode ?? 0;

          // Differentiate clear failure categories
          String errorCategory = 'UNKNOWN';
          if (e.type == DioExceptionType.connectionTimeout ||
              e.type == DioExceptionType.sendTimeout ||
              e.type == DioExceptionType.receiveTimeout) {
            errorCategory = 'TIMEOUT';
          } else if (e.type == DioExceptionType.connectionError) {
            errorCategory = 'CONNECTION_ERROR';
          } else if (statusCode >= 500) {
            errorCategory = 'SERVER_ERROR_5XX';
          } else if (statusCode == 401) {
            errorCategory = 'UNAUTHORIZED_401';
          } else if (statusCode == 404) {
            errorCategory = 'NOT_FOUND_404';
          }

          debugPrint(
            '[API-ERR] [CATEGORY:$errorCategory] [NET:$netType] [${e.requestOptions.method}] ${e.requestOptions.path} '
            '[HTTP:$statusCode] [LATENCY:${latency}ms] [REQ_SIZE:${payloadSize}B] [CLIENT_ID:$clientId] [ERR:${e.message}]',
          );

          return handler.next(e);
        },
      ),
    );

    // ── Interceptor 2: Auth token injection & Session Invalidation
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          final token = prefs.getString(AppConstants.tokenKey);
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          return handler.next(options);
        },
        onError: (DioException e, handler) async {
          // Don't broadcast 401 for login/refresh endpoints to avoid loops
          if (e.response?.statusCode == 401 &&
              !e.requestOptions.path.contains('login') &&
              !e.requestOptions.path.contains('refresh-token')) {
            // Check if 401 is due to duplicate login on another device
            final isDuplicateLogin = e.response?.data is Map &&
                e.response?.data['error_code'] == 'DUPLICATE_LOGIN';
            final customMessage = isDuplicateLogin
                ? (e.response?.data['message']?.toString() ??
                    'Akun Anda telah login di perangkat lain.')
                : null;

            final refreshToken = prefs.getString(AppConstants.refreshTokenKey);
            // Skip token refresh if session was invalidated due to duplicate login
            if (!isDuplicateLogin && refreshToken != null && refreshToken.isNotEmpty) {
              try {
                // Create a fresh Dio instance to avoid interceptor recursion
                final refreshDio =
                    Dio(BaseOptions(baseUrl: AppConstants.baseUrl));
                final response = await refreshDio.post('refresh-token', data: {
                  'refresh_token': refreshToken,
                });

                if (response.statusCode == 200) {
                  final data = response.data['data'];
                  final newToken = data['token'];
                  final newRefreshToken = data['refresh_token'];

                  // Save new tokens
                  await prefs.setString(AppConstants.tokenKey, newToken);
                  await prefs.setString(
                      AppConstants.refreshTokenKey, newRefreshToken);

                  // Update original request header and retry
                  e.requestOptions.headers['Authorization'] =
                      'Bearer $newToken';

                  // Use the original dio to retry, it will use the updated header
                  final retryResponse = await dio.fetch(e.requestOptions);
                  return handler.resolve(retryResponse);
                }
              } catch (refreshError) {
                // If refresh fails, we must logout
              }
            }

            // Remove tokens and broadcast unauthorized event with optional message
            await prefs.remove(AppConstants.tokenKey);
            await prefs.remove(AppConstants.refreshTokenKey);
            _unauthorizedController.add(customMessage);
          }
          return handler.next(e);
        },
      ),
    );

    // ── Interceptor 3: Automatic retry on timeout (GET only)
    dio.interceptors.add(
      InterceptorsWrapper(
        onError: (DioException e, handler) async {
          final isTimeout = e.type == DioExceptionType.connectionTimeout ||
              e.type == DioExceptionType.receiveTimeout ||
              e.type == DioExceptionType.sendTimeout;

          final isGet = e.requestOptions.method.toUpperCase() == 'GET';

          // Read retry counter stored in extra map
          final attempt = (e.requestOptions.extra['_retryCount'] as int?) ?? 0;

          if (isTimeout && isGet && attempt < _maxRetries) {
            // Exponential backoff: 1s, 2s
            final delay = Duration(seconds: 1 << attempt); // 1, 2
            await Future.delayed(delay);

            // Increment counter
            e.requestOptions.extra['_retryCount'] = attempt + 1;

            try {
              final response = await dio.fetch(e.requestOptions);
              return handler.resolve(response);
            } on DioException catch (retryErr) {
              return handler.next(retryErr);
            }
          }

          return handler.next(e);
        },
      ),
    );
  }

  Dio get client => dio;
  String get baseUrl => dio.options.baseUrl;

  // Cleanup stream controller
  void dispose() {
    _unauthorizedController.close();
  }
}
