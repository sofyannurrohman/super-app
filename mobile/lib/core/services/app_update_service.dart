import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:in_app_update/in_app_update.dart';

/// Service untuk mengelola Google Play In-App Updates API.
///
/// Mendukung dua mode pembaruan:
/// - **Immediate Update**: Update paksa full-screen, cocok untuk breaking changes.
///   Digunakan ketika [AppUpdateInfo.updatePriority] >= 4 (skala 0-5 dari Play Console).
/// - **Flexible Update**: Update background, cocok untuk update minor.
///   Digunakan ketika [AppUpdateInfo.updatePriority] < 4.
///
/// Prioritas update diatur melalui Google Play Console saat merilis versi baru.
class AppUpdateService {
  AppUpdateService._();

  static final AppUpdateService _instance = AppUpdateService._();
  static AppUpdateService get instance => _instance;

  // Threshold prioritas untuk memutuskan mode Immediate vs Flexible.
  // Nilai 0-5 diatur di Play Console. Priority >= _immediateThreshold -> Immediate.
  static const int _immediateThreshold = 4;

  /// Memeriksa ketersediaan update dari Play Store dan menjalankan alur
  /// yang sesuai (Immediate atau Flexible).
  ///
  /// Hanya berjalan di Android release mode; tidak aktif di iOS atau debug mode.
  Future<void> checkAndUpdate(BuildContext context) async {
    // In-App Update API hanya tersedia di Android
    if (kIsWeb || !Platform.isAndroid) return;

    // Jangan tampilkan di debug mode karena membutuhkan Play Store aktif
    if (kDebugMode) {
      debugPrint('[AppUpdateService] Skipped: debug mode.');
      return;
    }

    try {
      final AppUpdateInfo info = await InAppUpdate.checkForUpdate();

      if (info.updateAvailability != UpdateAvailability.updateAvailable) {
        debugPrint('[AppUpdateService] Tidak ada update tersedia.');
        return;
      }

      final int priority = info.updatePriority ?? 0;
      debugPrint('[AppUpdateService] Update tersedia. Priority: $priority');

      if (priority >= _immediateThreshold) {
        await _startImmediateUpdate();
      } else {
        if (context.mounted) {
          await _startFlexibleUpdate(context);
        }
      }
    } catch (e) {
      // Tidak memblokir aplikasi jika terjadi error pada pengecekan update
      debugPrint('[AppUpdateService] Error saat mengecek update: $e');
    }
  }

  // ---------------------------------------------------------------------------
  // Immediate Update (Update Paksa)
  // ---------------------------------------------------------------------------

  /// Memulai alur Immediate Update.
  ///
  /// Layar aplikasi akan sepenuhnya diblokir oleh UI Play Store.
  /// Pengguna **wajib** menyelesaikan update sebelum dapat kembali menggunakan
  /// aplikasi. Jika update dibatalkan (misal ditekan tombol back), aplikasi
  /// akan ditutup paksa.
  Future<void> _startImmediateUpdate() async {
    try {
      debugPrint('[AppUpdateService] Memulai Immediate Update...');
      final AppUpdateResult result = await InAppUpdate.performImmediateUpdate();

      if (result == AppUpdateResult.success) {
        debugPrint('[AppUpdateService] Immediate Update berhasil.');
      } else if (result == AppUpdateResult.userDeniedUpdate) {
        // Jika user menolak/membatalkan immediate update, tutup aplikasi
        // karena update bersifat kritis dan tidak bisa dilewati.
        debugPrint(
            '[AppUpdateService] User membatalkan Immediate Update. Menutup aplikasi.');
        exit(0);
      }
    } catch (e) {
      debugPrint('[AppUpdateService] Error pada Immediate Update: $e');
    }
  }

  // ---------------------------------------------------------------------------
  // Flexible Update (Update Latar Belakang)
  // ---------------------------------------------------------------------------

  /// Memulai alur Flexible Update.
  ///
  /// Menampilkan dialog penawaran update. Jika disetujui, download berjalan
  /// di background. Setelah download selesai, snackbar akan muncul dengan
  /// tombol "RESTART" untuk menginstal pembaruan.
  Future<void> _startFlexibleUpdate(BuildContext context) async {
    try {
      debugPrint('[AppUpdateService] Memulai Flexible Update...');
      final AppUpdateResult result = await InAppUpdate.startFlexibleUpdate();

      if (result == AppUpdateResult.success) {
        debugPrint(
            '[AppUpdateService] Flexible Update berhasil didownload. Menunggu install...');
        // Setelah download selesai, tampilkan snackbar untuk restart
        if (context.mounted) {
          _showInstallSnackbar(context);
        }
      } else if (result == AppUpdateResult.userDeniedUpdate) {
        debugPrint('[AppUpdateService] User menolak Flexible Update.');
      }
    } catch (e) {
      debugPrint('[AppUpdateService] Error pada Flexible Update: $e');
    }
  }

  /// Menampilkan snackbar setelah download Flexible Update selesai,
  /// memberi tahu pengguna untuk restart aplikasi agar update terinstal.
  void _showInstallSnackbar(BuildContext context) {
    final ScaffoldMessengerState? messenger =
        ScaffoldMessenger.maybeOf(context);
    if (messenger == null) return;

    messenger.showSnackBar(
      SnackBar(
        backgroundColor: const Color(0xFF1E293B),
        behavior: SnackBarBehavior.floating,
        margin: const EdgeInsets.all(16),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 10),
        content: const Row(
          children: [
            Icon(Icons.system_update_alt_rounded,
                color: Color(0xFF34D399), size: 20),
            SizedBox(width: 12),
            Expanded(
              child: Text(
                'Pembaruan siap diinstal!',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ],
        ),
        action: SnackBarAction(
          label: 'RESTART',
          textColor: const Color(0xFF34D399),
          onPressed: () async {
            try {
              await InAppUpdate.completeFlexibleUpdate();
            } catch (e) {
              debugPrint(
                  '[AppUpdateService] Error saat complete flexible update: $e');
            }
          },
        ),
      ),
    );
  }
}
