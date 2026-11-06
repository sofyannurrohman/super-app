import 'dart:io';
import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:timezone/timezone.dart' as tz;
import 'package:timezone/data/latest_all.dart' as tz_data;
import 'package:flutter_timezone/flutter_timezone.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_event.dart';
import 'package:hris_app/features/announcement/presentation/bloc/announcement_bloc.dart'
    as hris_announcement;
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';

// ─── Keyword set untuk mendeteksi notifikasi reminder absen ───────────────────
// Jika title atau body mengandung salah satu keyword ini, notifikasi
// akan ditampilkan dengan channel "shift_alerts" (suara alarm).
const _kAttendanceKeywords = [
  'presensi',
  'absen',
  'absensi',
  'clock',
  'masuk',
  'keluar',
  'shift',
  'reminder',
  'pengingat',
  'waktu kerja',
];

const _kWarningKeywords = [
  'sp1',
  'sp 1',
  'sp2',
  'sp 2',
  'sp3',
  'sp 3',
  'teguran',
  'peringatan',
  'pelanggaran',
  'disciplinary',
  'tindak lanjut',
  'tindak lanjut cs',
  'tindak lanjut spv',
  'follow up',
  'followup',
];

bool _isAttendanceReminder(String title, String body) {
  final combined = '${title.toLowerCase()} ${body.toLowerCase()}';
  return _kAttendanceKeywords.any((kw) => combined.contains(kw));
}

bool _isWarningNotification(String title, String body) {
  final combined = '${title.toLowerCase()} ${body.toLowerCase()}';
  return _kWarningKeywords.any((kw) => combined.contains(kw));
}

@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  try {
    await Firebase.initializeApp();
    debugPrint("Handling a background message: ${message.messageId}");

    final title =
        message.notification?.title ?? message.data['title'] ?? 'Pengingat';
    final body = message.notification?.body ?? message.data['body'];
    final type = message.data['type'];

    if (type == 'FORCE_LOGOUT') {
      try {
        final prefs = await SharedPreferences.getInstance();
        await prefs.remove(AppConstants.tokenKey);
        await prefs.remove(AppConstants.refreshTokenKey);
      } catch (_) {}
    }

    if (body != null) {
      final FlutterLocalNotificationsPlugin plugin =
          FlutterLocalNotificationsPlugin();
      const AndroidInitializationSettings initAndroid =
          AndroidInitializationSettings('@mipmap/ic_launcher');
      const DarwinInitializationSettings initDarwin =
          DarwinInitializationSettings();
      const InitializationSettings initSettings = InitializationSettings(
        android: initAndroid,
        iOS: initDarwin,
      );
      await plugin.initialize(settings: initSettings);

      // Pilih channel berdasarkan tipe notifikasi
      final isLeave = type == 'leave';
      final isAbsen = !isLeave && _isAttendanceReminder(title, body);
      final isWarning = !isLeave && _isWarningNotification(title, body);

      final androidDetails = (isAbsen || isWarning)
          ? AndroidNotificationDetails(
              isWarning ? 'hr_warning_alerts' : 'shift_alerts_alarm_v2',
              isWarning ? 'Peringatan HRD' : 'Pengingat Jadwal Kerja',
              channelDescription: isWarning
                  ? 'Peringatan pelanggaran dan SP'
                  : 'Notifikasi berdering untuk presensi tepat waktu',
              importance: Importance.max,
              priority: Priority.high,
              showWhen: true,
              enableVibration: true,
              playSound: true,
              sound: const RawResourceAndroidNotificationSound('alarm'),
            )
          : const AndroidNotificationDetails(
              'fcm_background_alerts',
              'Pemberitahuan Latar Belakang',
              channelDescription: 'Notifikasi saat aplikasi di latar belakang',
              importance: Importance.max,
              priority: Priority.high,
              showWhen: true,
            );

      final iosDetails = isAbsen
          ? const DarwinNotificationDetails(
              presentAlert: true,
              presentBadge: true,
              presentSound: true,
              sound: 'alarm.mp3',
            )
          : const DarwinNotificationDetails();

      final platformDetails = NotificationDetails(
        android: androidDetails,
        iOS: iosDetails,
      );

      await plugin.show(
        id: message.messageId.hashCode,
        title: title,
        body: body,
        notificationDetails: platformDetails,
      );
    }
  } catch (e) {
    debugPrint("Error in background messaging handler: $e");
  }
}

class NotificationService {
  static final NotificationService _instance = NotificationService._internal();
  factory NotificationService() => _instance;
  NotificationService._internal();

  final FlutterLocalNotificationsPlugin _notificationsPlugin =
      FlutterLocalNotificationsPlugin();

  Future<void> init() async {
    if (kIsWeb) return;

    // ── Firebase Core & Messaging ─────────────────────────────────────────────
    try {
      await Firebase.initializeApp();
      FirebaseMessaging.onBackgroundMessage(
        _firebaseMessagingBackgroundHandler,
      );

      // Tangani pesan saat app di foreground
      FirebaseMessaging.onMessage.listen((RemoteMessage message) {
        final notification = message.notification;
        final title =
            notification?.title ?? message.data['title'] ?? 'Notifikasi Baru';
        final body = notification?.body ?? message.data['body'];

        final type = message.data['type'];

        if (type == 'FORCE_LOGOUT') {
          try {
            final sharedPreferences = di.sl<SharedPreferences>();
            unawaited(sharedPreferences.remove(AppConstants.tokenKey));
            unawaited(sharedPreferences.remove(AppConstants.refreshTokenKey));
            di.sl<AuthBloc>().add(
              SessionExpired(
                message: body ?? 'Akun Anda telah login di perangkat lain.',
              ),
            );
          } catch (_) {}
          showAdHocNotification(
            title: title,
            body: body ?? 'Akun Anda telah login di perangkat lain.',
          );
          return;
        }

        if (type == 'REMOTE_SYNC' || type == 'FORCE_SYNC') {
          try {
            debugPrint(
              '[NotificationService] REMOTE_SYNC diterima dari server. Memicu syncAll.',
            );
            di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
          } catch (e) {
            debugPrint(
              '[NotificationService] Error triggering remote sync: $e',
            );
          }
          return;
        }

        if (type == 'popup') {
          // Jika aplikasi sedang di depan dan menerima popup, trigger polling ulang
          try {
            di.sl<hris_announcement.AnnouncementBloc>().add(
              hris_announcement.GetPendingPopupsEvent(),
            );
          } catch (_) {}
        } else if (type == 'warning' ||
            _isWarningNotification(title, body ?? '')) {
          _showWarningAlertNotification(
            title: title,
            body: body ?? '',
            id: message.messageId.hashCode,
          );
          try {
            // Hapus cache acknowledged warning agar popup dialog segera muncul jika aplikasi di foreground
            SharedPreferences.getInstance().then((p) {
              p.remove('acknowledged_warning_id');
            });
            di.sl<AuthBloc>().add(CheckAuthStatusRequested());
          } catch (_) {}
        } else if (type == 'leave') {
          if (body != null) {
            showAdHocNotification(
              title: title,
              body: body,
              id: message.messageId.hashCode,
            );
          }
        } else if (body != null) {
          // Gunakan alarm sound jika ini adalah reminder absen
          if (_isAttendanceReminder(title, body)) {
            _showAttendanceReminderNotification(
              title: title,
              body: body,
              id: message.messageId.hashCode,
            );
          } else {
            showAdHocNotification(title: title, body: body);
          }
        }
      });

      // ── Listener: FCM Token Refresh ─────────────────────────────────────────
      // Dipanggil otomatis jika Google FCM merotasi token (misal setelah
      // uninstall-reinstall, atau token expired). Pastikan backend selalu
      // mendapat token terbaru agar reminder absen bisa dikirim.
      FirebaseMessaging.instance.onTokenRefresh.listen((newToken) async {
        debugPrint('FCM Token refreshed: $newToken');
        // Sync token baru ke backend tanpa menunggu hasilnya
        unawaited(_syncFcmToken(newToken));
      });

      // Sync token saat startup (jika user sudah login)
      unawaited(syncTokenWithBackend());
    } catch (e) {
      debugPrint('Firebase/FCM Initialization failed: $e');
    }

    // ── Timezone ──────────────────────────────────────────────────────────────
    tz_data.initializeTimeZones();
    try {
      final String timeZoneName =
          (await FlutterTimezone.getLocalTimezone()).identifier;
      tz.setLocalLocation(tz.getLocation(timeZoneName));
    } catch (e) {
      try {
        tz.setLocalLocation(tz.getLocation('Asia/Jakarta'));
      } catch (_) {
        tz.setLocalLocation(tz.UTC);
      }
    }

    // ── Local Notifications ───────────────────────────────────────────────────
    const AndroidInitializationSettings initAndroid =
        AndroidInitializationSettings('@mipmap/ic_launcher');

    const DarwinInitializationSettings initDarwin =
        DarwinInitializationSettings(
          requestAlertPermission: true,
          requestBadgePermission: true,
          requestSoundPermission: true,
        );

    const InitializationSettings initSettings = InitializationSettings(
      android: initAndroid,
      iOS: initDarwin,
    );

    await _notificationsPlugin.initialize(
      settings: initSettings,
      onDidReceiveNotificationResponse: (NotificationResponse details) {
        // Handle notification tap jika diperlukan di masa depan
        debugPrint('Notification tapped: ${details.payload}');
      },
    );

    // Daftarkan Android Notification Channels secara native
    final androidPlugin = _notificationsPlugin
        .resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin
        >();
    if (androidPlugin != null) {
      const shiftChannel = AndroidNotificationChannel(
        'shift_alerts_alarm_v2',
        'Pengingat Jadwal Kerja',
        description: 'Notifikasi berdering untuk presensi tepat waktu',
        importance: Importance.max,
        playSound: true,
        sound: RawResourceAndroidNotificationSound('alarm'),
        enableVibration: true,
      );
      await androidPlugin.createNotificationChannel(shiftChannel);

      const warningChannel = AndroidNotificationChannel(
        'hr_warning_alerts',
        'Peringatan HRD',
        description: 'Peringatan pelanggaran dan SP',
        importance: Importance.max,
        playSound: true,
        sound: RawResourceAndroidNotificationSound('alarm'),
        enableVibration: true,
      );
      await androidPlugin.createNotificationChannel(warningChannel);

      const fcmChannel = AndroidNotificationChannel(
        'fcm_background_alerts',
        'Pemberitahuan Latar Belakang',
        description: 'Notifikasi saat aplikasi di latar belakang',
        importance: Importance.max,
        playSound: true,
        enableVibration: true,
      );
      await androidPlugin.createNotificationChannel(fcmChannel);

      const liveLocationChannel = AndroidNotificationChannel(
        'live_location_tracking',
        'Pelacakan Lokasi Kerja',
        description: 'Notifikasi pelacakan lokasi saat jam kerja aktif',
        importance: Importance.low,
      );
      await androidPlugin.createNotificationChannel(liveLocationChannel);
    }
  }

  Future<void> requestPermissions() async {
    if (kIsWeb) return;

    try {
      await FirebaseMessaging.instance.requestPermission(
        alert: true,
        badge: true,
        sound: true,
      );
    } catch (_) {}

    await _notificationsPlugin
        .resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin
        >()
        ?.requestNotificationsPermission();

    await _notificationsPlugin
        .resolvePlatformSpecificImplementation<
          IOSFlutterLocalNotificationsPlugin
        >()
        ?.requestPermissions(alert: true, badge: true, sound: true);
  }

  Future<void> showTestNotification() async {
    if (kIsWeb) return;
    const AndroidNotificationDetails androidDetails =
        AndroidNotificationDetails(
          'test_channel',
          'Test Notifications',
          channelDescription: 'Used for testing notifications',
          importance: Importance.max,
          priority: Priority.high,
          showWhen: true,
          enableVibration: true,
          playSound: true,
        );

    const NotificationDetails platformDetails = NotificationDetails(
      android: androidDetails,
      iOS: DarwinNotificationDetails(
        presentAlert: true,
        presentBadge: true,
        presentSound: true,
      ),
    );

    await _notificationsPlugin.show(
      id: 999,
      title: 'Uji Coba Notifikasi',
      body: 'Ini adalah notifikasi uji coba dari aplikasi Karsa HR Super App.',
      notificationDetails: platformDetails,
    );
  }

  // ── Notifikasi Absen (dengan alarm sound) ─────────────────────────────────
  /// Tampilkan notifikasi foreground khusus untuk reminder absen masuk/keluar.
  /// Menggunakan channel [shift_alerts_alarm_v2] yang memiliki suara alarm.
  Future<void> _showAttendanceReminderNotification({
    required String title,
    required String body,
    int id = 777,
  }) async {
    if (kIsWeb) return;
    const AndroidNotificationDetails androidDetails =
        AndroidNotificationDetails(
          'shift_alerts_alarm_v2',
          'Pengingat Jadwal Kerja',
          channelDescription: 'Notifikasi berdering untuk presensi tepat waktu',
          importance: Importance.max,
          priority: Priority.high,
          showWhen: true,
          enableVibration: true,
          playSound: true,
          sound: RawResourceAndroidNotificationSound('alarm'),
        );

    const NotificationDetails platformDetails = NotificationDetails(
      android: androidDetails,
      iOS: DarwinNotificationDetails(
        presentAlert: true,
        presentBadge: true,
        presentSound: true,
        sound: 'alarm.mp3',
      ),
    );

    await _notificationsPlugin.show(
      id: id,
      title: title,
      body: body,
      notificationDetails: platformDetails,
    );
  }

  // ── Notifikasi Peringatan HRD / SP / Tindak Lanjut (dengan alarm sound) ──
  Future<void> _showWarningAlertNotification({
    required String title,
    required String body,
    int id = 999,
  }) async {
    if (kIsWeb) return;
    const AndroidNotificationDetails androidDetails =
        AndroidNotificationDetails(
          'hr_warning_alerts',
          'Peringatan HRD',
          channelDescription: 'Peringatan pelanggaran dan SP',
          importance: Importance.max,
          priority: Priority.high,
          showWhen: true,
          enableVibration: true,
          playSound: true,
          sound: RawResourceAndroidNotificationSound('alarm'),
        );

    const NotificationDetails platformDetails = NotificationDetails(
      android: androidDetails,
      iOS: DarwinNotificationDetails(
        presentAlert: true,
        presentBadge: true,
        presentSound: true,
        sound: 'alarm.mp3',
      ),
    );

    await _notificationsPlugin.show(
      id: id,
      title: title,
      body: body,
      notificationDetails: platformDetails,
    );
  }

  // ── Penjadwalan Notifikasi Lokal ─────────────────────────────────────────
  /// Jadwalkan notifikasi lokal (offline, tanpa server) untuk reminder absen.
  /// Berguna sebagai fallback jika FCM tidak terkirim — misalnya user tidak
  /// terkoneksi internet saat jam absen tiba.
  Future<void> scheduleAttendanceReminder({
    required int id,
    required String title,
    required String body,
    required DateTime scheduledDate,
  }) async {
    if (kIsWeb) return;

    bool canScheduleExact = true;
    try {
      if (Platform.isAndroid) {
        final androidImpl = _notificationsPlugin
            .resolvePlatformSpecificImplementation<
              AndroidFlutterLocalNotificationsPlugin
            >();
        if (androidImpl != null) {
          canScheduleExact =
              await androidImpl.canScheduleExactNotifications() ?? false;
        }
      }
    } catch (_) {
      canScheduleExact = false;
    }

    final scheduleMode = canScheduleExact
        ? AndroidScheduleMode.exactAllowWhileIdle
        : AndroidScheduleMode.inexactAllowWhileIdle;

    await _notificationsPlugin.zonedSchedule(
      id: id,
      title: title,
      body: body,
      scheduledDate: tz.TZDateTime.from(scheduledDate, tz.local),
      notificationDetails: const NotificationDetails(
        android: AndroidNotificationDetails(
          'shift_alerts_alarm_v2',
          'Pengingat Jadwal Kerja',
          channelDescription: 'Notifikasi berdering untuk presensi tepat waktu',
          importance: Importance.max,
          priority: Priority.high,
          showWhen: true,
          enableVibration: true,
          playSound: true,
          sound: RawResourceAndroidNotificationSound('alarm'),
        ),
        iOS: DarwinNotificationDetails(
          presentAlert: true,
          presentBadge: true,
          presentSound: true,
          sound: 'alarm.mp3',
        ),
      ),
      androidScheduleMode: scheduleMode,
    );
  }

  // ── Notifikasi Ad-hoc (pengumuman, approval, dll) ─────────────────────────
  Future<void> showAdHocNotification({
    required String title,
    required String body,
    int id = 888,
  }) async {
    if (kIsWeb) return;
    const AndroidNotificationDetails androidDetails =
        AndroidNotificationDetails(
          'fcm_alerts_channel_v1',
          'Informasi & Pengumuman',
          channelDescription: 'Notifikasi real-time dari kantor',
          importance: Importance.max,
          priority: Priority.high,
          showWhen: true,
          enableVibration: true,
          playSound: true,
        );

    const NotificationDetails platformDetails = NotificationDetails(
      android: androidDetails,
      iOS: DarwinNotificationDetails(
        presentAlert: true,
        presentBadge: true,
        presentSound: true,
      ),
    );

    await _notificationsPlugin.show(
      id: id,
      title: title,
      body: body,
      notificationDetails: platformDetails,
    );
  }

  // ── FCM Token Sync ────────────────────────────────────────────────────────
  /// Sync token FCM ke backend. Dipanggil setelah login, saat startup (jika
  /// sudah login), dan setiap kali token di-refresh oleh Firebase.
  Future<void> syncTokenWithBackend() async {
    if (kIsWeb) return;
    try {
      final token = await FirebaseMessaging.instance.getToken();
      if (token == null) {
        debugPrint('FCM Token is null, skip sync');
        return;
      }
      await _syncFcmToken(token);
    } catch (e) {
      debugPrint('Error retrieving FCM Token: $e');
    }
  }

  /// Internal helper: kirim token ke endpoint backend dan subscribe topic.
  Future<void> _syncFcmToken(String token) async {
    if (kIsWeb) return;
    try {
      debugPrint('Syncing FCM Token: $token');

      // Subscribe ke topik-topik reminder absen
      await FirebaseMessaging.instance.subscribeToTopic('reminder_absen');
      await FirebaseMessaging.instance.subscribeToTopic('attendance_reminder');
      await FirebaseMessaging.instance.subscribeToTopic('all_users');

      // Resolve dependencies secara dinamis (menghindari circular DI)
      final sharedPreferences = di.sl<SharedPreferences>();
      final apiClient = di.sl<ApiClient>();

      final String? jwtToken = sharedPreferences.getString(
        AppConstants.tokenKey,
      );
      if (jwtToken != null && jwtToken.isNotEmpty) {
        final response = await apiClient.client.post(
          'users/fcm-token',
          data: {
            'fcm_token': token,
            'device_info': Platform.isAndroid ? 'Android' : 'iOS',
          },
        );
        if (response.statusCode == 200) {
          debugPrint('FCM Token successfully synced with backend.');
        } else {
          debugPrint('Failed to sync FCM Token: ${response.statusCode}');
        }
      } else {
        debugPrint('Skipping FCM Token sync: User not authenticated.');
      }
    } catch (e) {
      debugPrint('Error syncing FCM Token: $e');
    }
  }
}
