import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:shared_preferences/shared_preferences.dart';

class WarningPopupDialog extends StatelessWidget {
  final dynamic warningData;
  final VoidCallback? onAcknowledge;

  const WarningPopupDialog({
    super.key,
    required this.warningData,
    this.onAcknowledge,
  });

  static Future<void> show({
    required BuildContext context,
    required dynamic warningData,
    VoidCallback? onAcknowledge,
  }) async {
    if (warningData == null) return;

    // Check if this specific warning ID has already been acknowledged
    String warningId = '';
    if (warningData is Map) {
      warningId = (warningData['id'] ?? warningData['ID'] ?? '').toString();
    }

    if (warningId.isNotEmpty) {
      final prefs = await SharedPreferences.getInstance();
      final acknowledgedId = prefs.getString('acknowledged_warning_id');
      if (acknowledgedId == warningId) {
        // Already acknowledged this specific warning
        return;
      }
    }

    if (!context.mounted) return;

    await showGeneralDialog(
      context: context,
      barrierDismissible: false,
      barrierLabel: '',
      barrierColor: Colors.black.withOpacity(0.65),
      transitionDuration: const Duration(milliseconds: 350),
      pageBuilder: (dialogContext, anim1, anim2) {
        return Center(
          child: WarningPopupDialog(
            warningData: warningData,
            onAcknowledge: () async {
              if (warningId.isNotEmpty) {
                final prefs = await SharedPreferences.getInstance();
                await prefs.setString('acknowledged_warning_id', warningId);
              }
              if (onAcknowledge != null) {
                onAcknowledge();
              }
              if (dialogContext.mounted) {
                Navigator.of(dialogContext, rootNavigator: true).pop();
              }
            },
          ),
        );
      },
      transitionBuilder: (context, anim1, anim2, child) {
        final curvedValue = Curves.easeOutBack.transform(anim1.value);
        return Transform.scale(
          scale: curvedValue,
          child: Opacity(
            opacity: anim1.value,
            child: child,
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    String warningType = '-';
    String notes = '';
    String issueDesc = 'Tidak ada kunjungan & transaksi dalam 2 hari terakhir berturut-turut.';
    String handledBy = '';
    String dateStr = '';

    if (warningData is Map) {
      warningType = (warningData['warning_type'] ?? warningData['WarningType'] ?? '').toString().toLowerCase();
      notes = (warningData['issue_description'] ?? warningData['IssueDescription'] ?? warningData['notes'] ?? '').toString();
      if (warningData['handled_by'] != null) {
        handledBy = warningData['handled_by'].toString();
      }
      if (warningData['created_at'] != null) {
        dateStr = warningData['created_at'].toString().split('T').first;
      }
    }

    Color primaryColor = const Color(0xFFEF4444); // Default Red
    Color bgColor = const Color(0xFFFEF2F2);
    Color borderColor = const Color(0xFFFECACA);
    String title = 'Peringatan HRD';
    String levelLabel = 'Surat Peringatan';
    IconData icon = Icons.warning_amber_rounded;

    if (warningType == 'cs_followup') {
      primaryColor = const Color(0xFF2563EB); // Blue
      bgColor = const Color(0xFFEFF6FF);
      borderColor = const Color(0xFFBFDBFE);
      title = 'Tindak Lanjut CS';
      levelLabel = 'Level 1: Tindak Lanjut CS';
      icon = Icons.support_agent_rounded;
    } else if (warningType == 'spv_followup') {
      primaryColor = const Color(0xFFEA580C); // Orange
      bgColor = const Color(0xFFFFF7ED);
      borderColor = const Color(0xFFFFEDD5);
      title = 'Tindak Lanjut SPV Sales';
      levelLabel = 'Level 2: Tindak Lanjut SPV';
      icon = Icons.supervisor_account_rounded;
    } else if (warningType == 'teguran') {
      primaryColor = const Color(0xFFDC2626); // Red
      bgColor = const Color(0xFFFEF2F2);
      borderColor = const Color(0xFFFECACA);
      title = 'Teguran Disiplin';
      levelLabel = 'Teguran HRD';
      icon = Icons.report_problem_rounded;
    } else if (warningType.startsWith('sp')) {
      primaryColor = const Color(0xFFB91C1C); // Dark Red
      bgColor = const Color(0xFFFEF2F2);
      borderColor = const Color(0xFFFCA5A5);
      title = 'Surat Peringatan (${warningType.toUpperCase()})';
      levelLabel = 'Level 3: ${warningType.toUpperCase()}';
      icon = Icons.gavel_rounded;
    }

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
      elevation: 0,
      backgroundColor: Colors.transparent,
      child: Container(
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(28),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.18),
              blurRadius: 30,
              offset: const Offset(0, 15),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            // Top Badge / Icon
            Container(
              padding: const EdgeInsets.all(18),
              decoration: BoxDecoration(
                color: bgColor,
                shape: BoxShape.circle,
                border: Border.all(color: borderColor, width: 2),
              ),
              child: Icon(
                icon,
                color: primaryColor,
                size: 48,
              ),
            ),
            const SizedBox(height: 18),

            // Title
            Text(
              title,
              textAlign: TextAlign.center,
              style: GoogleFonts.plusJakartaSans(
                fontSize: 20,
                fontWeight: FontWeight.w800,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 6),

            // Level pill
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              decoration: BoxDecoration(
                color: bgColor,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: borderColor),
              ),
              child: Text(
                levelLabel,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                  color: primaryColor,
                ),
              ),
            ),
            const SizedBox(height: 18),

            // Message / Notes container
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: const Color(0xFFE2E8F0)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.info_outline_rounded, size: 16, color: primaryColor),
                      const SizedBox(width: 6),
                      Text(
                        'Penyebab Tindakan:',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                          color: const Color(0xFF334155),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    issueDesc,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      color: const Color(0xFF475569),
                      height: 1.4,
                    ),
                  ),
                  if (notes.isNotEmpty) ...[
                    const Divider(height: 16, color: Color(0xFFE2E8F0)),
                    Row(
                      children: [
                        const Icon(Icons.comment_rounded, size: 16, color: Color(0xFF64748B)),
                        const SizedBox(width: 6),
                        Text(
                          'Catatan Petugas:',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            fontWeight: FontWeight.w700,
                            color: const Color(0xFF334155),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Text(
                      notes,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: const Color(0xFF1E293B),
                        height: 1.4,
                      ),
                    ),
                  ],
                  if (dateStr.isNotEmpty || handledBy.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text(
                      [
                        if (dateStr.isNotEmpty) 'Tanggal: $dateStr',
                        if (handledBy.isNotEmpty) 'Oleh: $handledBy',
                      ].join(' • '),
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 11,
                        color: const Color(0xFF94A3B8),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: 16),

            // Instruction prompt
            Text(
              'Harap segera lakukan check-in kunjungan toko dan catat transaksi harian secara tertib melalui aplikasi.',
              textAlign: TextAlign.center,
              style: GoogleFonts.plusJakartaSans(
                fontSize: 12,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
            ),
            const SizedBox(height: 24),

            // Confirmation Button
            SizedBox(
              width: double.infinity,
              height: 48,
              child: ElevatedButton(
                onPressed: onAcknowledge,
                style: ElevatedButton.styleFrom(
                  backgroundColor: primaryColor,
                  foregroundColor: Colors.white,
                  elevation: 0,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(14),
                  ),
                ),
                child: Text(
                  'SAYA MENGERTI & AKAN TERTIB',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 0.5,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
