import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/features/attendance/domain/entities/attendance.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_bloc.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_event.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_state.dart';
import 'package:hris_app/main.dart';
import 'package:intl/intl.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:flutter_staggered_animations/flutter_staggered_animations.dart';
import 'package:hris_app/core/utils/constants.dart';

class ClockInSuccessPage extends StatefulWidget {
  final Attendance attendance;
  final String? branchName;
  final String? shiftName;
  final String? shiftStartTime;
  final String? shiftEndTime;

  const ClockInSuccessPage({
    super.key,
    required this.attendance,
    this.branchName,
    this.shiftName,
    this.shiftStartTime,
    this.shiftEndTime,
  });

  @override
  State<ClockInSuccessPage> createState() => _ClockInSuccessPageState();
}

class _ClockInSuccessPageState extends State<ClockInSuccessPage> {
  late Attendance _currentAttendance;
  bool _hasPromptedLateReason = false;

  @override
  void initState() {
    super.initState();
    _currentAttendance = widget.attendance;

    WidgetsBinding.instance.addPostFrameCallback((_) {
      _checkAndPromptLateReason();
    });
  }

  void _checkAndPromptLateReason() {
    final bool isLate =
        _currentAttendance.status.toUpperCase() == 'LATE';
    final bool isClockIn = _currentAttendance.checkOut == null;

    if (isLate && isClockIn && _currentAttendance.notes.isEmpty && !_hasPromptedLateReason) {
      _hasPromptedLateReason = true;
      _showLateReasonBottomSheet();
    }
  }

  void _showLateReasonBottomSheet() {
    final TextEditingController reasonController = TextEditingController();

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      isDismissible: false,
      enableDrag: false,
      backgroundColor: Colors.transparent,
      builder: (bottomSheetContext) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return BlocConsumer<AttendanceBloc, AttendanceState>(
              listener: (context, state) {
                if (state is LateReasonSuccess) {
                  setState(() {
                    _currentAttendance =
                        _currentAttendance.copyWith(notes: state.reason);
                  });
                  Navigator.of(bottomSheetContext).pop();
                  SnackBarUtils.showSuccess(
                    context,
                    'Alasan keterlambatan berhasil disimpan.',
                  );
                } else if (state is LateReasonFailure) {
                  SnackBarUtils.showError(context, state.message);
                }
              },
              builder: (context, state) {
                final bool isSubmitting = state is LateReasonSubmitting;

                return Container(
                  padding: EdgeInsets.only(
                    left: 24,
                    right: 24,
                    top: 24,
                    bottom: MediaQuery.of(bottomSheetContext).viewInsets.bottom + 24,
                  ),
                  decoration: const BoxDecoration(
                    color: Colors.white,
                    borderRadius:
                        BorderRadius.vertical(top: Radius.circular(28)),
                  ),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Center(
                        child: Container(
                          width: 40,
                          height: 4,
                          decoration: BoxDecoration(
                            color: Colors.grey.shade300,
                            borderRadius: BorderRadius.circular(2),
                          ),
                        ),
                      ),
                      const SizedBox(height: 20),
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: AppColors.warning.withOpacity(0.12),
                              shape: BoxShape.circle,
                            ),
                            child: const Icon(
                              Icons.access_time_filled_rounded,
                              color: AppColors.warning,
                              size: 28,
                            ),
                          ),
                          const SizedBox(width: 14),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Anda Terlambat Masuk',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w900,
                                    color: AppColors.textPrimary,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  'Waktu masuk melewati jadwal shift kerja.',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 12,
                                    color: AppColors.textSecondary,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 20),
                      Text(
                        'Ketik Alasan Keterlambatan *',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 13,
                          fontWeight: FontWeight.w800,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 8),
                      TextField(
                        controller: reasonController,
                        maxLines: 3,
                        enabled: !isSubmitting,
                        decoration: InputDecoration(
                          hintText:
                              'Tuliskan secara jelas alasan Anda terlambat absen masuk...',
                          hintStyle: GoogleFonts.plusJakartaSans(
                            fontSize: 13,
                            color: AppColors.textTertiary,
                          ),
                          contentPadding: const EdgeInsets.all(14),
                          filled: true,
                          fillColor: AppColors.grayLight.withOpacity(0.5),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(16),
                            borderSide:
                                BorderSide(color: Colors.grey.shade200),
                          ),
                          focusedBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(16),
                            borderSide: const BorderSide(
                                color: AppColors.warning, width: 1.5),
                          ),
                        ),
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 24),
                      SizedBox(
                        width: double.infinity,
                        height: 52,
                        child: ElevatedButton(
                          onPressed: isSubmitting
                              ? null
                              : () {
                                  final reason = reasonController.text.trim();
                                  if (reason.isEmpty) {
                                    SnackBarUtils.showError(
                                      bottomSheetContext,
                                      'Alasan keterlambatan wajib diisi.',
                                    );
                                    return;
                                  }

                                  context.read<AttendanceBloc>().add(
                                        SubmitLateReasonRequested(
                                          attendanceId:
                                              _currentAttendance.id,
                                          reason: reason,
                                        ),
                                      );
                                },
                          style: ElevatedButton.styleFrom(
                            backgroundColor: AppColors.warning,
                            foregroundColor: Colors.white,
                            disabledBackgroundColor: Colors.grey.shade300,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                            ),
                            elevation: 0,
                          ),
                          child: isSubmitting
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2.5,
                                    color: Colors.white,
                                  ),
                                )
                              : Text(
                                  'KIRIM CATATAN ALASAN',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w900,
                                    letterSpacing: 0.5,
                                  ),
                                ),
                        ),
                      ),
                    ],
                  ),
                );
              },
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final isClockOut = _currentAttendance.checkOut != null;
    final displayTime =
        isClockOut ? _currentAttendance.checkOut! : _currentAttendance.checkIn;

    // Formatting for Indonesian locale
    final timeStr = DateFormat('HH.mm').format(displayTime);
    final dateStr =
        DateFormat('EEEE, dd MMMM yyyy', 'id_ID').format(displayTime);

    final bool hasSelfie = _currentAttendance.selfiePath.isNotEmpty;
    final String selfieUrl = hasSelfie
        ? '${AppConstants.baseUrl.replaceAll('/api/v1/', '')}${_currentAttendance.selfiePath}'
        : '';

    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      body: SafeArea(
        child: AnimationLimiter(
          child: SingleChildScrollView(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: AnimationConfiguration.toStaggeredList(
                  duration: const Duration(milliseconds: 600),
                  childAnimationBuilder: (widget) => SlideAnimation(
                    verticalOffset: 50.0,
                    child: FadeInAnimation(
                      child: widget,
                    ),
                  ),
                  children: [
                    const SizedBox(height: 60),

                    // Success Icon with Pulse effect
                    Center(
                      child: Stack(
                        alignment: Alignment.center,
                        children: [
                          Container(
                            width: 120,
                            height: 120,
                            decoration: BoxDecoration(
                              color: AppColors.success.withOpacity(0.1),
                              shape: BoxShape.circle,
                            ),
                          ),
                          Container(
                            width: 90,
                            height: 90,
                            decoration: BoxDecoration(
                              color: AppColors.success,
                              shape: BoxShape.circle,
                              boxShadow: [
                                BoxShadow(
                                  color: AppColors.success.withOpacity(0.3),
                                  blurRadius: 20,
                                  offset: const Offset(0, 10),
                                )
                              ],
                            ),
                            child: const Icon(
                              Icons.check_rounded,
                              color: Colors.white,
                              size: 54,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 40),

                    // Title
                    Text(
                      isClockOut
                          ? 'Absensi Keluar Berhasil!'
                          : 'Absensi Masuk Berhasil!',
                      textAlign: TextAlign.center,
                      style: GoogleFonts.outfit(
                        fontSize: 28,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                        letterSpacing: -0.5,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      isClockOut
                          ? 'Terima kasih atas kerja keras Anda hari ini. Selamat beristirahat!'
                          : 'Selamat bekerja! Semoga hari Anda produktif dan menyenangkan.',
                      textAlign: TextAlign.center,
                      style: GoogleFonts.outfit(
                        fontSize: 16,
                        color: AppColors.textSecondary,
                        height: 1.5,
                      ),
                    ),
                    const SizedBox(height: 48),

                    // Info Card
                    Container(
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(24),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.03),
                            blurRadius: 20,
                            offset: const Offset(0, 10),
                          )
                        ],
                        border: Border.all(
                            color: AppColors.grayBorder.withOpacity(0.5)),
                      ),
                      child: Column(
                        children: [
                          Padding(
                            padding: const EdgeInsets.all(24.0),
                            child: Column(
                              children: [
                                Row(
                                  children: [
                                    Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          dateStr,
                                          style: GoogleFonts.outfit(
                                            fontSize: 14,
                                            fontWeight: FontWeight.w600,
                                            color: AppColors.textTertiary,
                                          ),
                                        ),
                                        const SizedBox(height: 4),
                                        Row(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.baseline,
                                          textBaseline: TextBaseline.alphabetic,
                                          children: [
                                            Text(
                                              timeStr,
                                              style: GoogleFonts.outfit(
                                                fontSize: 48,
                                                fontWeight: FontWeight.w800,
                                                color: AppColors.textPrimary,
                                                letterSpacing: -1,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                    const Spacer(),
                                    // Selfie Preview
                                    Container(
                                      width: 84,
                                      height: 84,
                                      decoration: BoxDecoration(
                                        color: AppColors.grayLight,
                                        borderRadius: BorderRadius.circular(20),
                                        border: Border.all(
                                            color: AppColors.success
                                                .withOpacity(0.2),
                                            width: 2),
                                      ),
                                      child: ClipRRect(
                                        borderRadius: BorderRadius.circular(18),
                                        child: hasSelfie
                                            ? CachedNetworkImage(imageUrl: selfieUrl, fit: BoxFit.cover, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,)
                                            : const Icon(Icons.person_rounded,
                                                color: AppColors.textTertiary,
                                                size: 40),
                                      ),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 24),
                                const Divider(height: 1),
                                const SizedBox(height: 20),

                                // Location Row
                                Row(
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.all(8),
                                      decoration: BoxDecoration(
                                        color: AppColors.info.withOpacity(0.1),
                                        borderRadius: BorderRadius.circular(10),
                                      ),
                                      child: const Icon(
                                          Icons.location_on_rounded,
                                          size: 20,
                                          color: AppColors.info),
                                    ),
                                    const SizedBox(width: 16),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            'LOKASI PRESENSI',
                                            style: GoogleFonts.outfit(
                                              fontSize: 10,
                                              fontWeight: FontWeight.w800,
                                              color: AppColors.textTertiary,
                                              letterSpacing: 1,
                                            ),
                                          ),
                                          const SizedBox(height: 2),
                                          Text(
                                            (widget.branchName != null &&
                                                    widget.branchName!.isNotEmpty)
                                                ? widget.branchName!
                                                : (_currentAttendance
                                                        .branchName.isNotEmpty
                                                    ? _currentAttendance.branchName
                                                    : 'Cabang Terdaftar'),
                                            style: GoogleFonts.outfit(
                                              fontSize: 15,
                                              fontWeight: FontWeight.w700,
                                              color: AppColors.textPrimary,
                                            ),
                                            overflow: TextOverflow.ellipsis,
                                          ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),

                                // Late Reason Note Card if available
                                if (_currentAttendance.notes.isNotEmpty) ...[
                                  const SizedBox(height: 16),
                                  Container(
                                    padding: const EdgeInsets.all(12),
                                    decoration: BoxDecoration(
                                      color: AppColors.warning.withOpacity(0.08),
                                      borderRadius: BorderRadius.circular(14),
                                      border: Border.all(
                                        color: AppColors.warning.withOpacity(0.25),
                                      ),
                                    ),
                                    child: Row(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        const Icon(
                                          Icons.edit_note_rounded,
                                          color: AppColors.warning,
                                          size: 20,
                                        ),
                                        const SizedBox(width: 10),
                                        Expanded(
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              Text(
                                                'Alasan Keterlambatan:',
                                                style:
                                                    GoogleFonts.plusJakartaSans(
                                                  fontSize: 11,
                                                  fontWeight: FontWeight.w800,
                                                  color: AppColors.warning,
                                                ),
                                              ),
                                              const SizedBox(height: 2),
                                              Text(
                                                _currentAttendance.notes,
                                                style:
                                                    GoogleFonts.plusJakartaSans(
                                                  fontSize: 12,
                                                  fontWeight: FontWeight.w600,
                                                  color: AppColors.textPrimary,
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              ],
                            ),
                          ),

                          // Status Bar bottom
                          Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 24, vertical: 16),
                            decoration: BoxDecoration(
                              color: AppColors.grayLight.withOpacity(0.5),
                              borderRadius: const BorderRadius.only(
                                bottomLeft: Radius.circular(24),
                                bottomRight: Radius.circular(24),
                              ),
                            ),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Row(
                                  children: [
                                    const Icon(Icons.schedule_rounded,
                                        size: 16,
                                        color: AppColors.textSecondary),
                                    const SizedBox(width: 8),
                                    Text(
                                      widget.shiftName != null &&
                                              widget.shiftStartTime != null &&
                                              widget.shiftEndTime != null
                                          ? '${widget.shiftName} (${widget.shiftStartTime} - ${widget.shiftEndTime})'
                                          : 'Shift: Pagi (07:30)',
                                      style: GoogleFonts.outfit(
                                        fontSize: 13,
                                        fontWeight: FontWeight.w600,
                                        color: AppColors.textSecondary,
                                      ),
                                    ),
                                  ],
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 12, vertical: 6),
                                  decoration: BoxDecoration(
                                    color: (_currentAttendance.status.toLowerCase() ==
                                                'on_time'
                                            ? AppColors.success
                                            : AppColors.warning)
                                        .withOpacity(0.1),
                                    borderRadius: BorderRadius.circular(20),
                                  ),
                                  child: Text(
                                    _currentAttendance.status.toLowerCase() == 'on_time'
                                        ? 'TEPAT WAKTU'
                                        : 'TERLAMBAT',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                      color: _currentAttendance.status.toLowerCase() ==
                                              'on_time'
                                          ? AppColors.success
                                          : AppColors.warning,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 32),

                    // Back to Home Button
                    ElevatedButton(
                      onPressed: () {
                        Navigator.of(context).pushAndRemoveUntil(
                          MaterialPageRoute(
                              builder: (_) => const AuthWrapper()),
                          (route) => false,
                        );
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.textPrimary,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 18),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                        elevation: 0,
                      ),
                      child: Text(
                        'Kembali ke Beranda',
                        style: GoogleFonts.outfit(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
