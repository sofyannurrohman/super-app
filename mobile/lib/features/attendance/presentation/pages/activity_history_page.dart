import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_bloc.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_event.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_state.dart';
import 'package:hris_app/features/attendance/domain/entities/attendance.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:intl/intl.dart';

class ActivityHistoryPage extends StatefulWidget {
  const ActivityHistoryPage({super.key});

  @override
  State<ActivityHistoryPage> createState() => _ActivityHistoryPageState();
}

class _ActivityHistoryPageState extends State<ActivityHistoryPage> {
  DateTime _selectedDate = DateTime.now();

  @override
  void initState() {
    super.initState();
    _fetchHistory();
  }

  void _fetchHistory() {
    final startOfMonth = DateTime(_selectedDate.year, _selectedDate.month, 1);
    final endOfMonth =
        DateTime(_selectedDate.year, _selectedDate.month + 1, 0, 23, 59, 59);

    context.read<AttendanceBloc>().add(FetchHistoryRequested(
          startDate: DateFormat('yyyy-MM-dd').format(startOfMonth),
          endDate: DateFormat('yyyy-MM-dd').format(endOfMonth),
        ));
  }

  Future<void> _selectMonth(BuildContext context) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: _selectedDate,
      firstDate: DateTime(2020),
      lastDate: DateTime.now(),
      builder: (context, child) {
        return Theme(
          data: Theme.of(context).copyWith(
            colorScheme: const ColorScheme.light(
              primary: AppColors.primaryRed,
              onPrimary: Colors.white,
              onSurface: AppColors.textPrimary,
            ),
          ),
          child: child!,
        );
      },
      initialDatePickerMode: DatePickerMode.year,
    );

    if (picked != null && picked != _selectedDate) {
      setState(() {
        _selectedDate = DateTime(picked.year, picked.month, 1);
      });
      _fetchHistory();
    }
  }

  // Dummy data to match mockup perfectly
  final List<Map<String, dynamic>> _dummyHistory = [
    {
      'date': '12',
      'day': 'Thu',
      'month': 'Oct',
      'isLate': false,
      'isOvertime': false,
      'overtimeHours': null,
      'checkIn': '08:00 AM',
      'checkOut': '17:00 PM',
      'totalTime': '8h 0m',
    },
    {
      'date': '11',
      'day': 'Wed',
      'month': 'Oct',
      'isLate': true,
      'isOvertime': false,
      'overtimeHours': null,
      'checkIn': '08:15 AM',
      'checkOut': '17:00 PM',
      'totalTime': '7h 45m',
    },
    {
      'date': '10',
      'day': 'Tue',
      'month': 'Oct',
      'isLate': false,
      'isOvertime': true,
      'overtimeHours': '+2h',
      'checkIn': '08:00 AM',
      'checkOut': '19:00 PM',
      'totalTime': '10h 0m',
    },
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      body: SafeArea(
        child: Container(
          decoration: const BoxDecoration(color: AppColors.backgroundAlt),
          child: BlocBuilder<AttendanceBloc, AttendanceState>(
            builder: (context, state) {
              return Column(
                children: [
                  _buildHeader(context, state),
                  Expanded(
                    child: _buildBody(context, state),
                  ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _buildBody(BuildContext context, AttendanceState state) {
    if (state is AttendanceLoading) {
      return const Center(
          child: CircularProgressIndicator(color: AppColors.primaryRed));
    }

    if (state is AttendanceHistoryLoaded || state is HomeDataLoaded) {
      final List<Attendance> history = (state is AttendanceHistoryLoaded)
          ? state.history
          : (state as HomeDataLoaded).history;

      return RefreshIndicator(
        onRefresh: () async {
          _fetchHistory();
        },
        child: history.isEmpty
            ? ListView(
                physics: const AlwaysScrollableScrollPhysics(),
                children: [
                  SizedBox(
                      height: MediaQuery.of(context).size.height * 0.2),
                  Center(
                    child: Column(
                      children: [
                        Icon(Icons.history_toggle_off_rounded,
                            size: 80,
                            color: AppColors.textTertiary.withOpacity(0.2)),
                        const SizedBox(height: 16),
                        Text('Belum ada riwayat',
                            style: GoogleFonts.plusJakartaSans(
                                color: AppColors.textTertiary,
                                fontWeight: FontWeight.w700)),
                      ],
                    ),
                  ),
                ],
              )
            : ListView.builder(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.fromLTRB(24, 8, 24, 100),
                itemCount: history.length,
                itemBuilder: (context, index) =>
                    _buildModernHistoryCard(history[index]),
              ),
      );
    }

    return const Center(
        child: CircularProgressIndicator(color: AppColors.primaryRed));
  }

  Widget _buildHeader(BuildContext context, AttendanceState state) {
    int totalCheckIn = 0;
    int totalLeave = 0;

    if (state is AttendanceHistoryLoaded || state is HomeDataLoaded) {
      final List<Attendance> history = (state is AttendanceHistoryLoaded)
          ? state.history
          : (state as HomeDataLoaded).history;
          
      for (var record in history) {
        final status = record.status.toLowerCase();
        if (status == 'izin' || status == 'cuti' || status == 'leave') {
          totalLeave++;
        } else {
          totalCheckIn++;
        }
      }
    }

    return Container(
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 24),
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          colors: AppColors.primaryGradient,
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.only(
          bottomLeft: Radius.circular(32),
          bottomRight: Radius.circular(32),
        ),
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              GestureDetector(
                onTap: () => Navigator.pop(context),
                child: Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.2),
                      borderRadius: BorderRadius.circular(14)),
                  child: const Icon(Icons.arrow_back_ios_new_rounded,
                      color: Colors.white, size: 18),
                ),
              ),
              Text(
                'RIWAYAT AKTIVITAS',
                style: GoogleFonts.plusJakartaSans(
                    color: Colors.white,
                    fontWeight: FontWeight.w800,
                    fontSize: 13,
                    letterSpacing: 1.5),
              ),
              const SizedBox(width: 40), // Spacing for alignment
            ],
          ),
          const SizedBox(height: 24),
          Row(
            children: [
              Expanded(
                flex: 12,
                child: _buildMonthSelector(context),
              ),
              const SizedBox(width: 16),
              Expanded(
                flex: 11,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    _buildStatItem('Masuk', totalCheckIn.toString(), Icons.check_circle_outline_rounded),
                    _buildStatItem('Izin/Cuti', totalLeave.toString(), Icons.event_busy_rounded),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem(String label, String value, IconData icon) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, color: Colors.white.withOpacity(0.9), size: 12),
            const SizedBox(width: 4),
            Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                color: Colors.white.withOpacity(0.9),
                fontSize: 10,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Text(
          value,
          style: GoogleFonts.plusJakartaSans(
            color: Colors.white,
            fontSize: 22,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }

  Widget _buildMonthSelector(BuildContext context) {
    return GestureDetector(
      onTap: () => _selectMonth(context),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
                color: Colors.black.withOpacity(0.1),
                blurRadius: 10,
                offset: const Offset(0, 4))
          ],
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Expanded(
              child: Row(
                children: [
                  const Icon(Icons.calendar_month_rounded,
                      color: AppColors.primaryRed, size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                        DateFormat('MMMM yyyy', 'id_ID')
                            .format(_selectedDate)
                            .toUpperCase(),
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 11,
                            fontWeight: FontWeight.w800,
                            color: AppColors.textPrimary,
                            letterSpacing: 0.5),
                        overflow: TextOverflow.ellipsis),
                  ),
                ],
              ),
            ),
            const Icon(Icons.keyboard_arrow_down_rounded,
                color: AppColors.textTertiary, size: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildModernHistoryCard(Attendance record) {
    final dayStr =
        DateFormat('EEE', 'id_ID').format(record.checkIn).toUpperCase();
    final dateStr = DateFormat('dd').format(record.checkIn);
    final monthStr =
        DateFormat('MMM', 'id_ID').format(record.checkIn).toUpperCase();
    final timeStr =
        '${DateFormat('HH:mm').format(record.checkIn)} - ${record.checkOut != null ? DateFormat('HH:mm').format(record.checkOut!) : '--:--'}';

    // Duration calculation
    int hours = record.workDuration ~/ 60;
    int minutes = record.workDuration % 60;
    String durationStr = hours > 0 ? '${hours}j ${minutes}m' : '${minutes}m';
    if (record.workDuration == 0 && record.checkOut == null) durationStr = '--';

    bool isLate = record.status.toLowerCase() == 'late';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(28),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 20,
              offset: const Offset(0, 8)),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              color: AppColors.grayLight.withOpacity(0.5),
              borderRadius: BorderRadius.circular(16),
            ),
            child: Column(
              children: [
                Text(dayStr,
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textTertiary)),
                Text(dateStr,
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 20,
                        fontWeight: FontWeight.w900,
                        color: AppColors.textPrimary)),
                Text(monthStr,
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textTertiary)),
              ],
            ),
          ),
          const SizedBox(width: 20),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    _buildBadge(isLate ? 'TERLAMBAT' : 'TEPAT WAKTU',
                        isLate ? AppColors.primaryRed : AppColors.success),
                    const SizedBox(width: 8),
                    _buildBadge('REGULAR', AppColors.textTertiary),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    const Icon(Icons.access_time_filled_rounded,
                        size: 14, color: AppColors.textTertiary),
                    const SizedBox(width: 8),
                    Text(timeStr,
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 14,
                            color: AppColors.textPrimary,
                            fontWeight: FontWeight.w700)),
                  ],
                ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(durationStr,
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 16,
                      fontWeight: FontWeight.w900,
                      color: AppColors.textPrimary)),
              Text('TOTAL',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 9,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textTertiary,
                      letterSpacing: 1)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBadge(String text, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(text,
          style: GoogleFonts.plusJakartaSans(
              fontSize: 9,
              fontWeight: FontWeight.w800,
              color: color,
              letterSpacing: 0.5)),
    );
  }
}
