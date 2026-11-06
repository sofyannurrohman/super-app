import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';

/// Widget banner yang menampilkan status sinkronisasi data offline.
///
/// Pasang widget ini di halaman utama sales (SalesDashboardPage) sebagai
/// lapisan paling atas sebelum konten utama.
///
/// Perilaku:
/// - Hilang jika semua data sudah tersinkronisasi.
/// - Menampilkan badge kuning jika ada data pending.
/// - Menampilkan banner merah jika ada data yang STUCK (gagal sync berulang kali)
///   atau sudah terlalu lama pending (> 4 jam), menandakan user perlu segera
///   mencari koneksi yang lebih baik.
class SyncStatusBanner extends StatelessWidget {
  const SyncStatusBanner({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<SyncBloc, SyncState>(
      listenWhen: (prev, curr) => curr is SyncIntegrityWarning,
      listener: (context, state) {
        if (state is SyncIntegrityWarning) {
          _showIntegrityAlert(context, state.summary);
        }
      },
      buildWhen: (prev, curr) =>
          curr is SyncInProgress ||
          curr is SyncSuccess ||
          curr is SyncFailure ||
          curr is SyncIntegrityWarning,
      builder: (context, state) {
        if (state is SyncInProgress) {
          return _buildBanner(
            icon: Icons.sync_rounded,
            message: 'Menyinkronisasikan data...',
            color: const Color(0xFF2563EB),
            isAnimated: true,
          );
        }

        if (state is SyncIntegrityWarning) {
          final summary = state.summary;
          if (summary.stuckItemCount > 0) {
            return _buildBanner(
              icon: Icons.warning_amber_rounded,
              message:
                  '⚠️ ${summary.stuckItemCount} transaksi GAGAL sync berulang kali. '
                  'Segera cari sinyal yang lebih baik atau hubungi IT.',
              color: const Color(0xFFDC2626),
              isAnimated: false,
              showRetry: true,
              onRetry: () => context.read<SyncBloc>().add(const SyncDataRequested(force: true)),
            );
          }

          if (summary.stalePendingCount > 0) {
            return _buildBanner(
              icon: Icons.hourglass_bottom_rounded,
              message:
                  '⏳ ${summary.stalePendingCount} transaksi belum tersinkron > 4 jam. '
                  'Pastikan HP terhubung ke internet.',
              color: const Color(0xFFD97706),
              isAnimated: false,
              showRetry: true,
              onRetry: () => context.read<SyncBloc>().add(const SyncDataRequested(force: true)),
            );
          }
        }

        if (state is SyncFailure) {
          return _buildBanner(
            icon: Icons.cloud_off_rounded,
            message: 'Sync gagal. Data aman di HP. Akan dicoba ulang otomatis.',
            color: const Color(0xFF6B7280),
            isAnimated: false,
          );
        }

        return const SizedBox.shrink();
      },
    );
  }

  Widget _buildBanner({
    required IconData icon,
    required String message,
    required Color color,
    required bool isAnimated,
    bool showRetry = false,
    VoidCallback? onRetry,
  }) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          if (isAnimated)
            const SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(
                strokeWidth: 2,
                valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
              ),
            )
          else
            Icon(icon, color: Colors.white, size: 16),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              message,
              style: GoogleFonts.outfit(
                color: Colors.white,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          if (showRetry && onRetry != null) ...[
            const SizedBox(width: 8),
            GestureDetector(
              onTap: onRetry,
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.25),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  'Coba Ulang',
                  style: GoogleFonts.outfit(
                    color: Colors.white,
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  void _showIntegrityAlert(BuildContext context, DataIntegritySummary summary) {
    if (!summary.hasWarnings) return;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        duration: const Duration(seconds: 8),
        backgroundColor: summary.stuckItemCount > 0
            ? const Color(0xFFDC2626)
            : const Color(0xFFD97706),
        behavior: SnackBarBehavior.floating,
        margin: const EdgeInsets.all(12),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              summary.stuckItemCount > 0
                  ? '🚨 Perhatian: Data Berisiko!'
                  : '⚠️ Peringatan Integritas Data',
              style: GoogleFonts.outfit(
                color: Colors.white,
                fontWeight: FontWeight.w700,
                fontSize: 13,
              ),
            ),
            const SizedBox(height: 4),
            if (summary.pendingTransactions > 0)
              Text(
                '• ${summary.pendingTransactions} transaksi belum tersinkron',
                style: GoogleFonts.outfit(color: Colors.white70, fontSize: 12),
              ),
            if (summary.stalePendingCount > 0)
              Text(
                '• ${summary.stalePendingCount} transaksi sudah pending > 4 jam',
                style: GoogleFonts.outfit(color: Colors.white70, fontSize: 12),
              ),
            if (summary.stuckItemCount > 0)
              Text(
                '• ${summary.stuckItemCount} item GAGAL sync berulang kali!',
                style: GoogleFonts.outfit(
                  color: Colors.white,
                  fontWeight: FontWeight.w700,
                  fontSize: 12,
                ),
              ),
          ],
        ),
        action: SnackBarAction(
          label: 'Mengerti',
          textColor: Colors.white,
          onPressed: () {},
        ),
      ),
    );
  }
}
