import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:path_provider/path_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/database/database.dart';

/// Severity level untuk setiap entri audit log.
enum AuditSeverity { info, warning, error }

/// Model untuk satu entri di audit trail.
class AuditEntry {
  final String id;
  final DateTime timestamp;
  final String category;
  final String action;
  final String detail;
  final AuditSeverity severity;

  AuditEntry({
    required this.id,
    required this.timestamp,
    required this.category,
    required this.action,
    required this.detail,
    this.severity = AuditSeverity.info,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'timestamp': timestamp.toIso8601String(),
        'category': category,
        'action': action,
        'detail': detail,
        'severity': severity.name,
      };

  factory AuditEntry.fromJson(Map<String, dynamic> json) => AuditEntry(
        id: json['id'] ?? '',
        timestamp: DateTime.tryParse(json['timestamp'] ?? '') ?? DateTime.now(),
        category: json['category'] ?? '',
        action: json['action'] ?? '',
        detail: json['detail'] ?? '',
        severity: AuditSeverity.values.firstWhere(
          (e) => e.name == json['severity'],
          orElse: () => AuditSeverity.info,
        ),
      );
}

/// Service utama untuk penjaminan data offline dari data loss.
///
/// Tanggung jawab:
/// 1. Menyimpan audit trail semua operasi tulis data kritis ke file persisten.
/// 2. Menyediakan mekanisme retry dengan exponential backoff.
/// 3. Mendeteksi dan melaporkan transaksi yang terlalu lama belum tersinkron.
/// 4. Menyimpan antrian offline delivery di direktori aplikasi (bukan temp).
class DataIntegrityService {
  static final DataIntegrityService _instance =
      DataIntegrityService._internal();
  factory DataIntegrityService() => _instance;
  DataIntegrityService._internal();

  // Batas maksimum retry untuk satu item sebelum dianggap "stuck"
  static const int maxRetryCount = 8;

  // Durasi basis untuk exponential backoff (dalam detik):
  // Attempt 1: 30s, 2: 60s, 3: 120s, 4: 240s, 5: 480s, 6: 960s, 7: 1920s, 8: 3600s
  static const int baseBackoffSeconds = 30;

  // Transaksi yang belum sync lebih dari threshold ini dianggap berisiko
  static const Duration stalePendingThreshold = Duration(hours: 4);

  // Audit log dibatasi 500 entri untuk efisiensi storage
  static const int maxAuditLogEntries = 500;

  final List<AuditEntry> _auditBuffer = [];

  // ─── AUDIT TRAIL ──────────────────────────────────────────────────────────

  Future<File?> _getAuditFile() async {
    if (kIsWeb) return null;
    try {
      final dir = await getApplicationDocumentsDirectory();
      return File('${dir.path}/data_audit_trail.jsonl');
    } catch (e) {
      debugPrint('[DataIntegrity] Gagal mendapatkan audit file: $e');
      return null;
    }
  }

  /// Catat satu entri audit. Thread-safe via buffer flush.
  Future<void> log({
    required String category,
    required String action,
    required String detail,
    AuditSeverity severity = AuditSeverity.info,
  }) async {
    final entry = AuditEntry(
      id: DateTime.now().millisecondsSinceEpoch.toRadixString(16),
      timestamp: DateTime.now(),
      category: category,
      action: action,
      detail: detail,
      severity: severity,
    );

    _auditBuffer.add(entry);
    debugPrint(
        '[AUDIT][${severity.name.toUpperCase()}] [$category] $action: $detail');

    // Flush ke disk setiap 5 entri atau langsung untuk error
    if (_auditBuffer.length >= 5 || severity == AuditSeverity.error) {
      await _flushAuditBuffer();
    }
  }

  Future<void> _flushAuditBuffer() async {
    if (_auditBuffer.isEmpty) return;
    if (kIsWeb) {
      _auditBuffer.clear();
      return;
    }

    try {
      final file = await _getAuditFile();
      if (file == null) return;

      final lines = _auditBuffer.map((e) => jsonEncode(e.toJson())).join('\n');
      await file.writeAsString('$lines\n', mode: FileMode.append, flush: true);
      _auditBuffer.clear();

      // Trim log jika terlalu besar
      await _trimAuditLog(file);
    } catch (e) {
      debugPrint('[DataIntegrity] Gagal flush audit buffer: $e');
    }
  }

  Future<void> _trimAuditLog(File file) async {
    try {
      final lines = await file.readAsLines();
      if (lines.length > maxAuditLogEntries) {
        final trimmed =
            lines.skip(lines.length - maxAuditLogEntries).join('\n');
        await file.writeAsString('$trimmed\n', flush: true);
      }
    } catch (_) {}
  }

  /// Baca semua entri audit trail (untuk debug/monitoring).
  Future<List<AuditEntry>> readAuditLog() async {
    if (kIsWeb) return [];
    try {
      final file = await _getAuditFile();
      if (file == null || !await file.exists()) return [];
      final lines = await file.readAsLines();
      return lines
          .where((l) => l.trim().isNotEmpty)
          .map((l) {
            try {
              return AuditEntry.fromJson(jsonDecode(l));
            } catch (_) {
              return null;
            }
          })
          .whereType<AuditEntry>()
          .toList();
    } catch (e) {
      debugPrint('[DataIntegrity] Gagal baca audit log: $e');
      return [];
    }
  }

  // ─── RETRY METADATA ───────────────────────────────────────────────────────

  static const String _webRetryMetaKey = 'sync_retry_meta_json';

  Future<File?> _getRetryMetaFile() async {
    if (kIsWeb) return null;
    try {
      final dir = await getApplicationDocumentsDirectory();
      return File('${dir.path}/sync_retry_meta.json');
    } catch (_) {
      return null;
    }
  }

  Future<Map<String, dynamic>> _loadRetryMeta() async {
    try {
      if (kIsWeb) {
        final prefs = await SharedPreferences.getInstance();
        final raw = prefs.getString(_webRetryMetaKey);
        if (raw == null || raw.isEmpty) return {};
        return Map<String, dynamic>.from(jsonDecode(raw));
      }
      final file = await _getRetryMetaFile();
      if (file == null || !await file.exists()) return {};
      final content = await file.readAsString();
      return Map<String, dynamic>.from(jsonDecode(content));
    } catch (_) {
      return {};
    }
  }

  Future<void> _saveRetryMeta(Map<String, dynamic> meta) async {
    try {
      if (kIsWeb) {
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString(_webRetryMetaKey, jsonEncode(meta));
        return;
      }
      final file = await _getRetryMetaFile();
      if (file == null) return;
      await file.writeAsString(jsonEncode(meta), flush: true);
    } catch (e) {
      debugPrint('[DataIntegrity] Gagal simpan retry meta: $e');
    }
  }

  /// Catat bahwa satu item (transaksi/checkin) baru saja gagal sync.
  /// Menambah counter retry dan mencatat waktu retry berikutnya beserta alasannya.
  Future<void> recordSyncFailure(String localId, {String? reason}) async {
    final meta = await _loadRetryMeta();
    final current = meta[localId] as Map<String, dynamic>? ?? {};
    final retryCount = (current['retryCount'] as int? ?? 0) + 1;

    // Exponential backoff: baseBackoffSeconds * 2^(retryCount-1)
    final backoffSeconds =
        baseBackoffSeconds * (1 << (retryCount - 1).clamp(0, 6));
    final nextRetryAt = DateTime.now().add(Duration(seconds: backoffSeconds));

    meta[localId] = {
      'retryCount': retryCount,
      'lastFailedAt': DateTime.now().toIso8601String(),
      'nextRetryAt': nextRetryAt.toIso8601String(),
      if (reason != null && reason.isNotEmpty) 'lastError': reason,
    };

    await _saveRetryMeta(meta);
    await log(
      category: 'SYNC',
      action: 'RETRY_SCHEDULED',
      detail:
          'localId=$localId retryCount=$retryCount reason=${reason ?? 'unknown'} nextRetryAt=${nextRetryAt.toIso8601String()}',
      severity: retryCount >= maxRetryCount
          ? AuditSeverity.error
          : AuditSeverity.warning,
    );
  }

  /// Dapatkan informasi retry untuk satu transaksi lokal (misal untuk ditampilkan di UI).
  Future<Map<String, dynamic>?> getRetryInfo(String localId) async {
    final meta = await _loadRetryMeta();
    if (meta.containsKey(localId)) {
      return Map<String, dynamic>.from(meta[localId] as Map);
    }
    return null;
  }

  /// Dapatkan seluruh map retry meta.
  Future<Map<String, dynamic>> getAllRetryMeta() async {
    return await _loadRetryMeta();
  }

  /// Catat bahwa satu item berhasil disinkronisasikan. Hapus metadata retry-nya.
  Future<void> recordSyncSuccess(String localId) async {
    final meta = await _loadRetryMeta();
    if (meta.containsKey(localId)) {
      meta.remove(localId);
      await _saveRetryMeta(meta);
    }
    await log(
      category: 'SYNC',
      action: 'SYNC_SUCCESS',
      detail: 'localId=$localId berhasil disinkronisasikan ke server.',
    );
  }

  /// Reset metadata retry untuk satu item spesifik (misalnya saat user meminta force sync).
  Future<void> resetRetry(String localId) async {
    final meta = await _loadRetryMeta();
    if (meta.containsKey(localId)) {
      meta.remove(localId);
      await _saveRetryMeta(meta);
    }
  }

  /// Reset semua metadata retry untuk seluruh transaksi pending (misal saat manual refresh).
  Future<void> resetAllPendingRetries() async {
    try {
      if (kIsWeb) {
        final prefs = await SharedPreferences.getInstance();
        await prefs.remove(_webRetryMetaKey);
        return;
      }
      final file = await _getRetryMetaFile();
      if (file != null && await file.exists()) {
        await file.delete();
      }
    } catch (e) {
      debugPrint('[DataIntegrity] resetAllPendingRetries error: $e');
    }
  }

  /// Periksa apakah sebuah item sudah boleh dicoba sync ulang.
  /// Mengembalikan true jika belum pernah gagal (fresh item) atau sudah melewati
  /// waktu backoff-nya.
  Future<bool> isReadyForRetry(String localId, {bool force = false}) async {
    if (force) {
      await resetRetry(localId);
      return true;
    }

    final meta = await _loadRetryMeta();
    if (!meta.containsKey(localId))
      return true; // fresh item, langsung boleh sync

    final itemMeta = meta[localId] as Map<String, dynamic>;
    final retryCount = itemMeta['retryCount'] as int? ?? 0;

    if (retryCount >= maxRetryCount) {
      // Sudah mencapai batas maksimum retry, tidak perlu dicoba lagi secara otomatis
      return false;
    }

    final nextRetryStr = itemMeta['nextRetryAt'] as String?;
    if (nextRetryStr == null) return true;

    final nextRetryAt = DateTime.tryParse(nextRetryStr);
    if (nextRetryAt == null) return true;

    return DateTime.now().isAfter(nextRetryAt);
  }

  /// Dapatkan daftar localId yang sudah stuck (melebihi maxRetryCount).
  Future<List<String>> getStuckItems() async {
    final meta = await _loadRetryMeta();
    return meta.entries
        .where((e) {
          final retryCount =
              (e.value as Map<String, dynamic>)['retryCount'] as int? ?? 0;
          return retryCount >= maxRetryCount;
        })
        .map((e) => e.key)
        .toList();
  }

  // ─── STALE DATA DETECTOR ──────────────────────────────────────────────────

  /// Cek apakah ada transaksi pending yang sudah terlalu lama belum tersinkron.
  /// Mengembalikan daftar transaksi yang berisiko hilang.
  Future<List<LocalTransaction>> detectStalePendingTransactions(
      AppDatabase db) async {
    try {
      final pending = await (db.select(db.localTransactions)
            ..where((t) => t.syncStatus.equals('pending')))
          .get();

      final threshold = DateTime.now().subtract(stalePendingThreshold);
      final stale =
          pending.where((t) => t.createdAt.isBefore(threshold)).toList();

      if (stale.isNotEmpty) {
        await log(
          category: 'INTEGRITY',
          action: 'STALE_DATA_DETECTED',
          detail:
              'Ditemukan ${stale.length} transaksi pending > ${stalePendingThreshold.inHours} jam: '
              '${stale.map((t) => t.localId.substring(0, 8)).join(', ')}',
          severity: AuditSeverity.warning,
        );
      }

      return stale;
    } catch (e) {
      debugPrint('[DataIntegrity] detectStalePendingTransactions error: $e');
      return [];
    }
  }

  /// Pastikan semua buffer audit ter-flush ke disk sebelum app di-close.
  Future<void> dispose() async {
    await _flushAuditBuffer();
  }

  // ─── INTEGRITY CHECK ──────────────────────────────────────────────────────

  /// Hitung total data yang belum tersinkronisasi. Berguna untuk UI badge.
  Future<DataIntegritySummary> getSummary(AppDatabase db) async {
    try {
      final pendingTx = await (db.select(db.localTransactions)
            ..where((t) => t.syncStatus.equals('pending')))
          .get();

      final pendingCheckins = await (db.select(db.localCheckins)
            ..where((t) => t.syncStatus.equals('pending')))
          .get();

      final stuckItems = await getStuckItems();
      final stale = await detectStalePendingTransactions(db);

      return DataIntegritySummary(
        pendingTransactions: pendingTx.length,
        pendingCheckins: pendingCheckins.length,
        stuckItemCount: stuckItems.length,
        stalePendingCount: stale.length,
        stuckLocalIds: stuckItems,
      );
    } catch (e) {
      debugPrint('[DataIntegrity] getSummary error: $e');
      return DataIntegritySummary.empty();
    }
  }
}

/// Ringkasan status integritas data.
class DataIntegritySummary {
  final int pendingTransactions;
  final int pendingCheckins;
  final int stuckItemCount;
  final int stalePendingCount;
  final List<String> stuckLocalIds;

  const DataIntegritySummary({
    required this.pendingTransactions,
    required this.pendingCheckins,
    required this.stuckItemCount,
    required this.stalePendingCount,
    required this.stuckLocalIds,
  });

  factory DataIntegritySummary.empty() => const DataIntegritySummary(
        pendingTransactions: 0,
        pendingCheckins: 0,
        stuckItemCount: 0,
        stalePendingCount: 0,
        stuckLocalIds: [],
      );

  bool get hasWarnings => stalePendingCount > 0 || stuckItemCount > 0;

  int get totalPending => pendingTransactions + pendingCheckins;

  @override
  String toString() =>
      'DataIntegritySummary(pendingTx=$pendingTransactions, pendingCheckins=$pendingCheckins, '
      'stuck=$stuckItemCount, stale=$stalePendingCount)';
}
