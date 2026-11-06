import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/announcement/data/models/announcement_model.dart';

class AnnouncementPopupDialog extends StatefulWidget {
  final dynamic announcement;
  final bool isAlreadyConfirmed;
  final VoidCallback? onAcknowledge;

  const AnnouncementPopupDialog({
    super.key,
    required this.announcement,
    this.isAlreadyConfirmed = false,
    this.onAcknowledge,
  });

  static Future<void> show({
    required BuildContext context,
    required dynamic announcement,
    bool isAlreadyConfirmed = false,
    VoidCallback? onAcknowledge,
  }) {
    return showGeneralDialog(
      context: context,
      barrierDismissible: isAlreadyConfirmed,
      barrierLabel: 'Tutup Pengumuman',
      barrierColor: Colors.black.withOpacity(0.75),
      transitionDuration: const Duration(milliseconds: 250),
      transitionBuilder: (context, anim1, anim2, child) {
        return FadeTransition(
          opacity: CurvedAnimation(parent: anim1, curve: Curves.easeOut),
          child: ScaleTransition(
            scale: Tween<double>(begin: 0.95, end: 1.0).animate(
              CurvedAnimation(parent: anim1, curve: Curves.easeOutCubic),
            ),
            child: child,
          ),
        );
      },
      pageBuilder: (dialogContext, anim1, anim2) {
        return PopScope(
          canPop: isAlreadyConfirmed,
          child: AnnouncementPopupDialog(
            announcement: announcement,
            isAlreadyConfirmed: isAlreadyConfirmed,
            onAcknowledge: onAcknowledge,
          ),
        );
      },
    );
  }

  @override
  State<AnnouncementPopupDialog> createState() =>
      _AnnouncementPopupDialogState();
}

class _AnnouncementPopupDialogState extends State<AnnouncementPopupDialog> {
  final ScrollController _scrollController = ScrollController();
  bool _hasScrolledToBottom = false;
  late final bool _needsScroll;

  @override
  void initState() {
    super.initState();

    final String content = widget.announcement.content ?? '';
    // Teks di atas 220 karakter atau yang memiliki baris baru biasanya memerlukan scroll
    // pada container berukuran maxHeight: 220
    final hasMultipleLines = content.split('\n').length > 5;
    _needsScroll = !widget.isAlreadyConfirmed && (content.length > 250 || hasMultipleLines);

    if (widget.isAlreadyConfirmed || !_needsScroll) {
      _hasScrolledToBottom = true;
    } else {
      _scrollController.addListener(_onScroll);
    }
  }

  void _onScroll() {
    if (!mounted || !_scrollController.hasClients) return;
    final maxScroll = _scrollController.position.maxScrollExtent;
    final currentScroll = _scrollController.position.pixels;

    if (maxScroll <= 0 || currentScroll >= maxScroll - 25) {
      if (!_hasScrolledToBottom) {
        setState(() {
          _hasScrolledToBottom = true;
        });
      }
    }
  }

  @override
  void dispose() {
    _scrollController.removeListener(_onScroll);
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final popup = widget.announcement;

    final isUrgent = popup.category == AnnouncementCategory.urgentCall ||
        popup.category == AnnouncementCategory.alert;
    final isEvent = popup.category == AnnouncementCategory.event;

    final Color headerColor = isUrgent
        ? Colors.red.shade600
        : (isEvent ? Colors.blue.shade700 : AppColors.primaryGreen);
    final Color badgeBg = isUrgent
        ? Colors.red.shade50
        : (isEvent ? Colors.blue.shade50 : Colors.green.shade50);
    final IconData headerIcon = isUrgent
        ? Icons.warning_amber_rounded
        : (isEvent ? Icons.event_available_rounded : Icons.campaign_rounded);
    final String badgeText = isUrgent
        ? 'PANGGILAN MENGHADAP SEGERA'
        : (isEvent ? 'UNDANGAN MEETING / ACARA' : 'PENGUMUMAN PENTING');

    final String title = popup.title ?? 'Pemberitahuan Penting';
    final String content = popup.content ?? '';

    return Dialog(
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
      elevation: 0,
      child: Container(
        width: double.infinity,
        constraints: const BoxConstraints(maxWidth: 420),
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(28),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.25),
              blurRadius: 30,
              offset: const Offset(0, 15),
            )
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            // Header Icon Badge
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: badgeBg,
                shape: BoxShape.circle,
              ),
              child: Icon(headerIcon, size: 42, color: headerColor),
            ),
            const SizedBox(height: 14),

            // Category Badge
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              decoration: BoxDecoration(
                color: badgeBg,
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(
                badgeText,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  color: headerColor,
                  letterSpacing: 0.8,
                ),
              ),
            ),
            const SizedBox(height: 12),

            // Title
            Text(
              title,
              style: GoogleFonts.plusJakartaSans(
                fontSize: 17,
                fontWeight: FontWeight.w900,
                color: AppColors.textPrimary,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),

            // Content Box with ScrollController
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: Colors.grey.shade200),
              ),
              constraints: const BoxConstraints(maxHeight: 220),
              child: Scrollbar(
                controller: _scrollController,
                child: SingleChildScrollView(
                  controller: _scrollController,
                  physics: const BouncingScrollPhysics(),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: Text(
                      content,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13.5,
                        height: 1.55,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500,
                      ),
                      textAlign: TextAlign.left,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(height: 12),

            // Scroll Indicator / Read Status Hint
            if (!widget.isAlreadyConfirmed) ...[
              if (_needsScroll && !_hasScrolledToBottom)
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: Colors.amber.shade50,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.amber.shade200),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.arrow_downward_rounded,
                            size: 14, color: Colors.amber.shade900),
                        const SizedBox(width: 6),
                        Flexible(
                          child: Text(
                            'Gulir ke bawah untuk mengaktifkan tombol konfirmasi',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              color: Colors.amber.shade900,
                            ),
                          ),
                        ),
                      ],
                    ),
                  )
                else if (_needsScroll && _hasScrolledToBottom)
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: Colors.green.shade50,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.green.shade200),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.check_circle_rounded,
                            size: 14, color: Colors.green),
                        const SizedBox(width: 6),
                        Text(
                          'Anda telah membaca seluruh isi pengumuman',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 11,
                            fontWeight: FontWeight.w700,
                            color: Colors.green.shade800,
                          ),
                        ),
                      ],
                    ),
                  ),
                const SizedBox(height: 14),
              ],

              // Action Button
              if (widget.isAlreadyConfirmed)
                SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: OutlinedButton.icon(
                    style: OutlinedButton.styleFrom(
                      side: BorderSide(color: Colors.grey.shade300),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                    ),
                    onPressed: () => Navigator.of(context).pop(),
                    icon: const Icon(Icons.check_circle_rounded,
                        color: Colors.green, size: 18),
                    label: Text(
                      'TUTUP (SUDAH DIKONFIRMASI)',
                      style: GoogleFonts.plusJakartaSans(
                        color: AppColors.textPrimary,
                        fontWeight: FontWeight.w800,
                        fontSize: 12,
                      ),
                    ),
                  ),
                )
              else
                SizedBox(
                  width: double.infinity,
                  height: 50,
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: _hasScrolledToBottom
                          ? headerColor
                          : Colors.grey.shade200,
                      foregroundColor: _hasScrolledToBottom
                          ? Colors.white
                          : Colors.grey.shade500,
                      disabledBackgroundColor: Colors.grey.shade200,
                      disabledForegroundColor: Colors.grey.shade400,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      elevation: _hasScrolledToBottom ? 3 : 0,
                    ),
                    onPressed: _hasScrolledToBottom
                        ? () {
                            widget.onAcknowledge?.call();
                            Navigator.of(context).pop();
                          }
                        : null,
                    icon: Icon(
                      _hasScrolledToBottom
                          ? Icons.check_circle_rounded
                          : Icons.lock_outline_rounded,
                      color: _hasScrolledToBottom
                          ? Colors.white
                          : Colors.grey.shade400,
                      size: 20,
                    ),
                    label: Text(
                      _hasScrolledToBottom
                          ? 'SAYA MENGERTI & KONFIRMASI'
                          : 'GULIR KE BAWAH UNTUK KONFIRMASI',
                      style: GoogleFonts.plusJakartaSans(
                        color: _hasScrolledToBottom
                            ? Colors.white
                            : Colors.grey.shade500,
                        fontWeight: FontWeight.w800,
                        fontSize: 12,
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
