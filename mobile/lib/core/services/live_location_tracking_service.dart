import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_background_service/flutter_background_service.dart';
import 'package:geolocator/geolocator.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';

/// Service pelacakan lokasi real-time terintegrasi:
/// 1. **Foreground Stream Mode (UI Isolate)**: Mengalirkan GPS secara real-time
///    menggunakan [Geolocator.getPositionStream] dan periodic heartbeat (25 detik) saat aplikasi aktif di layar.
/// 2. **Background Service Mode (Background Isolate)**: Mengirim koordinat periodik
///    (interval 2.5-3 menit) saat aplikasi diminimalkan / layar mati.
class LiveLocationTrackingService with WidgetsBindingObserver {
  static const _backgroundInterval = Duration(minutes: 3);
  static final _service = FlutterBackgroundService();

  // Foreground Streamer state (UI Isolate)
  static StreamSubscription<Position>? _foregroundPositionSubscription;
  static Timer? _foregroundHeartbeatTimer;
  static DateTime? _lastSentTimestamp;
  static double? _lastSentLat;
  static double? _lastSentLng;
  static bool _isTrackingActive = false;
  static final LiveLocationTrackingService _instance = LiveLocationTrackingService._internal();

  factory LiveLocationTrackingService() => _instance;
  LiveLocationTrackingService._internal();

  static Future<void> configure() async {
    if (kIsWeb) return;

    // Pasang lifecycle observer untuk transisi otomatis Foreground <-> Background
    WidgetsBinding.instance.removeObserver(_instance);
    WidgetsBinding.instance.addObserver(_instance);

    await _service.configure(
      androidConfiguration: AndroidConfiguration(
        onStart: _onStart,
        autoStart: false,
        isForegroundMode: true,
        notificationChannelId: 'live_location_tracking',
        initialNotificationTitle: 'Pelacakan Lokasi Petugas Aktif',
        initialNotificationContent: 'Lokasi diperbarui secara real-time selama jam kerja',
        foregroundServiceNotificationId: 9101,
      ),
      iosConfiguration: IosConfiguration(
        autoStart: false,
        onForeground: _onStart,
        onBackground: _onIosBackground,
      ),
    );
  }

  /// Memulai pelacakan lokasi (Background Service + Foreground Real-time Stream).
  /// [initialLat] dan [initialLng] dapat dikirimkan langsung dari layar Absen Masuk untuk instant tracking.
  static Future<bool> start({double? initialLat, double? initialLng}) async {
    if (kIsWeb) return false;
    _isTrackingActive = true;
    try {
      final permission = await _ensurePermission();
      if (!permission) {
        debugPrint('[LiveLocationTrackingService] Izin lokasi belum diberikan.');
        return false;
      }

      // 1. Jika koordinat awal diberikan (misal dari Clock In), kirim seketika
      if (initialLat != null && initialLng != null && initialLat != 0.0 && initialLng != 0.0) {
        await _sendCoordinatesToBackend(
          latitude: initialLat,
          longitude: initialLng,
          isInstant: true,
        );
      } else {
        // Coba kirim posisi terkini / last known seketika
        sendInstantLocation();
      }

      // 2. Mulai Foreground Real-time Stream & Heartbeat (UI Isolate)
      startForegroundStream();

      // 3. Pastikan background service berjalan untuk standby saat diminimalkan
      final isRunning = await _service.isRunning();
      if (!isRunning) {
        final started = await _service.startService();
        debugPrint('[LiveLocationTrackingService] Background service running: $started');
      }

      return true;
    } catch (e) {
      debugPrint('[LiveLocationTrackingService] Error starting tracking service: $e');
      return false;
    }
  }

  /// Menghentikan seluruh proses pelacakan lokasi (saat Clock Out / Shift selesai).
  static Future<void> stop() async {
    _isTrackingActive = false;
    stopForegroundStream();
    if (kIsWeb) return;
    try {
      _service.invoke('stopTracking');
    } catch (e) {
      debugPrint('[LiveLocationTrackingService] Error stopping background service: $e');
    }
  }

  /// Mengalirkan koordinat GPS secara real-time saat aplikasi berada di latar depan (Foreground).
  static void startForegroundStream() {
    if (kIsWeb || !_isTrackingActive) return;

    // Batalkan subscription & timer sebelumnya jika ada
    stopForegroundStream();

    const locationSettings = LocationSettings(
      accuracy: LocationAccuracy.high,
      distanceFilter: 15, // Pemicu pembaruan pergeseran
    );

    _foregroundPositionSubscription = Geolocator.getPositionStream(
      locationSettings: locationSettings,
    ).listen(
      (position) {
        _handleForegroundPositionUpdate(position);
      },
      onError: (error) {
        debugPrint('[LiveLocationTrackingService] Foreground GPS stream error: $error');
      },
    );

    // Heartbeat berkala di foreground (setiap 25 detik) agar tetap live meski posisi diam
    _foregroundHeartbeatTimer = Timer.periodic(const Duration(seconds: 25), (_) {
      if (_isTrackingActive) {
        sendInstantLocation();
      }
    });

    debugPrint('[LiveLocationTrackingService] Foreground real-time stream & heartbeat aktif');
  }

  /// Menghentikan sementara streaming foreground (misal saat aplikasi diminimalkan).
  static void stopForegroundStream() {
    _foregroundPositionSubscription?.cancel();
    _foregroundPositionSubscription = null;
    _foregroundHeartbeatTimer?.cancel();
    _foregroundHeartbeatTimer = null;
  }

  /// Mengirim koordinat GPS terkini seketika (dengan fallback Last Known Position).
  static Future<void> sendInstantLocation() async {
    if (kIsWeb || !_isTrackingActive) return;
    try {
      Position? position;
      try {
        position = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.high,
          timeLimit: const Duration(seconds: 8),
        );
      } catch (timeoutErr) {
        debugPrint('[LiveLocationTrackingService] getCurrentPosition timeout/error, mencoba lastKnownPosition: $timeoutErr');
        position = await Geolocator.getLastKnownPosition();
      }

      if (position != null) {
        await _sendPositionToBackend(position, isInstant: true);
      }
    } catch (e) {
      debugPrint('[LiveLocationTrackingService] Gagal mengambil instant GPS fix: $e');
    }
  }

  /// Filter & kirim posisi GPS dari foreground stream ke backend dengan throttling.
  static Future<void> _handleForegroundPositionUpdate(Position position) async {
    final now = DateTime.now();

    // Throttling: Jangan kirim lebih cepat dari 10 detik kecuali jarak geser > 25 meter
    if (_lastSentTimestamp != null) {
      final diff = now.difference(_lastSentTimestamp!).inSeconds;
      if (diff < 10) {
        if (_lastSentLat != null && _lastSentLng != null) {
          final distance = Geolocator.distanceBetween(
            _lastSentLat!,
            _lastSentLng!,
            position.latitude,
            position.longitude,
          );
          if (distance < 25) {
            return; // Skip jika pergerakan terlalu kecil dalam < 10 detik
          }
        } else {
          return;
        }
      }
    }

    await _sendPositionToBackend(position);
  }

  static Future<void> _sendCoordinatesToBackend({
    required double latitude,
    required double longitude,
    double? accuracy,
    DateTime? recordedAt,
    bool isInstant = false,
  }) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final api = SalesApiService(apiClient: ApiClient(prefs: prefs));

      final success = await api.sendLiveLocation(
        latitude: latitude,
        longitude: longitude,
        accuracy: accuracy ?? 10.0,
        recordedAt: recordedAt ?? DateTime.now().toUtc(),
      );

      if (success) {
        _lastSentTimestamp = DateTime.now();
        _lastSentLat = latitude;
        _lastSentLng = longitude;
        debugPrint('[LiveLocationTrackingService] Lokasi terkirim (${isInstant ? "INSTANT" : "STREAM"}): $latitude, $longitude');
      } else {
        debugPrint('[LiveLocationTrackingService] Server menolak update lokasi (403/Forbidden / Not Field Staff). Menghentikan tracking service.');
        stop();
      }
    } catch (e) {
      debugPrint('[LiveLocationTrackingService] Gagal mengirim lokasi ke server: $e');
    }
  }

  static Future<void> _sendPositionToBackend(Position position, {bool isInstant = false}) async {
    await _sendCoordinatesToBackend(
      latitude: position.latitude,
      longitude: position.longitude,
      accuracy: position.accuracy,
      recordedAt: position.timestamp,
      isInstant: isInstant,
    );
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (!_isTrackingActive) return;

    if (state == AppLifecycleState.resumed) {
      debugPrint('[LiveLocationTrackingService] App RESUMED -> Memulai kembali Foreground Real-time Stream');
      startForegroundStream();
      sendInstantLocation();
    } else if (state == AppLifecycleState.paused || state == AppLifecycleState.detached) {
      debugPrint('[LiveLocationTrackingService] App PAUSED/MINIMIZED -> Pause Foreground Stream, background worker aktif');
      stopForegroundStream();
    }
  }

  static Future<bool> _ensurePermission() async {
    if (!await Geolocator.isLocationServiceEnabled()) return false;
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    return permission == LocationPermission.always ||
        permission == LocationPermission.whileInUse;
  }

  @pragma('vm:entry-point')
  static Future<bool> _onIosBackground(ServiceInstance service) async {
    return true;
  }

  @pragma('vm:entry-point')
  static void _onStart(ServiceInstance service) async {
    if (service is AndroidServiceInstance) {
      service.setAsForegroundService();
      service.setForegroundNotificationInfo(
        title: 'Pelacakan Lokasi Petugas Aktif',
        content: 'Lokasi diperbarui secara real-time selama jam kerja',
      );
    }

    final prefs = await SharedPreferences.getInstance();
    final api = SalesApiService(apiClient: ApiClient(prefs: prefs));
    Timer? timer;

    Future<void> send() async {
      try {
        final position = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.medium,
          timeLimit: const Duration(seconds: 15),
        );
        final success = await api.sendLiveLocation(
          latitude: position.latitude,
          longitude: position.longitude,
          accuracy: position.accuracy,
          recordedAt: position.timestamp,
        );
        if (!success) {
          debugPrint('[LiveLocationTrackingService] Background send returned false (403/Forbidden). Stopping service.');
          timer?.cancel();
          service.stopSelf();
        }
      } catch (e) {
        debugPrint('[LiveLocationTrackingService] Background position tick error: $e');
      }
    }

    await send();
    timer = Timer.periodic(_backgroundInterval, (_) => send());
    service.on('stopTracking').listen((_) {
      timer?.cancel();
      service.stopSelf();
    });
  }
}
