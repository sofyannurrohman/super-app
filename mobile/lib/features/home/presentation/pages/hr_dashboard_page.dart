import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/features/approval/presentation/bloc/approval_bloc.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/reimbursement/domain/entities/reimbursement.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/announcement/presentation/pages/announcement_management_page.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:shimmer/shimmer.dart';

class HrDashboardPage extends StatefulWidget {
  const HrDashboardPage({super.key});

  @override
  State<HrDashboardPage> createState() => _HrDashboardPageState();
}

class _HrDashboardPageState extends State<HrDashboardPage> {
  final PageController _pageController = PageController();
  int _currentPage = 0;

  // View state
  String _currentView = 'monitoring'; // 'monitoring', 'cuti', 'reimburse'

  // Attendance Monitoring State
  bool _isAttendanceLoading = true;
  List<dynamic> _attendanceList = [];
  List<Map<String, dynamic>> _branchesList = [];
  String _selectedBranchId = 'all';
  String _filterType =
      'month'; // Default to 'month' to ensure stats are populated with the current month's data
  DateTime _selectedDate = DateTime.now();

  // Pagination State
  int _attendancePage = 1;
  bool _hasMoreAttendance = true;
  bool _isLoadingMoreAttendance = false;
  final int _attendanceLimit = 15;

  // Approvals State
  // Removed _approvalSubTab usage, using _currentView instead.

  @override
  void initState() {
    super.initState();

    // Set default branch for branch-scoped users
    try {
      final authState = context.read<AuthBloc>().state;
      if (authState is Authenticated) {
        final profile = authState.userProfile;
        final branch = profile?['branch'] ?? profile?['Branch'];
        final userRole =
            (profile?['user']?['role'] ?? profile?['user']?['Role'] ?? '')
                .toString()
                .toLowerCase();

        if (userRole != 'superadmin' && branch != null) {
          final branchId = (branch['id'] ?? branch['ID'] ?? '').toString();
          if (branchId.isNotEmpty) {
            _selectedBranchId = branchId;
          }
        }
      }
    } catch (e) {
      debugPrint("Error initializing default branch: $e");
    }

    _fetchBranches();
    _fetchAttendance();
    context.read<ApprovalBloc>().add(FetchPendingApprovalsRequested());
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  Future<void> _fetchBranches() async {
    final result = await di.sl<AuthRepository>().getBranches();
    result.fold(
      (failure) => debugPrint("Error fetching branches: ${failure.message}"),
      (branches) {
        if (mounted) {
          setState(() {
            final authState = context.read<AuthBloc>().state;
            if (authState is Authenticated) {
              final profile = authState.userProfile;
              final userRole =
                  (profile?['user']?['role'] ?? profile?['user']?['Role'] ?? '')
                      .toString()
                      .toLowerCase();
              final branch = profile?['branch'] ?? profile?['Branch'];
              if (userRole != 'superadmin' && branch != null) {
                final branchId = (branch['id'] ?? branch['ID'] ?? '')
                    .toString();
                _branchesList = branches
                    .where(
                      (b) => (b['id'] ?? b['ID'] ?? '').toString() == branchId,
                    )
                    .toList()
                    .cast<Map<String, dynamic>>();
                return;
              }
            }
            _branchesList = branches;
          });
        }
      },
    );
  }

  Future<void> _fetchAttendance({bool loadMore = false}) async {
    if (!mounted) return;

    if (loadMore) {
      if (!_hasMoreAttendance || _isLoadingMoreAttendance) return;
      setState(() => _isLoadingMoreAttendance = true);
      _attendancePage++;
    } else {
      setState(() {
        _isAttendanceLoading = true;
        _attendancePage = 1;
        _hasMoreAttendance = true;
      });
    }

    try {
      final apiClient = di.sl<ApiClient>();

      String monthParam = '';
      if (_filterType == 'daily') {
        monthParam = DateFormat('yyyy-MM-dd').format(_selectedDate);
      } else if (_filterType == 'month') {
        monthParam = DateFormat('yyyy-MM').format(_selectedDate);
      } else if (_filterType == 'year') {
        monthParam = DateFormat('yyyy').format(_selectedDate);
      }

      final queryParams = {
        'limit': _attendanceLimit,
        'page': _attendancePage,
        if (_selectedBranchId != 'all') 'branch_id': _selectedBranchId,
        'month': monthParam,
      };

      final response = await apiClient.client.get(
        'attendance/all',
        queryParameters: queryParams,
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        if (mounted) {
          setState(() {
            if (loadMore) {
              _attendanceList.addAll(data);
              _isLoadingMoreAttendance = false;
            } else {
              _attendanceList = data;
              _isAttendanceLoading = false;
            }
            if (data.length < _attendanceLimit) {
              _hasMoreAttendance = false;
            }
          });
        }
      } else {
        debugPrint(
          'Error fetching HR attendance: status code ${response.statusCode}, response: ${response.data}',
        );
        if (mounted) {
          setState(() {
            if (loadMore) {
              _isLoadingMoreAttendance = false;
            } else {
              _isAttendanceLoading = false;
            }
          });
        }
      }
    } catch (e) {
      debugPrint('Error fetching HR attendance: $e');
      if (mounted) {
        setState(() {
          if (loadMore) {
            _isLoadingMoreAttendance = false;
          } else {
            _isAttendanceLoading = false;
          }
        });
      }
    }
  }

  // Helper properties for statistics
  int get _presentCount {
    return _attendanceList.where((item) {
      final s = (item['status'] ?? '').toString().toUpperCase();
      return s == 'ON_TIME' || s == 'PRESENT';
    }).length;
  }

  int get _lateCount {
    return _attendanceList.where((item) {
      final s = (item['status'] ?? '').toString().toUpperCase();
      return s == 'LATE';
    }).length;
  }

  int get _leaveCount {
    return _attendanceList.where((item) {
      final s = (item['status'] ?? '').toString().toUpperCase();
      return ['SICK', 'LEAVE', 'PERMISSION'].contains(s);
    }).length;
  }

  int get _absentCount {
    return _attendanceList.where((item) {
      final s = (item['status'] ?? '').toString().toUpperCase();
      return s == 'ABSENT';
    }).length;
  }

  // Format Helper Methods
  String _formatTime(String? timeStr) {
    if (timeStr == null || timeStr.isEmpty) return '-';
    try {
      final dt = DateTime.parse(timeStr).toLocal();
      return DateFormat('HH:mm').format(dt);
    } catch (_) {
      // Handle fallback format HH:mm:ss if direct parsing fails
      if (timeStr.contains(':')) {
        final parts = timeStr.split(':');
        if (parts.length >= 2) return '${parts[0]}:${parts[1]}';
      }
      return '-';
    }
  }

  String _formatDate(String? dateStr) {
    if (dateStr == null || dateStr.isEmpty) return '-';
    try {
      final dt = DateTime.parse(dateStr).toLocal();
      return DateFormat('dd MMM yyyy').format(dt);
    } catch (_) {
      return '-';
    }
  }

  String _getDuration(String? checkIn, String? checkOut) {
    if (checkIn == null || checkOut == null) return '-';
    try {
      final inTime = DateTime.parse(checkIn);
      final outTime = DateTime.parse(checkOut);
      final diff = outTime.difference(inTime);
      final hours = diff.inHours;
      final minutes = diff.inMinutes % 60;
      return '${hours}j ${minutes}m';
    } catch (_) {
      return '-';
    }
  }

  String _buildImageUrl(String path) {
    if (path.startsWith('http')) return path;
    final originUrl = Uri.parse(AppConstants.baseUrl).origin;
    return '$originUrl${path.startsWith('/') ? path : '/$path'}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC), // Premium Ice Blue Background
      appBar: AppBar(
        title: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'HR PERFORMANCE',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900,
                color: Colors.white,
                fontSize: 14,
                letterSpacing: 1,
              ),
            ),
            Text(
              'Perusahaan Demo',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w700,
                color: AppColors.blueHRLight,
                fontSize: 8,
                letterSpacing: 1,
              ),
            ),
          ],
        ),
        backgroundColor: AppColors.blueHRDark,
        centerTitle: true,
        elevation: 0,
        leadingWidth: 100,
        automaticallyImplyLeading: false,
        leading: Center(
          child: Padding(
            padding: const EdgeInsets.only(left: 12),
            child: TextButton(
              onPressed: () => Navigator.pop(context),
              style: TextButton.styleFrom(
                backgroundColor: Colors.white.withOpacity(0.1),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              ),
              child: Text(
                'KEMBALI',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 11,
                  color: Colors.white,
                ),
              ),
            ),
          ),
        ),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await _fetchBranches();
          await _fetchAttendance();
          if (mounted) {
            context.read<ApprovalBloc>().add(FetchPendingApprovalsRequested());
          }
        },
        child: SingleChildScrollView(
          physics: const BouncingScrollPhysics(),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 10),
              _buildSwipeableHeader(),
              const SizedBox(height: 24),
              _buildQuickAccess(),
              const SizedBox(height: 32),
              if (_currentView == 'monitoring') ...[
                _buildMonitoringContent(),
              ] else if (_currentView == 'cuti') ...[
                _buildApprovalsContent('cuti'),
              ],
              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }

  // ==========================================
  // HEADER CARDS
  // ==========================================
  Widget _buildSwipeableHeader() {
    return Column(
      children: [
        SizedBox(
          height: 220,
          child: PageView(
            controller: _pageController,
            physics: const BouncingScrollPhysics(),
            pageSnapping: true,
            onPageChanged: (v) => setState(() => _currentPage = v),
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 2),
                child: _buildAttendanceKpiCard(),
              ),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 2),
                child: _buildApprovalKpiCard(),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(
            2,
            (index) => Container(
              margin: const EdgeInsets.symmetric(horizontal: 4),
              width: _currentPage == index ? 24 : 8,
              height: 8,
              decoration: BoxDecoration(
                color: _currentPage == index
                    ? AppColors.blueHR
                    : const Color(0xFFCBD5E1),
                borderRadius: BorderRadius.circular(4),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildAttendanceKpiCard() {
    final int total = _attendanceList.length;
    final int present = _presentCount;
    final int late = _lateCount;
    final int leave = _leaveCount;
    final int absent = _absentCount;

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: AppColors.blueHRDark.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
        border: Border.all(
          color: AppColors.blueHR.withOpacity(0.1),
          width: 1.5,
        ),
      ),
      child: Stack(
        children: [
          Positioned(
            right: -10,
            top: -10,
            child: Icon(
              Icons.groups_rounded,
              size: 100,
              color: AppColors.blueHRLight.withOpacity(0.04),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.blueHRLight.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Icon(
                            Icons.calendar_today_rounded,
                            color: AppColors.blueHR,
                            size: 20,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Text(
                          'STATISTIK PRESENSI',
                          style: GoogleFonts.outfit(
                            color: AppColors.blueHRDark,
                            fontWeight: FontWeight.w800,
                            fontSize: 11,
                            letterSpacing: 1.5,
                          ),
                        ),
                      ],
                    ),
                    Text(
                      DateFormat('dd MMM yyyy').format(DateTime.now()),
                      style: GoogleFonts.outfit(
                        color: AppColors.blueHR,
                        fontWeight: FontWeight.w700,
                        fontSize: 11,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: _buildStatItem(
                        'Total',
                        total.toString(),
                        AppColors.blueHRDark,
                      ),
                    ),
                    Expanded(
                      child: _buildStatItem(
                        'Hadir',
                        present.toString(),
                        Colors.green,
                      ),
                    ),
                    Expanded(
                      child: _buildStatItem(
                        'Telat',
                        late.toString(),
                        Colors.orange,
                      ),
                    ),
                    Expanded(
                      child: _buildStatItem(
                        'Cuti',
                        leave.toString(),
                        Colors.blue,
                      ),
                    ),
                    Expanded(
                      child: _buildStatItem(
                        'Alfa',
                        absent.toString(),
                        Colors.red,
                      ),
                    ),
                  ],
                ),
                const Spacer(),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  decoration: BoxDecoration(
                    color: AppColors.blueHR,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Center(
                    child: Text(
                      'DASHBOARD PRESENSI',
                      style: GoogleFonts.outfit(
                        color: Colors.white,
                        fontWeight: FontWeight.w800,
                        fontSize: 10,
                        letterSpacing: 1,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildApprovalKpiCard() {
    return BlocBuilder<ApprovalBloc, ApprovalState>(
      builder: (context, state) {
        int pendingLeavesCount = 0;
        if (state is ApprovalDataLoaded) {
          pendingLeavesCount = state.pendingLeaves.length;
        }

        return Container(
          width: double.infinity,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(24),
            boxShadow: [
              BoxShadow(
                color: AppColors.blueHRDark.withOpacity(0.08),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
            border: Border.all(
              color: AppColors.blueHR.withOpacity(0.1),
              width: 1.5,
            ),
          ),
          child: Stack(
            children: [
              Positioned(
                right: -10,
                top: -10,
                child: Icon(
                  Icons.assignment_turned_in_rounded,
                  size: 100,
                  color: AppColors.blueHRLight.withOpacity(0.04),
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.blueHRLight.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Icon(
                            Icons.pending_actions_rounded,
                            color: AppColors.blueHR,
                            size: 20,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Text(
                          'TUGAS PERSETUJUAN',
                          style: GoogleFonts.outfit(
                            color: AppColors.blueHRDark,
                            fontWeight: FontWeight.w800,
                            fontSize: 11,
                            letterSpacing: 1.5,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Expanded(
                          child: Column(
                            children: [
                              Text(
                                pendingLeavesCount.toString(),
                                style: GoogleFonts.outfit(
                                  fontSize: 32,
                                  fontWeight: FontWeight.w900,
                                  color: Colors.blue,
                                ),
                              ),
                              Text(
                                'CUTI & IZIN',
                                style: GoogleFonts.outfit(
                                  fontSize: 10,
                                  fontWeight: FontWeight.w800,
                                  color: Colors.blueGrey,
                                  letterSpacing: 1,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const Spacer(),
                    Text(
                      'Ada tugas approval yang menunggu tindakan Anda',
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        color: Colors.blueGrey,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildStatItem(String label, String value, Color color) {
    return Column(
      children: [
        FittedBox(
          fit: BoxFit.scaleDown,
          child: Text(
            value,
            style: GoogleFonts.outfit(
              fontSize: 20,
              fontWeight: FontWeight.w900,
              color: color,
            ),
          ),
        ),
        Text(
          label.toUpperCase(),
          style: GoogleFonts.outfit(
            fontSize: 9,
            fontWeight: FontWeight.w800,
            color: Colors.blueGrey,
            letterSpacing: 1,
          ),
        ),
      ],
    );
  }

  // ==========================================
  // QUICK ACCESS
  // ==========================================
  Widget _buildQuickAccess() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'AKSES OPERASIONAL',
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF64748B),
                letterSpacing: 1.5,
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),
        GridView.count(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          crossAxisCount: 3,
          mainAxisSpacing: 12,
          crossAxisSpacing: 12,
          childAspectRatio: 0.85,
          children: [
            _buildActionItem(
              Icons.history_rounded,
              'Log\nPresensi',
              const Color(0xFF0D9488),
              'Riwayat',
              'monitoring',
            ),
            _buildActionItem(
              Icons.event_note_rounded,
              'Cuti &\nIzin',
              const Color(0xFF3B82F6),
              'Persetujuan',
              'cuti',
            ),
            _buildActionItem(
              Icons.campaign_rounded,
              'Pengumuman',
              const Color(0xFF8B5CF6),
              'Operasional',
              'pengumuman',
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const AnnouncementManagementPage(),
                  ),
                );
              },
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildActionItem(
    IconData icon,
    String label,
    Color color,
    String subLabel,
    String viewId, {
    VoidCallback? onTap,
  }) {
    final bool isSelected = _currentView == viewId;
    return Container(
      decoration: BoxDecoration(
        color: isSelected ? color.withOpacity(0.05) : Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
          color: isSelected ? color.withOpacity(0.5) : Colors.transparent,
          width: 2,
        ),
        boxShadow: [
          BoxShadow(
            color: color.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap:
              onTap ??
              () {
                setState(() => _currentView = viewId);
              },
          borderRadius: BorderRadius.circular(24),
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        color.withOpacity(0.15),
                        color.withOpacity(0.05),
                      ],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(icon, color: color, size: 24),
                ),
                const SizedBox(height: 10),
                Text(
                  label,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B),
                  ),
                ),
                Text(
                  subLabel,
                  textAlign: TextAlign.center,
                  style: GoogleFonts.outfit(
                    fontSize: 9,
                    fontWeight: FontWeight.w500,
                    color: const Color(0xFF94A3B8),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // ==========================================
  // MONITORING PRESENSI VIEW
  // ==========================================
  Widget _buildMonitoringContent() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFilterSection(),
        const SizedBox(height: 24),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'RIWAYAT PRESENSI KARYAWAN',
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF64748B),
                letterSpacing: 1.5,
              ),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: AppColors.blueHR.withOpacity(0.1),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                '${_attendanceList.length} Total Logs',
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w900,
                  color: AppColors.blueHRDark,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        _isAttendanceLoading
            ? _buildSkeleton()
            : _attendanceList.isEmpty
            ? _buildEmptyState('Belum ada log presensi untuk filter ini.')
            : Column(
                children: [
                  ListView.builder(
                    shrinkWrap: true,
                    physics: const NeverScrollableScrollPhysics(),
                    itemCount: _attendanceList.length,
                    itemBuilder: (context, index) =>
                        _buildAttendanceCard(_attendanceList[index]),
                  ),
                  if (_hasMoreAttendance)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 16.0),
                      child: _isLoadingMoreAttendance
                          ? const CircularProgressIndicator(
                              color: AppColors.blueHR,
                            )
                          : TextButton(
                              onPressed: () => _fetchAttendance(loadMore: true),
                              style: TextButton.styleFrom(
                                foregroundColor: AppColors.blueHR,
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 24,
                                  vertical: 12,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(16),
                                  side: BorderSide(
                                    color: AppColors.blueHR.withOpacity(0.5),
                                    width: 1,
                                  ),
                                ),
                              ),
                              child: Text(
                                'Muat Lebih Banyak',
                                style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ),
                    ),
                ],
              ),
      ],
    );
  }

  Widget _buildSkeleton() {
    return Shimmer.fromColors(
      baseColor: Colors.grey.shade300,
      highlightColor: Colors.grey.shade100,
      child: Column(
        children: List.generate(
          3,
          (index) => Container(
            margin: const EdgeInsets.only(bottom: 12),
            height: 120,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFilterSection() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF0F172A).withOpacity(0.03),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(
                Icons.filter_list_rounded,
                color: AppColors.blueHR,
                size: 20,
              ),
              const SizedBox(width: 8),
              Text(
                'FILTER DATA PRESENSI',
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                  color: const Color(0xFF1E293B),
                  letterSpacing: 1,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          // Branch Dropdown Selector
          DropdownButtonFormField<String>(
            initialValue: _selectedBranchId,
            decoration: InputDecoration(
              labelText: 'Pilih Cabang / Kantor',
              labelStyle: GoogleFonts.outfit(
                fontSize: 12,
                fontWeight: FontWeight.w700,
              ),
              filled: true,
              fillColor: const Color(0xFFF1F5F9),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide.none,
              ),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 12,
              ),
            ),
            items: [
              const DropdownMenuItem(value: 'all', child: Text('Semua Cabang')),
              if (_selectedBranchId != 'all' &&
                  !_branchesList.any(
                    (b) => (b['id'] ?? b['ID']).toString() == _selectedBranchId,
                  ))
                DropdownMenuItem(
                  value: _selectedBranchId,
                  child: const Text('Memuat Cabang...'),
                ),
              ..._branchesList
                  .map(
                    (branch) => DropdownMenuItem(
                      value: (branch['id'] ?? branch['ID'] ?? '').toString(),
                      child: Text(
                        (branch['name'] ?? branch['Name'] ?? 'Cabang')
                            .toString(),
                      ),
                    ),
                  )
                  .toSet(), // toSet().toList() prevents any duplicates if API returns them
            ],
            onChanged: (val) {
              if (val != null) {
                setState(() => _selectedBranchId = val);
                _fetchAttendance();
              }
            },
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              // Filter Type Selector (Harian, Bulanan, Tahunan)
              Expanded(
                child: DropdownButtonFormField<String>(
                  initialValue: _filterType,
                  decoration: InputDecoration(
                    labelText: 'Jenis Filter Waktu',
                    labelStyle: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                    filled: true,
                    fillColor: const Color(0xFFF1F5F9),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(16),
                      borderSide: BorderSide.none,
                    ),
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 12,
                    ),
                  ),
                  items: const [
                    DropdownMenuItem(value: 'daily', child: Text('Harian')),
                    DropdownMenuItem(value: 'month', child: Text('Bulanan')),
                    DropdownMenuItem(value: 'year', child: Text('Tahunan')),
                  ],
                  onChanged: (val) {
                    if (val != null) {
                      setState(() => _filterType = val);
                      _fetchAttendance();
                    }
                  },
                ),
              ),
              const SizedBox(width: 12),
              // Date picker trigger button
              Expanded(
                child: InkWell(
                  onTap: () async {
                    final picked = await showDatePicker(
                      context: context,
                      initialDate: _selectedDate,
                      firstDate: DateTime(2020),
                      lastDate: DateTime.now().add(const Duration(days: 365)),
                      builder: (context, child) {
                        return Theme(
                          data: Theme.of(context).copyWith(
                            colorScheme: const ColorScheme.light(
                              primary: AppColors.blueHR,
                              onPrimary: Colors.white,
                              onSurface: AppColors.blueHRDark,
                            ),
                          ),
                          child: child!,
                        );
                      },
                    );
                    if (picked != null) {
                      setState(() => _selectedDate = picked);
                      _fetchAttendance();
                    }
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 14,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Pilih Periode',
                          style: GoogleFonts.outfit(
                            fontSize: 9,
                            color: Colors.blueGrey,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                        const SizedBox(height: 2),
                        FittedBox(
                          fit: BoxFit.scaleDown,
                          child: Text(
                            _filterType == 'daily'
                                ? DateFormat(
                                    'dd MMM yyyy',
                                  ).format(_selectedDate)
                                : (_filterType == 'month'
                                      ? DateFormat(
                                          'MMMM yyyy',
                                        ).format(_selectedDate)
                                      : DateFormat(
                                          'yyyy',
                                        ).format(_selectedDate)),
                            style: GoogleFonts.outfit(
                              fontSize: 13,
                              fontWeight: FontWeight.w800,
                              color: AppColors.blueHRDark,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAttendanceCard(dynamic item) {
    final emp = item['employee'] ?? item['Employee'];
    final firstName = emp?['first_name'] ?? emp?['FirstName'] ?? 'Karyawan';
    final lastName = emp?['last_name'] ?? emp?['LastName'] ?? '';
    final fullName = '$firstName $lastName'.trim();
    final job = emp?['job_position'] ?? emp?['JobPosition'];
    final jobTitle = job?['title'] ?? job?['Title'] ?? '-';
    final branch = emp?['branch'] ?? emp?['Branch'];
    final branchName = branch?['name'] ?? branch?['Name'] ?? '-';

    final status = (item['status'] ?? '').toString().toUpperCase();
    final clockInTime = item['clock_in_time']?.toString();
    final clockOutTime = item['clock_out_time']?.toString();
    final notes = item['notes'] ?? item['Notes'];

    final clockInPhoto =
        item['clock_in_photo_url'] ?? item['ClockInPhotoUrl'] ?? '';
    final clockOutPhoto =
        item['clock_out_photo_url'] ?? item['ClockOutPhotoUrl'] ?? '';

    Color badgeColor;
    String statusLabel;
    if (status == 'ON_TIME' || status == 'PRESENT') {
      badgeColor = Colors.green;
      statusLabel = 'Tepat Waktu';
    } else if (status == 'LATE') {
      badgeColor = Colors.orange;
      statusLabel = 'Terlambat';
    } else if (['SICK', 'LEAVE', 'PERMISSION'].contains(status)) {
      badgeColor = Colors.blue;
      statusLabel = status == 'SICK'
          ? 'Sakit'
          : (status == 'LEAVE' ? 'Cuti' : 'Izin');
    } else {
      badgeColor = Colors.redAccent;
      statusLabel = 'Mangkir / Alfa';
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF0F172A).withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              // Avatar
              Container(
                width: 44,
                height: 44,
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    colors: [AppColors.blueHRLight, AppColors.blueHRDark],
                  ),
                  shape: BoxShape.circle,
                ),
                child: Center(
                  child: Text(
                    fullName.isNotEmpty ? fullName[0].toUpperCase() : 'K',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900,
                      color: Colors.white,
                      fontSize: 16,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      fullName,
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w800,
                        color: const Color(0xFF0F172A),
                        fontSize: 14,
                      ),
                    ),
                    Text(
                      '${jobTitle.toUpperCase()} • $branchName',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        color: Colors.blueGrey,
                        fontSize: 9,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ],
                ),
              ),
              // Status Badge
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: badgeColor.withOpacity(0.12),
                  borderRadius: BorderRadius.circular(100),
                ),
                child: Text(
                  statusLabel.toUpperCase(),
                  style: GoogleFonts.outfit(
                    fontSize: 8,
                    fontWeight: FontWeight.w900,
                    color: badgeColor,
                    letterSpacing: 0.5,
                  ),
                ),
              ),
            ],
          ),
          const Divider(height: 24, color: Color(0xFFF1F5F9)),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildTimeIndicator(
                'JAM MASUK',
                _formatTime(clockInTime),
                Colors.green,
              ),
              _buildTimeIndicator(
                'JAM KELUAR',
                _formatTime(clockOutTime),
                Colors.orange,
              ),
              _buildTimeIndicator(
                'DURASI',
                _getDuration(clockInTime, clockOutTime),
                AppColors.blueHR,
              ),
            ],
          ),
          // Photos evidence or notes if present
          if (clockInPhoto.isNotEmpty ||
              clockOutPhoto.isNotEmpty ||
              (notes != null && notes.toString().isNotEmpty)) ...[
            const Divider(height: 24, color: Color(0xFFF1F5F9)),
            Row(
              children: [
                if (clockInPhoto.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: _buildPhotoThumbnail('CLOCK-IN', clockInPhoto),
                  ),
                if (clockOutPhoto.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: _buildPhotoThumbnail('CLOCK-OUT', clockOutPhoto),
                  ),
                if (notes != null && notes.toString().isNotEmpty)
                  Expanded(
                    child: Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAF9),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Icon(
                            Icons.sticky_note_2_rounded,
                            size: 14,
                            color: Colors.blueGrey,
                          ),
                          const SizedBox(width: 6),
                          Expanded(
                            child: Text(
                              notes.toString(),
                              style: GoogleFonts.outfit(
                                fontSize: 10,
                                color: const Color(0xFF475569),
                                fontWeight: FontWeight.w500,
                                height: 1.3,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildTimeIndicator(String label, String value, Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.outfit(
            fontSize: 8,
            fontWeight: FontWeight.w800,
            color: Colors.blueGrey,
            letterSpacing: 0.5,
          ),
        ),
        const SizedBox(height: 2),
        Row(
          children: [
            Container(
              width: 4,
              height: 12,
              decoration: BoxDecoration(
                color: color,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(width: 6),
            Text(
              value,
              style: GoogleFonts.outfit(
                fontSize: 13,
                fontWeight: FontWeight.w900,
                color: const Color(0xFF1E293B),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildPhotoThumbnail(String label, String photoPath) {
    return InkWell(
      onTap: () => _showPhotoLightbox(label, photoPath),
      borderRadius: BorderRadius.circular(12),
      child: Container(
        width: 50,
        height: 50,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        clipBehavior: Clip.antiAlias,
        child: Stack(
          fit: StackFit.expand,
          children: [
            CachedNetworkImage(
              imageUrl: _buildImageUrl(photoPath),
              fit: BoxFit.cover,
              placeholder: (context, url) =>
                  const Center(child: CircularProgressIndicator()),
              errorWidget: (context, url, error) => const Icon(Icons.error),
              memCacheWidth: 800,
            ),
            Container(
              alignment: Alignment.bottomCenter,
              color: Colors.black26,
              padding: const EdgeInsets.symmetric(vertical: 2),
              child: Text(
                label,
                style: GoogleFonts.outfit(
                  color: Colors.white,
                  fontSize: 6,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showPhotoLightbox(String title, String path) {
    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.all(16),
        child: Stack(
          alignment: Alignment.center,
          children: [
            Container(
              decoration: BoxDecoration(
                color: Colors.black,
                borderRadius: BorderRadius.circular(24),
                border: Border.all(color: Colors.white12, width: 2),
              ),
              clipBehavior: Clip.antiAlias,
              child: CachedNetworkImage(
                imageUrl: _buildImageUrl(path),
                fit: BoxFit.contain,
                placeholder: (context, url) =>
                    const Center(child: CircularProgressIndicator()),
                errorWidget: (context, url, error) => const Icon(Icons.error),
                memCacheWidth: 800,
              ),
            ),
            Positioned(
              top: 16,
              right: 16,
              child: IconButton(
                icon: const Icon(
                  Icons.close_rounded,
                  color: Colors.white,
                  size: 28,
                ),
                onPressed: () => Navigator.pop(context),
              ),
            ),
            Positioned(
              bottom: 24,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
                decoration: BoxDecoration(
                  color: Colors.black54,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Text(
                  'BUKTI SWAFOTO $title',
                  style: GoogleFonts.outfit(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                    fontSize: 11,
                    letterSpacing: 1,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ==========================================
  // APPROVALS (PERSETUJUAN) VIEW
  // ==========================================
  Widget _buildApprovalsContent(String type) {
    return BlocListener<ApprovalBloc, ApprovalState>(
      listener: (context, state) {
        if (state is ApprovalActionSuccess) {
          SnackBarUtils.showSuccess(context, state.message);
        } else if (state is ApprovalFailure) {
          SnackBarUtils.showError(context, state.message);
        }
      },
      child: BlocBuilder<ApprovalBloc, ApprovalState>(
        builder: (context, state) {
          if (state is ApprovalLoading) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.symmetric(vertical: 40),
                child: CircularProgressIndicator(color: AppColors.blueHR),
              ),
            );
          }

          if (state is ApprovalDataLoaded) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  type == 'cuti'
                      ? 'PERSETUJUAN CUTI & IZIN'
                      : 'PERSETUJUAN REIMBURSEMENT',
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFF64748B),
                    letterSpacing: 1.5,
                  ),
                ),
                const SizedBox(height: 16),
                type == 'cuti'
                    ? _buildPendingLeavesList(state)
                    : _buildPendingReimbursementList(state),
              ],
            );
          }

          return const SizedBox();
        },
      ),
    );
  }

  Widget _buildPendingLeavesList(ApprovalDataLoaded state) {
    final leaves = state.pendingLeaves;
    if (leaves.isEmpty) {
      return _buildEmptyState('Tidak ada pengajuan cuti atau izin pending.');
    }

    return Column(
      children: [
        ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: leaves.length,
          itemBuilder: (context, index) =>
              _buildLeaveApprovalCard(leaves[index]),
        ),
        if (state.hasMoreLeaves)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 16.0),
            child: state.isLoadingMoreLeaves
                ? const CircularProgressIndicator(color: AppColors.blueHR)
                : TextButton(
                    onPressed: () => context.read<ApprovalBloc>().add(
                      LoadMoreLeavesRequested(),
                    ),
                    style: TextButton.styleFrom(
                      foregroundColor: AppColors.blueHR,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 24,
                        vertical: 12,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                        side: BorderSide(
                          color: AppColors.blueHR.withOpacity(0.5),
                          width: 1,
                        ),
                      ),
                    ),
                    child: Text(
                      'Muat Lebih Banyak',
                      style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
                    ),
                  ),
          ),
      ],
    );
  }

  Widget _buildPendingReimbursementList(ApprovalDataLoaded state) {
    final reimbursements = state.pendingReimbursements;
    if (reimbursements.isEmpty) {
      return _buildEmptyState('Tidak ada pengajuan reimbursement pending.');
    }

    return Column(
      children: [
        ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: reimbursements.length,
          itemBuilder: (context, index) =>
              _buildReimbursementApprovalCard(reimbursements[index]),
        ),
        if (state.hasMoreReimbursements)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 16.0),
            child: state.isLoadingMoreReimbursements
                ? const CircularProgressIndicator(color: AppColors.blueHR)
                : TextButton(
                    onPressed: () => context.read<ApprovalBloc>().add(
                      LoadMoreReimbursementsRequested(),
                    ),
                    style: TextButton.styleFrom(
                      foregroundColor: AppColors.blueHR,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 24,
                        vertical: 12,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                        side: BorderSide(
                          color: AppColors.blueHR.withOpacity(0.5),
                          width: 1,
                        ),
                      ),
                    ),
                    child: Text(
                      'Muat Lebih Banyak',
                      style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
                    ),
                  ),
          ),
      ],
    );
  }

  Widget _buildLeaveApprovalCard(Leave leave) {
    final createdAtStr = leave.createdAt != null
        ? DateFormat('dd MMM yyyy').format(leave.createdAt!)
        : '-';
    final rangeStr =
        '${DateFormat('dd MMM').format(leave.startDate)} - ${DateFormat('dd MMM yyyy').format(leave.endDate)}';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: AppColors.blueHR.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  (leave.leaveTypeName ?? 'PENGAJUAN').toUpperCase(),
                  style: GoogleFonts.outfit(
                    fontSize: 8,
                    fontWeight: FontWeight.w900,
                    color: AppColors.blueHR,
                  ),
                ),
              ),
              Text(
                createdAtStr,
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  color: Colors.blueGrey,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          Text(
            leave.employeeName ?? 'Karyawan',
            style: GoogleFonts.outfit(
              fontSize: 16,
              fontWeight: FontWeight.w900,
              color: const Color(0xFF0F172A),
            ),
          ),
          const SizedBox(height: 6),
          Row(
            children: [
              const Icon(
                Icons.calendar_month_rounded,
                size: 14,
                color: Colors.blueGrey,
              ),
              const SizedBox(width: 6),
              Text(
                rangeStr,
                style: GoogleFonts.outfit(
                  fontSize: 12,
                  color: const Color(0xFF475569),
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAF9),
              borderRadius: BorderRadius.circular(16),
            ),
            width: double.infinity,
            child: Text(
              leave.reason,
              style: GoogleFonts.outfit(
                fontSize: 12,
                color: const Color(0xFF475569),
                fontWeight: FontWeight.w500,
                height: 1.4,
              ),
            ),
          ),
          const SizedBox(height: 16),
          _buildActionButtons(
            onApprove: () => context.read<ApprovalBloc>().add(
              ApproveLeaveRequested(leave.id, 'approved'),
            ),
            onReject: () => context.read<ApprovalBloc>().add(
              ApproveLeaveRequested(leave.id, 'rejected'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildReimbursementApprovalCard(Reimbursement r) {
    final currencyFormat = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final createdAtStr = DateFormat('dd MMM yyyy').format(r.createdAt);

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: AppColors.blueHR.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  'REIMBURSEMENT',
                  style: GoogleFonts.outfit(
                    fontSize: 8,
                    fontWeight: FontWeight.w900,
                    color: AppColors.blueHR,
                  ),
                ),
              ),
              Text(
                createdAtStr,
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  color: Colors.blueGrey,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          Text(
            r.employeeName ?? 'Karyawan',
            style: GoogleFonts.outfit(
              fontSize: 16,
              fontWeight: FontWeight.w900,
              color: const Color(0xFF0F172A),
            ),
          ),
          const SizedBox(height: 4),
          Text(
            r.title,
            style: GoogleFonts.outfit(
              fontSize: 13,
              fontWeight: FontWeight.w800,
              color: const Color(0xFF1E293B),
            ),
          ),
          if (r.description != null && r.description!.isNotEmpty) ...[
            const SizedBox(height: 6),
            Text(
              r.description!,
              style: GoogleFonts.outfit(
                fontSize: 11,
                color: Colors.blueGrey,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
          const SizedBox(height: 14),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: AppColors.blueHR.withOpacity(0.1),
              borderRadius: BorderRadius.circular(16),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'TOTAL NOMINAL:',
                  style: GoogleFonts.outfit(
                    fontSize: 10,
                    fontWeight: FontWeight.w800,
                    color: AppColors.blueHR,
                  ),
                ),
                Text(
                  currencyFormat.format(r.amount),
                  style: GoogleFonts.outfit(
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                    color: AppColors.blueHR,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          _buildActionButtons(
            onApprove: () => context.read<ApprovalBloc>().add(
              ApproveReimbursementRequested(r.id, 'approved'),
            ),
            onReject: () => context.read<ApprovalBloc>().add(
              ApproveReimbursementRequested(r.id, 'rejected'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildActionButtons({
    required VoidCallback onApprove,
    required VoidCallback onReject,
  }) {
    return Row(
      children: [
        Expanded(
          child: InkWell(
            onTap: onReject,
            borderRadius: BorderRadius.circular(14),
            child: Container(
              padding: const EdgeInsets.symmetric(vertical: 12),
              decoration: BoxDecoration(
                border: Border.all(color: Colors.redAccent),
                borderRadius: BorderRadius.circular(14),
              ),
              child: Center(
                child: Text(
                  'TOLAK',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    color: Colors.redAccent,
                    fontSize: 12,
                  ),
                ),
              ),
            ),
          ),
        ),
        const SizedBox(width: 14),
        Expanded(
          child: InkWell(
            onTap: onApprove,
            borderRadius: BorderRadius.circular(14),
            child: Container(
              padding: const EdgeInsets.symmetric(vertical: 12),
              decoration: BoxDecoration(
                color: Colors.green,
                borderRadius: BorderRadius.circular(14),
                boxShadow: [
                  BoxShadow(
                    color: Colors.green.withOpacity(0.2),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: Center(
                child: Text(
                  'SETUJUI',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    color: Colors.white,
                    fontSize: 12,
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildEmptyState(String message) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 80),
      child: Center(
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: const BoxDecoration(
                color: Color(0xFFF1F5F9),
                shape: BoxShape.circle,
              ),
              child: Icon(
                Icons.assignment_turned_in_rounded,
                size: 48,
                color: AppColors.blueHRLight.withOpacity(0.5),
              ),
            ),
            const SizedBox(height: 16),
            Text(
              message,
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(
                color: Colors.blueGrey,
                fontWeight: FontWeight.w700,
                fontSize: 13,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
