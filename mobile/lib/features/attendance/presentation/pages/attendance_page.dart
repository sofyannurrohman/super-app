import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_bloc.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_event.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_state.dart';
import 'package:hris_app/features/attendance/presentation/pages/face_verification_page.dart';
import 'package:hris_app/features/attendance/domain/entities/attendance.dart';
import 'package:shimmer/shimmer.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_event.dart';

class AttendancePage extends StatefulWidget {
  const AttendancePage({super.key});

  @override
  State<AttendancePage> createState() => _AttendancePageState();
}

class _AttendancePageState extends State<AttendancePage> {
  late Timer _timer;
  DateTime _currentTime = DateTime.now();
  bool _isOutOfTown = false;
  final TextEditingController _destinationController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (mounted) {
        setState(() {
          _currentTime = DateTime.now();
        });
      }
    });
    context.read<AttendanceBloc>().add(FetchHomeDataRequested());
  }

  @override
  void dispose() {
    _timer.cancel();
    _destinationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      body: BlocBuilder<AttendanceBloc, AttendanceState>(
        builder: (context, state) {
          if (state is AttendanceLoading) {
            return _buildLoadingState();
          }

          // Handle offline state
          final authState = context.read<AuthBloc>().state;
          if (authState is Authenticated && authState.userProfile == null) {
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24.0),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.wifi_off_rounded,
                        size: 80, color: AppColors.textTertiary),
                    const SizedBox(height: 16),
                    Text(
                      'Tidak Ada Koneksi Internet',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 18,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Data profil Anda diperlukan untuk memverifikasi absensi. Pastikan koneksi internet stabil untuk memuat data awal.',
                      textAlign: TextAlign.center,
                      style: GoogleFonts.plusJakartaSans(
                          color: AppColors.textSecondary),
                    ),
                    const SizedBox(height: 24),
                    ElevatedButton.icon(
                      onPressed: () {
                        context
                            .read<AttendanceBloc>()
                            .add(FetchHomeDataRequested());
                        context
                            .read<AuthBloc>()
                            .add(CheckAuthStatusRequested());
                      },
                      icon: const Icon(Icons.refresh_rounded,
                          color: Colors.white),
                      label: Text(
                        'Coba Lagi',
                        style: GoogleFonts.plusJakartaSans(
                          fontWeight: FontWeight.w700,
                          color: Colors.white,
                        ),
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primaryRed,
                        padding: const EdgeInsets.symmetric(
                            horizontal: 24, vertical: 12),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12)),
                      ),
                    )
                  ],
                ),
              ),
            );
          }

          final status = state.attendanceStatus;
          final bool isClockedIn = status == AttendanceStatus.clockedIn;
          final bool isCompleted = status == AttendanceStatus.completed;

          Attendance? lastAttendance;
          if (state is HomeDataLoaded && state.history.isNotEmpty) {
            lastAttendance = state.history.first;
          }

          final String workDuration =
              _calculateWorkDuration(lastAttendance?.checkIn);

          return RefreshIndicator(
            onRefresh: () async {
              context.read<AttendanceBloc>().add(FetchHomeDataRequested());
            },
            child: CustomScrollView(
              slivers: [
                _buildSliverAppBar(),
                SliverPadding(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
                  sliver: SliverList(
                    delegate: SliverChildListDelegate([
                      _buildAttendanceTypeSelector(isClockedIn, lastAttendance),
                      const SizedBox(height: 16),
                      _buildTimerCard(isClockedIn, workDuration, lastAttendance,
                          isCompleted),
                      const SizedBox(height: 32),
                      _buildMainActionButton(isClockedIn, isCompleted, lastAttendance),
                      const SizedBox(height: 32),
                      _buildStatusIndicators(),
                      const SizedBox(height: 32),
                      _buildDailySummary(state),
                      const SizedBox(height: 100),
                    ]),
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildSliverAppBar() {
    return SliverAppBar(
      expandedHeight: 180,
      pinned: true,
      backgroundColor: AppColors.primaryRed,
      elevation: 0,
      flexibleSpace: FlexibleSpaceBar(
        background: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const SizedBox(height: 40),
              Text(
                DateFormat('HH:mm:ss').format(_currentTime),
                style: GoogleFonts.plusJakartaSans(
                  color: Colors.white,
                  fontSize: 48,
                  fontWeight: FontWeight.w900,
                  letterSpacing: -1,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                DateFormat('EEEE, dd MMMM yyyy', 'id_ID')
                    .format(_currentTime)
                    .toUpperCase(),
                style: GoogleFonts.plusJakartaSans(
                  color: Colors.white.withOpacity(0.8),
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.5,
                ),
              ),
            ],
          ),
        ),
      ),
      title: const Text('PRESENSI KERJA',
          style: TextStyle(
              color: Colors.white,
              fontWeight: FontWeight.w800,
              fontSize: 16,
              letterSpacing: 1)),
      actions: [
        IconButton(
          icon: const Icon(Icons.refresh_rounded, color: Colors.white),
          tooltip: 'Segarkan',
          onPressed: () {
            context.read<AttendanceBloc>().add(FetchHomeDataRequested());
          },
        ),
      ],
    );
  }

  Widget _buildTimerCard(bool isClockedIn, String duration,
      Attendance? lastAttendance, bool isCompleted) {
    final bool showCheckIn = isClockedIn || isCompleted;
    final bool showCheckOut = isCompleted;

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.grayBorder),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 20,
              offset: const Offset(0, 4))
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('CHECK IN',
                      style: TextStyle(
                          fontSize: 10,
                          fontWeight: FontWeight.w900,
                          color: AppColors.textTertiary,
                          letterSpacing: 1)),
                  const SizedBox(height: 4),
                  Text(
                    showCheckIn
                        ? DateFormat('HH:mm').format(lastAttendance!.checkIn)
                        : '--:--',
                    style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.w900,
                        color: AppColors.textPrimary),
                  ),
                ],
              ),
              Container(
                height: 40,
                width: 1,
                color: AppColors.grayBorder,
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  const Text('CHECK OUT',
                      style: TextStyle(
                          fontSize: 10,
                          fontWeight: FontWeight.w900,
                          color: AppColors.textTertiary,
                          letterSpacing: 1)),
                  const SizedBox(height: 4),
                  Text(
                    showCheckOut
                        ? DateFormat('HH:mm').format(lastAttendance!.checkOut!)
                        : '--:--',
                    style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.w900,
                        color: AppColors.textPrimary),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 24),
          const Divider(height: 1),
          const SizedBox(height: 24),
          Column(
            children: [
              Text(
                _isBreakTime() && isClockedIn
                    ? 'WAKTU ISTIRAHAT (PAUSED)'
                    : (isClockedIn
                        ? 'DURASI KERJA HARI INI'
                        : 'STATUS TERAKHIR'),
                style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w800,
                    color: _isBreakTime() && isClockedIn
                        ? Colors.orange
                        : AppColors.textTertiary,
                    letterSpacing: 1),
              ),
              const SizedBox(height: 8),
              Text(
                isClockedIn
                    ? duration
                    : (isCompleted ? 'Sesi Terakhir Selesai' : 'Belum Absen'),
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 28,
                  fontWeight: FontWeight.w900,
                  color: _isBreakTime() && isClockedIn
                      ? Colors.orange
                      : (isClockedIn
                          ? const Color(0xFF10B981)
                          : AppColors.textPrimary),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAttendanceTypeSelector(
      bool isClockedIn, Attendance? lastAttendance) {
    // If user is already clocked in with out of town, display active badge
    if (isClockedIn && lastAttendance != null && lastAttendance.isOutOfTown) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: const Color(0xFF4F46E5).withOpacity(0.08),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFF4F46E5).withOpacity(0.2)),
        ),
        child: Row(
          children: [
            const Icon(Icons.flight_takeoff_rounded,
                color: Color(0xFF4F46E5), size: 22),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'SESI DINAS LUAR KOTA AKTIF',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                      color: const Color(0xFF4F46E5),
                      letterSpacing: 0.5,
                    ),
                  ),
                  if (lastAttendance.destination.isNotEmpty)
                    Text(
                      'Tujuan: ${lastAttendance.destination}',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: AppColors.textPrimary,
                      ),
                    ),
                ],
              ),
            ),
          ],
        ),
      );
    }

    if (isClockedIn) {
      return const SizedBox.shrink();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: InkWell(
                onTap: () {
                  setState(() {
                    _isOutOfTown = false;
                  });
                },
                borderRadius: BorderRadius.circular(14),
                child: Container(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  decoration: BoxDecoration(
                    color: !_isOutOfTown ? Colors.white : Colors.transparent,
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(
                      color: !_isOutOfTown
                          ? AppColors.primaryRed
                          : AppColors.grayBorder,
                      width: !_isOutOfTown ? 1.5 : 1,
                    ),
                    boxShadow: !_isOutOfTown
                        ? [
                            BoxShadow(
                              color: AppColors.primaryRed.withOpacity(0.08),
                              blurRadius: 10,
                              offset: const Offset(0, 4),
                            )
                          ]
                        : null,
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.business_rounded,
                        size: 18,
                        color: !_isOutOfTown
                            ? AppColors.primaryRed
                            : AppColors.textTertiary,
                      ),
                      const SizedBox(width: 8),
                      Text(
                        'Kantor / Reguler',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          fontWeight: FontWeight.w800,
                          color: !_isOutOfTown
                              ? AppColors.primaryRed
                              : AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: InkWell(
                onTap: () {
                  setState(() {
                    _isOutOfTown = true;
                  });
                },
                borderRadius: BorderRadius.circular(14),
                child: Container(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  decoration: BoxDecoration(
                    color: _isOutOfTown ? Colors.white : Colors.transparent,
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(
                      color: _isOutOfTown
                          ? const Color(0xFF4F46E5)
                          : AppColors.grayBorder,
                      width: _isOutOfTown ? 1.5 : 1,
                    ),
                    boxShadow: _isOutOfTown
                        ? [
                            BoxShadow(
                              color: const Color(0xFF4F46E5).withOpacity(0.1),
                              blurRadius: 10,
                              offset: const Offset(0, 4),
                            )
                          ]
                        : null,
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.flight_takeoff_rounded,
                        size: 18,
                        color: _isOutOfTown
                            ? const Color(0xFF4F46E5)
                            : AppColors.textTertiary,
                      ),
                      const SizedBox(width: 8),
                      Text(
                        'Dinas Luar Kota',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          fontWeight: FontWeight.w800,
                          color: _isOutOfTown
                              ? const Color(0xFF4F46E5)
                              : AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
        if (_isOutOfTown) ...[
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              border:
                  Border.all(color: const Color(0xFF4F46E5).withOpacity(0.3)),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.02),
                  blurRadius: 10,
                  offset: const Offset(0, 2),
                )
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(Icons.edit_location_alt_rounded,
                        color: Color(0xFF4F46E5), size: 16),
                    const SizedBox(width: 6),
                    Text(
                      'KOTA / TEMPAT TUJUAN DINAS',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                        color: const Color(0xFF4F46E5),
                        letterSpacing: 0.5,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                TextField(
                  controller: _destinationController,
                  decoration: InputDecoration(
                    hintText: 'Contoh: Surabaya - Kunjungan Proyek',
                    hintStyle: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      color: AppColors.textTertiary,
                    ),
                    isDense: true,
                    contentPadding: const EdgeInsets.symmetric(
                        horizontal: 12, vertical: 10),
                    filled: true,
                    fillColor: const Color(0xFF4F46E5).withOpacity(0.04),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(10),
                      borderSide: BorderSide.none,
                    ),
                  ),
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildMainActionButton(
      bool isClockedIn, bool isCompleted, Attendance? lastAttendance) {
    // If completed (already clocked out once today), we allow another Clock In for a new session
    final color = isClockedIn ? AppColors.primaryRed : const Color(0xFF10B981);

    return Center(
      child: InkWell(
        onTap: () {
          final isOut =
              (isClockedIn && lastAttendance != null && lastAttendance.isOutOfTown)
                  ? true
                  : _isOutOfTown;
          final dest = (isClockedIn &&
                  lastAttendance != null &&
                  lastAttendance.destination.isNotEmpty)
              ? lastAttendance.destination
              : _destinationController.text.trim();

          if (isOut && !isClockedIn && dest.isEmpty) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: const Text(
                    'Harap masukkan kota atau tempat tujuan dinas luar kota terlebih dahulu.'),
                backgroundColor: AppColors.primaryRed,
                behavior: SnackBarBehavior.floating,
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10)),
              ),
            );
            return;
          }

          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => FaceVerificationPage(
                isClockIn: !isClockedIn,
                isOutOfTown: isOut,
                destination: dest,
              ),
            ),
          );
        },
        borderRadius: BorderRadius.circular(100),
        child: Container(
          width: 180,
          height: 180,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: Colors.white,
            border: Border.all(color: color.withOpacity(0.1), width: 10),
            boxShadow: [
              BoxShadow(
                color: color.withOpacity(0.2),
                blurRadius: 40,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                isClockedIn ? Icons.logout_rounded : Icons.login_rounded,
                size: 48,
                color: color,
              ),
              const SizedBox(height: 12),
              Text(
                isClockedIn
                    ? 'CLOCK OUT'
                    : (isCompleted ? 'NEW CLOCK IN' : 'CLOCK IN'),
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w900,
                  color: AppColors.textPrimary,
                  letterSpacing: 1,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildStatusIndicators() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        _buildSmallIndicator(Icons.location_on_rounded, 'Location Ready',
            const Color(0xFF10B981)),
        const SizedBox(width: 16),
        _buildSmallIndicator(
            Icons.face_rounded, 'Biometric Active', const Color(0xFF10B981)),
      ],
    );
  }

  Widget _buildSmallIndicator(IconData icon, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: color.withOpacity(0.08),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: color),
          const SizedBox(width: 8),
          Text(
            label,
            style: TextStyle(
                fontSize: 10, fontWeight: FontWeight.w800, color: color),
          ),
        ],
      ),
    );
  }

  Widget _buildDailySummary(AttendanceState state) {
    List<Attendance> history = [];
    if (state is HomeDataLoaded) {
      history = state.history
          .where((a) => a.checkIn.day == DateTime.now().day)
          .toList();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('AKTIVITAS HARI INI',
            style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: AppColors.textTertiary,
                letterSpacing: 1.5)),
        const SizedBox(height: 16),
        if (history.isEmpty)
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: AppColors.grayBorder),
            ),
            child: const Center(
              child: Text(
                'Belum ada aktivitas hari ini',
                style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textTertiary),
              ),
            ),
          )
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: history.length,
            itemBuilder: (context, index) {
              final att = history[index];
              return Container(
                margin: const EdgeInsets.only(bottom: 12),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: AppColors.grayBorder),
                ),
                child: Row(
                  children: [
                    _buildLogBadge(att.checkOut == null),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            att.checkOut == null
                                ? 'Absen Masuk'
                                : 'Absen Keluar',
                            style: const TextStyle(
                                fontWeight: FontWeight.w800,
                                fontSize: 14,
                                color: AppColors.textPrimary),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            DateFormat('HH:mm')
                                .format(att.checkOut ?? att.checkIn),
                            style: const TextStyle(
                                fontSize: 12,
                                color: AppColors.textSecondary,
                                fontWeight: FontWeight.w600),
                          ),
                        ],
                      ),
                    ),
                    const Icon(Icons.check_circle_rounded,
                        color: Color(0xFF10B981), size: 20),
                  ],
                ),
              );
            },
          ),
      ],
    );
  }

  Widget _buildLogBadge(bool isCheckIn) {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: (isCheckIn ? const Color(0xFF10B981) : AppColors.primaryRed)
            .withOpacity(0.08),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Icon(
        isCheckIn ? Icons.login_rounded : Icons.logout_rounded,
        color: isCheckIn ? const Color(0xFF10B981) : AppColors.primaryRed,
        size: 20,
      ),
    );
  }

  bool _isBreakTime() {
    final now = DateTime.now();
    return now.hour == 12; // 12:00 to 12:59
  }

  String _calculateWorkDuration(DateTime? checkIn) {
    if (checkIn == null) return "00:00:00";
    final now = DateTime.now();
    Duration duration = now.difference(checkIn);

    final breakStart = DateTime(now.year, now.month, now.day, 12, 0);
    final breakEnd = DateTime(now.year, now.month, now.day, 13, 0);

    if (checkIn.isBefore(breakStart) && now.isAfter(breakStart)) {
      if (now.isBefore(breakEnd)) {
        duration -= now.difference(breakStart);
      } else {
        duration -= const Duration(hours: 1);
      }
    } else if (checkIn.isAfter(breakStart) && checkIn.isBefore(breakEnd)) {
      if (now.isBefore(breakEnd)) {
        duration -= now.difference(checkIn);
      } else {
        duration -= breakEnd.difference(checkIn);
      }
    }

    final hours = duration.inHours.toString().padLeft(2, '0');
    final minutes = (duration.inMinutes % 60).toString().padLeft(2, '0');
    final seconds = (duration.inSeconds % 60).toString().padLeft(2, '0');
    return "$hours:$minutes:$seconds";
  }

  Widget _buildLoadingState() {
    return Shimmer.fromColors(
      baseColor: AppColors.grayLight,
      highlightColor: Colors.white,
      child: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const SizedBox(height: 60),
          Center(
              child: Container(
                  width: 200,
                  height: 40,
                  decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(12)))),
          const SizedBox(height: 32),
          Container(
              width: double.infinity,
              height: 200,
              decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(24))),
          const SizedBox(height: 32),
          Center(
              child: Container(
                  width: 180,
                  height: 180,
                  decoration: const BoxDecoration(
                      color: Colors.white, shape: BoxShape.circle))),
        ],
      ),
    );
  }
}
