import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';

class AnnouncementManagementPage extends StatefulWidget {
  const AnnouncementManagementPage({super.key});

  @override
  State<AnnouncementManagementPage> createState() =>
      _AnnouncementManagementPageState();
}

class _AnnouncementManagementPageState extends State<AnnouncementManagementPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  bool _isLoading = true;
  List<dynamic> _announcements = [];
  List<dynamic> _allEmployees = [];
  List<dynamic> _jobPositions = [];
  List<dynamic> _accessRoles = [];

  final ApiClient _apiClient = di.sl<ApiClient>();

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _fetchData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _fetchData() async {
    setState(() => _isLoading = true);
    await Future.wait([
      _fetchAnnouncements(),
      _fetchMasterData(),
    ]);
    if (mounted) {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _fetchAnnouncements() async {
    try {
      final response = await _apiClient.client.get('admin/announcements');
      if (response.statusCode == 200 && mounted) {
        setState(() {
          _announcements = response.data['data'] ?? [];
        });
      }
    } catch (e) {
      if (mounted) {
        SnackBarUtils.showError(context, 'Gagal mengambil data pengumuman: $e');
      }
    }
  }

  Future<void> _fetchMasterData() async {
    try {
      final empRes = await _apiClient.client.get('employees?limit=all&status=active');
      if (empRes.statusCode == 200 && mounted) {
        _allEmployees = empRes.data['data'] ?? [];
      }
    } catch (_) {}

    try {
      final posRes = await _apiClient.client.get('job-positions');
      if (posRes.statusCode == 200 && mounted) {
        _jobPositions = posRes.data['data'] ?? [];
      }
    } catch (_) {}

    try {
      final roleRes = await _apiClient.client.get('access-roles');
      if (roleRes.statusCode == 200 && mounted) {
        _accessRoles = roleRes.data['data'] ?? [];
      }
    } catch (_) {}
  }

  List<dynamic> get _directAnnouncements {
    return _announcements.where((a) {
      final targetType = a['target_type'] ?? 'broadcast';
      final category = a['category'] ?? '';
      return targetType == 'individual' || category == 'urgent_call';
    }).toList();
  }

  List<dynamic> get _generalAnnouncements {
    return _announcements.where((a) {
      final targetType = a['target_type'] ?? 'broadcast';
      final category = a['category'] ?? '';
      return targetType != 'individual' && category != 'urgent_call';
    }).toList();
  }

  Future<void> _deleteAnnouncement(String id) async {
    final bool confirm = await showDialog(
          context: context,
          builder: (ctx) => AlertDialog(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
            title: Text('Hapus Pengumuman?',
                style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w800)),
            content: Text('Apakah Anda yakin ingin menghapus pengumuman ini secara permanen?',
                style: GoogleFonts.plusJakartaSans(fontSize: 14)),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx, false),
                child: Text('BATAL',
                    style: GoogleFonts.plusJakartaSans(
                        color: Colors.grey, fontWeight: FontWeight.w700)),
              ),
              ElevatedButton(
                onPressed: () => Navigator.pop(ctx, true),
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.error,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                child: Text('HAPUS',
                    style: GoogleFonts.plusJakartaSans(
                        color: Colors.white, fontWeight: FontWeight.w800)),
              ),
            ],
          ),
        ) ??
        false;

    if (!confirm) return;

    try {
      final response = await _apiClient.client.delete('announcements/$id');
      if (response.statusCode == 200) {
        SnackBarUtils.showSuccess(context, 'Pengumuman berhasil dihapus');
        _fetchAnnouncements();
      } else {
        SnackBarUtils.showError(context, 'Gagal menghapus pengumuman');
      }
    } catch (e) {
      SnackBarUtils.showError(context, 'Terjadi kesalahan jaringan: $e');
    }
  }

  void _showAnnouncementForm({Map<String, dynamic>? initialData, String? defaultMode}) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => AnnouncementFormSheet(
        initialData: initialData,
        defaultMode: defaultMode,
        allEmployees: _allEmployees,
        jobPositions: _jobPositions,
        accessRoles: _accessRoles,
        onSubmit: (data) async {
          Navigator.pop(ctx);
          try {
            if (initialData == null) {
              final response = await _apiClient.client.post('announcements', data: data);
              if (response.statusCode == 201 || response.statusCode == 200) {
                SnackBarUtils.showSuccess(
                    context, 'Pengumuman & Direct Notifikasi FCM berhasil dikirimkan!');
                _fetchAnnouncements();
              } else {
                SnackBarUtils.showError(
                    context, response.data?['message'] ?? 'Gagal membuat pengumuman');
              }
            } else {
              final response =
                  await _apiClient.client.put('announcements/${initialData['id']}', data: data);
              if (response.statusCode == 200) {
                SnackBarUtils.showSuccess(context, 'Pengumuman berhasil diperbarui');
                _fetchAnnouncements();
              } else {
                SnackBarUtils.showError(
                    context, response.data?['message'] ?? 'Gagal mengupdate pengumuman');
              }
            }
          } catch (e) {
            SnackBarUtils.showError(context, 'Terjadi kesalahan jaringan: $e');
          }
        },
      ),
    );
  }

  void _showTrackingModal(String announcementId, String title) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => AnnouncementTrackingSheet(
        announcementId: announcementId,
        announcementTitle: title,
        apiClient: _apiClient,
      ),
    );
  }

  void _showResendFcmDialog(Map<String, dynamic> announcement) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => ResendFcmSheet(
        announcement: announcement,
        allEmployees: _allEmployees,
        apiClient: _apiClient,
        onSuccess: () {
          _fetchAnnouncements();
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'MANAJEMEN PENGUMUMAN',
          style: GoogleFonts.plusJakartaSans(
              fontWeight: FontWeight.w900,
              color: Colors.white,
              fontSize: 15,
              letterSpacing: 0.8),
        ),
        backgroundColor: AppColors.blueHRDark,
        centerTitle: true,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded, color: Colors.white, size: 20),
          onPressed: () => Navigator.pop(context),
        ),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(54),
          child: Container(
            color: AppColors.blueHRDark,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
            child: Container(
              height: 42,
              decoration: BoxDecoration(
                color: Colors.black.withOpacity(0.15),
                borderRadius: BorderRadius.circular(12),
              ),
              child: TabBar(
                controller: _tabController,
                indicatorSize: TabBarIndicatorSize.tab,
                indicator: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(10),
                ),
                labelColor: AppColors.blueHRDark,
                unselectedLabelColor: Colors.white70,
                labelStyle: GoogleFonts.plusJakartaSans(
                    fontSize: 12, fontWeight: FontWeight.w800),
                unselectedLabelStyle: GoogleFonts.plusJakartaSans(
                    fontSize: 12, fontWeight: FontWeight.w600),
                tabs: [
                  Tab(text: 'Panggilan & Direct (${_directAnnouncements.length})'),
                  Tab(text: 'Pengumuman Umum (${_generalAnnouncements.length})'),
                ],
              ),
            ),
          ),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: AppColors.blueHRDark))
          : TabBarView(
              controller: _tabController,
              children: [
                _buildAnnouncementList(_directAnnouncements, isDirectTab: true),
                _buildAnnouncementList(_generalAnnouncements, isDirectTab: false),
              ],
            ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          final isDirect = _tabController.index == 0;
          _showAnnouncementForm(defaultMode: isDirect ? 'direct' : 'general');
        },
        backgroundColor: AppColors.blueHRDark,
        elevation: 6,
        icon: const Icon(Icons.add_rounded, color: Colors.white),
        label: Text(
          'BUAT PENGUMUMAN',
          style: GoogleFonts.plusJakartaSans(
              color: Colors.white, fontWeight: FontWeight.w800, fontSize: 13, letterSpacing: 0.5),
        ),
      ),
    );
  }

  Widget _buildAnnouncementList(List<dynamic> items, {required bool isDirectTab}) {
    if (items.isEmpty) {
      return _buildEmptyState(isDirectTab);
    }

    return RefreshIndicator(
      onRefresh: _fetchData,
      color: AppColors.blueHRDark,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 90),
        itemCount: items.length,
        itemBuilder: (context, index) {
          final item = items[index];
          return _buildAnnouncementCard(item);
        },
      ),
    );
  }

  Widget _buildEmptyState(bool isDirectTab) {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: AppColors.blueHRDark.withOpacity(0.08),
                shape: BoxShape.circle,
              ),
              child: Icon(
                isDirectTab ? Icons.person_search_rounded : Icons.campaign_outlined,
                size: 64,
                color: AppColors.blueHRDark,
              ),
            ),
            const SizedBox(height: 20),
            Text(
              isDirectTab ? 'Belum Ada Panggilan Direct' : 'Belum Ada Pengumuman Umum',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 17, fontWeight: FontWeight.w800, color: const Color(0xFF1E293B)),
            ),
            const SizedBox(height: 8),
            Text(
              isDirectTab
                  ? 'Kirim panggilan menghadap direct ke karyawan tertentu dengan FCM Push & Popup'
                  : 'Siarkan pengumuman broadcast atau berdasarkan jabatan ke seluruh staff',
              textAlign: TextAlign.center,
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 13, color: const Color(0xFF64748B), height: 1.5),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () =>
                  _showAnnouncementForm(defaultMode: isDirectTab ? 'direct' : 'general'),
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.blueHRDark,
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
              ),
              icon: const Icon(Icons.add_rounded, color: Colors.white, size: 18),
              label: Text(
                isDirectTab ? 'Buat Panggilan Direct' : 'Buat Pengumuman Baru',
                style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w800, color: Colors.white, fontSize: 13),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAnnouncementCard(Map<String, dynamic> item) {
    final id = item['id']?.toString() ?? '';
    final title = item['title'] ?? '-';
    final content = item['content'] ?? '-';
    final category = item['category'] ?? 'info';
    final targetType = item['target_type'] ?? 'broadcast';
    final isPopup = item['is_popup'] ?? false;
    final isActive = item['is_active'] ?? false;
    final targets = item['targets'] as List<dynamic>? ?? [];

    DateTime createdAt = DateTime.now();
    if (item['created_at'] != null) {
      try {
        createdAt = DateTime.parse(item['created_at']).toLocal();
      } catch (_) {}
    }

    Color categoryColor = const Color(0xFF3B82F6);
    IconData categoryIcon = Icons.info_outline_rounded;
    String categoryLabel = 'Informasi Umum';

    if (category == 'urgent_call' || category == 'alert') {
      categoryColor = const Color(0xFFEF4444);
      categoryIcon = Icons.warning_amber_rounded;
      categoryLabel = 'Panggilan Menghadap';
    } else if (category == 'event') {
      categoryColor = const Color(0xFF2563EB);
      categoryIcon = Icons.calendar_month_rounded;
      categoryLabel = 'Undangan & Acara';
    } else if (category == 'policy') {
      categoryColor = const Color(0xFF10B981);
      categoryIcon = Icons.verified_user_rounded;
      categoryLabel = 'Kebijakan Perusahaan';
    }

    String targetLabel = 'Semua Karyawan (Broadcast)';
    if (targetType == 'role') {
      final roleCount = targets.length;
      targetLabel = 'Jabatan / Role ($roleCount Target)';
    } else if (targetType == 'individual') {
      final userCount = targets.length;
      targetLabel = 'Panggilan Direct ($userCount Karyawan)';
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 12,
              offset: const Offset(0, 4)),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header Bar
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: categoryColor.withOpacity(0.08),
              borderRadius: const BorderRadius.only(
                  topLeft: Radius.circular(20), topRight: Radius.circular(20)),
              border: Border(bottom: BorderSide(color: categoryColor.withOpacity(0.15))),
            ),
            child: Row(
              children: [
                Icon(categoryIcon, size: 16, color: categoryColor),
                const SizedBox(width: 8),
                Text(
                  categoryLabel.toUpperCase(),
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: categoryColor,
                      letterSpacing: 0.8),
                ),
                const Spacer(),
                if (isPopup) ...[
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: Colors.purple.shade50,
                      border: Border.all(color: Colors.purple.shade200),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.fullscreen_rounded, size: 12, color: Colors.purple.shade700),
                        const SizedBox(width: 4),
                        Text(
                          'POPUP',
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 9,
                              fontWeight: FontWeight.w900,
                              color: Colors.purple.shade700),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 6),
                ],
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: isActive ? Colors.green.shade50 : Colors.grey.shade100,
                    border: Border.all(
                        color: isActive ? Colors.green.shade300 : Colors.grey.shade300),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    isActive ? 'AKTIF' : 'NONAKTIF',
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 9,
                        fontWeight: FontWeight.w800,
                        color: isActive ? Colors.green.shade800 : Colors.grey.shade600),
                  ),
                ),
              ],
            ),
          ),

          // Body Content
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 16,
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF1E293B)),
                ),
                const SizedBox(height: 6),
                Text(
                  content,
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 13, color: const Color(0xFF475569), height: 1.45),
                  maxLines: 3,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 14),

                // Target Chip & Date
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F5F9),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.people_alt_rounded,
                              size: 13, color: Color(0xFF64748B)),
                          const SizedBox(width: 6),
                          Text(
                            targetLabel,
                            style: GoogleFonts.plusJakartaSans(
                                fontSize: 11,
                                fontWeight: FontWeight.w700,
                                color: const Color(0xFF475569)),
                          ),
                        ],
                      ),
                    ),
                    const Spacer(),
                    Text(
                      DateFormat('dd MMM yyyy, HH:mm').format(createdAt),
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 11,
                          color: const Color(0xFF94A3B8),
                          fontWeight: FontWeight.w600),
                    ),
                  ],
                ),
                const SizedBox(height: 14),
                const Divider(height: 1, color: Color(0xFFF1F5F9)),
                const SizedBox(height: 12),

                // Action Bar
                Row(
                  children: [
                    // Tracking Button
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: () => _showTrackingModal(id, title),
                        style: OutlinedButton.styleFrom(
                          side: const BorderSide(color: Color(0xFFCBD5E1)),
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12)),
                          padding: const EdgeInsets.symmetric(vertical: 8),
                        ),
                        icon: const Icon(Icons.analytics_outlined,
                            size: 16, color: Color(0xFF475569)),
                        label: Text(
                          'Tracking',
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                              color: const Color(0xFF475569)),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),

                    // Resend FCM Button
                    Expanded(
                      child: ElevatedButton.icon(
                        onPressed: () => _showResendFcmDialog(item),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.blueHRDark,
                          elevation: 0,
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12)),
                          padding: const EdgeInsets.symmetric(vertical: 8),
                        ),
                        icon: const Icon(Icons.send_rounded, size: 14, color: Colors.white),
                        label: Text(
                          'Kirim Ulang',
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                              color: Colors.white),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),

                    // Edit
                    IconButton(
                      icon: const Icon(Icons.edit_outlined,
                          color: Color(0xFF3B82F6), size: 20),
                      style: IconButton.styleFrom(
                        backgroundColor: const Color(0xFFEFF6FF),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(10)),
                      ),
                      onPressed: () => _showAnnouncementForm(initialData: item),
                    ),
                    const SizedBox(width: 4),

                    // Delete
                    IconButton(
                      icon: const Icon(Icons.delete_outline_rounded,
                          color: AppColors.error, size: 20),
                      style: IconButton.styleFrom(
                        backgroundColor: const Color(0xFFFEF2F2),
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(10)),
                      ),
                      onPressed: () => _deleteAnnouncement(id),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ─────────────────────────────────────────────────────────────────
// ANNOUNCEMENT FORM SHEET
// ─────────────────────────────────────────────────────────────────
class AnnouncementFormSheet extends StatefulWidget {
  final Map<String, dynamic>? initialData;
  final String? defaultMode;
  final List<dynamic> allEmployees;
  final List<dynamic> jobPositions;
  final List<dynamic> accessRoles;
  final Function(Map<String, dynamic>) onSubmit;

  const AnnouncementFormSheet({
    super.key,
    this.initialData,
    this.defaultMode,
    required this.allEmployees,
    required this.jobPositions,
    required this.accessRoles,
    required this.onSubmit,
  });

  @override
  State<AnnouncementFormSheet> createState() => _AnnouncementFormSheetState();
}

class _AnnouncementFormSheetState extends State<AnnouncementFormSheet> {
  final _formKey = GlobalKey<FormState>();
  late TextEditingController _titleController;
  late TextEditingController _contentController;
  late TextEditingController _employeeSearchController;
  late TextEditingController _roleSearchController;

  String _category = 'urgent_call';
  String _targetType = 'individual'; // 'broadcast' | 'role' | 'individual'
  List<String> _targetRoles = [];
  List<String> _targetUserIds = [];
  bool _isPopup = true;
  DateTime? _popupExpiresAt;
  bool _isActive = true;
  bool _isSubmitting = false;

  int _roleSubTabIndex = 0; // 0: Job Positions, 1: Access Roles

  @override
  void initState() {
    super.initState();
    _titleController =
        TextEditingController(text: widget.initialData?['title'] ?? '');
    _contentController =
        TextEditingController(text: widget.initialData?['content'] ?? '');
    _employeeSearchController = TextEditingController();
    _roleSearchController = TextEditingController();

    if (widget.initialData != null) {
      final init = widget.initialData!;
      _category = init['category'] ?? 'policy';
      _targetType = init['target_type'] ?? 'broadcast';
      _isPopup = init['is_popup'] ?? false;
      _isActive = init['is_active'] ?? true;

      if (init['popup_expires_at'] != null) {
        try {
          _popupExpiresAt = DateTime.parse(init['popup_expires_at']).toLocal();
        } catch (_) {}
      }

      final targets = init['targets'] as List<dynamic>? ?? [];
      _targetRoles = [];
      _targetUserIds = [];
      for (final t in targets) {
        if (t['target_role'] != null && t['target_role'].toString().isNotEmpty) {
          _targetRoles.add(t['target_role'].toString());
        }
        if (t['target_user_id'] != null &&
            t['target_user_id'].toString().isNotEmpty) {
          _targetUserIds.add(t['target_user_id'].toString());
        }
      }
    } else {
      if (widget.defaultMode == 'general') {
        _category = 'policy';
        _targetType = 'broadcast';
        _isPopup = false;
      } else {
        _category = 'urgent_call';
        _targetType = 'individual';
        _isPopup = true;
      }
    }
  }

  @override
  void dispose() {
    _titleController.dispose();
    _contentController.dispose();
    _employeeSearchController.dispose();
    _roleSearchController.dispose();
    super.dispose();
  }

  List<dynamic> get _filteredEmployees {
    final q = _employeeSearchController.text.trim().toLowerCase();
    if (q.isEmpty) return widget.allEmployees;
    return widget.allEmployees.where((emp) {
      final name =
          '${emp['first_name'] ?? ''} ${emp['last_name'] ?? ''}'.toLowerCase();
      final nik = (emp['employee_id_number'] ?? '').toString().toLowerCase();
      final dept = (emp['department']?['name'] ?? '').toString().toLowerCase();
      final job = (emp['job_position']?['title'] ?? '').toString().toLowerCase();
      return name.contains(q) || nik.contains(q) || dept.contains(q) || job.contains(q);
    }).toList();
  }

  List<dynamic> get _filteredJobPositions {
    final q = _roleSearchController.text.trim().toLowerCase();
    if (q.isEmpty) return widget.jobPositions;
    return widget.jobPositions.where((pos) {
      final title = (pos['title'] ?? '').toString().toLowerCase();
      return title.contains(q);
    }).toList();
  }

  List<dynamic> get _filteredAccessRoles {
    final q = _roleSearchController.text.trim().toLowerCase();
    if (q.isEmpty) return widget.accessRoles;
    return widget.accessRoles.where((r) {
      final name = (r['name'] ?? '').toString().toLowerCase();
      return name.contains(q);
    }).toList();
  }

  int get _selectedRolesAffectedCount {
    if (_targetRoles.isEmpty) return 0;
    final selected = _targetRoles.map((r) => r.toLowerCase()).toSet();
    final matched = <String>{};

    for (final emp in widget.allEmployees) {
      final jobTitle = (emp['job_position']?['title'] ?? '').toString().toLowerCase();
      final role = (emp['user']?['role'] ?? emp['role'] ?? '').toString().toLowerCase();
      final uid = (emp['user_id'] ?? emp['user']?['id'] ?? emp['id'])?.toString();

      if (uid != null && (selected.contains(jobTitle) || selected.contains(role))) {
        matched.add(uid);
      }
    }
    return matched.length;
  }

  void _handleSubmit() {
    if (!_formKey.currentState!.validate()) return;

    if (_targetType == 'role' && _targetRoles.isEmpty) {
      SnackBarUtils.showError(context, 'Pilih minimal satu jabatan/role target penerima');
      return;
    }

    if (_targetType == 'individual' && _targetUserIds.isEmpty) {
      SnackBarUtils.showError(context, 'Pilih minimal satu karyawan target penerima');
      return;
    }

    setState(() => _isSubmitting = true);

    final payload = <String, dynamic>{
      'title': _titleController.text.trim(),
      'content': _contentController.text.trim(),
      'category': _category,
      'target_type': _targetType,
      'is_popup': _isPopup,
      'is_active': _isActive,
      if (_targetType == 'role') 'target_roles': _targetRoles,
      if (_targetType == 'individual') 'target_user_ids': _targetUserIds,
      if (_isPopup && _popupExpiresAt != null)
        'popup_expires_at': _popupExpiresAt!.toUtc().toIso8601String(),
    };

    widget.onSubmit(payload);
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      height: MediaQuery.of(context).size.height * 0.92,
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      child: Column(
        children: [
          // Header Sheet
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: Color(0xFFF1F5F9))),
            ),
            child: Row(
              children: [
                Text(
                  widget.initialData == null ? 'Buat Pengumuman Baru' : 'Edit Pengumuman',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 16, fontWeight: FontWeight.w900, color: const Color(0xFF0F172A)),
                ),
                const Spacer(),
                IconButton(
                  icon: const Icon(Icons.close_rounded, size: 20, color: Color(0xFF64748B)),
                  onPressed: () => Navigator.pop(context),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
          ),

          // Scrollable Form Body
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(20),
              physics: const BouncingScrollPhysics(),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Direct Push Banner Info
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: const Color(0xFFEFF6FF),
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(color: const Color(0xFFBFDBFE)),
                      ),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Icon(Icons.notifications_active_rounded,
                              size: 18, color: Color(0xFF2563EB)),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Text(
                              'Notifikasi FCM Direct akan otomatis terkirim langsung ke HP seluruh target penerima begitu disimpan.',
                              style: GoogleFonts.plusJakartaSans(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w600,
                                  color: const Color(0xFF1E40AF),
                                  height: 1.4),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 20),

                    // Kategori Pengumuman
                    _buildSectionTitle('KATEGORI PENGUMUMAN'),
                    const SizedBox(height: 10),
                    _buildCategoryGrid(),
                    const SizedBox(height: 20),

                    // Judul
                    _buildSectionTitle('JUDUL PENGUMUMAN *'),
                    const SizedBox(height: 8),
                    TextFormField(
                      controller: _titleController,
                      style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.w600),
                      decoration: _buildInputDecoration(
                        hint: 'Contoh: Panggilan Menghadap Direksi / Kebijakan Shift Baru',
                        icon: Icons.title_rounded,
                      ),
                      validator: (v) =>
                          (v == null || v.trim().isEmpty) ? 'Judul pengumuman wajib diisi' : null,
                    ),
                    const SizedBox(height: 20),

                    // Isi Konten
                    _buildSectionTitle('ISI PENGUMUMAN *'),
                    const SizedBox(height: 8),
                    TextFormField(
                      controller: _contentController,
                      maxLines: 4,
                      style: GoogleFonts.plusJakartaSans(fontSize: 14),
                      decoration: _buildInputDecoration(
                        hint: 'Tuliskan rincian pengumuman secara lengkap...',
                        icon: Icons.notes_rounded,
                      ),
                      validator: (v) =>
                          (v == null || v.trim().isEmpty) ? 'Isi pengumuman wajib diisi' : null,
                    ),
                    const SizedBox(height: 24),

                    // Target Penerima Section
                    _buildSectionTitle('TARGET PENERIMA (SIAPA YANG MENERIMA)'),
                    const SizedBox(height: 10),
                    _buildTargetTypeSelector(),
                    const SizedBox(height: 16),

                    // Target Specific Content
                    if (_targetType == 'role') _buildRoleTargetPicker(),
                    if (_targetType == 'individual') _buildIndividualTargetPicker(),
                    const SizedBox(height: 24),

                    // Popup & Expiry Setting
                    _buildSectionTitle('PENGATURAN POPUP & TAMPILAN'),
                    const SizedBox(height: 10),
                    Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Column(
                        children: [
                          SwitchListTile(
                            contentPadding: EdgeInsets.zero,
                            title: Text(
                              'Tampilkan Sebagai Popup Interaktif',
                              style: GoogleFonts.plusJakartaSans(
                                  fontSize: 14,
                                  fontWeight: FontWeight.w800,
                                  color: const Color(0xFF0F172A)),
                            ),
                            subtitle: Text(
                              'Penerima wajib membaca scroll sampai bawah & menekan tombol konfirmasi.',
                              style: GoogleFonts.plusJakartaSans(
                                  fontSize: 12, color: const Color(0xFF64748B)),
                            ),
                            value: _isPopup,
                            activeColor: AppColors.blueHRDark,
                            onChanged: (val) => setState(() => _isPopup = val),
                          ),
                          if (_isPopup) ...[
                            const Divider(height: 24, color: Color(0xFFE2E8F0)),
                            ListTile(
                              contentPadding: EdgeInsets.zero,
                              title: Text(
                                'Batas Waktu Popup Kadaluarsa (Opsional)',
                                style: GoogleFonts.plusJakartaSans(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w700,
                                    color: const Color(0xFF334155)),
                              ),
                              subtitle: Text(
                                _popupExpiresAt == null
                                    ? 'Tidak ada batas waktu (selalu muncul sampai dikonfirmasi)'
                                    : 'Berakhir pada: ${DateFormat('dd MMM yyyy, HH:mm').format(_popupExpiresAt!)}',
                                style: GoogleFonts.plusJakartaSans(
                                    fontSize: 12,
                                    color: _popupExpiresAt == null
                                        ? const Color(0xFF94A3B8)
                                        : const Color(0xFF2563EB)),
                              ),
                              trailing: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  if (_popupExpiresAt != null)
                                    IconButton(
                                      icon: const Icon(Icons.clear_rounded,
                                          size: 18, color: Colors.grey),
                                      onPressed: () =>
                                          setState(() => _popupExpiresAt = null),
                                    ),
                                  OutlinedButton(
                                    onPressed: _pickPopupExpiryDateTime,
                                    style: OutlinedButton.styleFrom(
                                      padding: const EdgeInsets.symmetric(
                                          horizontal: 12, vertical: 6),
                                      shape: RoundedRectangleBorder(
                                          borderRadius: BorderRadius.circular(10)),
                                    ),
                                    child: Text('PILIH TGL',
                                        style: GoogleFonts.plusJakartaSans(
                                            fontSize: 11, fontWeight: FontWeight.w800)),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                    const SizedBox(height: 16),

                    // Active Toggle if edit mode
                    if (widget.initialData != null)
                      SwitchListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text(
                          'Status Pengumuman Aktif',
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 14, fontWeight: FontWeight.w800),
                        ),
                        subtitle: Text('Pengumuman hanya tampil jika berstatus aktif',
                            style: GoogleFonts.plusJakartaSans(fontSize: 12, color: Colors.grey)),
                        value: _isActive,
                        activeColor: Colors.green,
                        onChanged: (val) => setState(() => _isActive = val),
                      ),
                    const SizedBox(height: 32),

                    // Submit Button
                    SizedBox(
                      width: double.infinity,
                      height: 52,
                      child: ElevatedButton(
                        onPressed: _isSubmitting ? null : _handleSubmit,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.blueHRDark,
                          elevation: 4,
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16)),
                        ),
                        child: _isSubmitting
                            ? const SizedBox(
                                width: 22,
                                height: 22,
                                child: CircularProgressIndicator(
                                    color: Colors.white, strokeWidth: 2.5),
                              )
                            : Text(
                                widget.initialData == null
                                    ? 'KIRIM PENGUMUMAN SEKARANG'
                                    : 'SIMPAN PERUBAHAN',
                                style: GoogleFonts.plusJakartaSans(
                                    fontWeight: FontWeight.w900,
                                    color: Colors.white,
                                    fontSize: 13,
                                    letterSpacing: 0.8),
                              ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: GoogleFonts.plusJakartaSans(
          fontSize: 11,
          fontWeight: FontWeight.w900,
          color: const Color(0xFF64748B),
          letterSpacing: 1),
    );
  }

  Widget _buildCategoryGrid() {
    final list = [
      {
        'value': 'urgent_call',
        'label': 'Panggilan Menghadap',
        'sub': 'Urgent & Prioritas',
        'icon': Icons.warning_amber_rounded,
        'color': const Color(0xFFEF4444),
      },
      {
        'value': 'event',
        'label': 'Undangan & Meeting',
        'sub': 'Acara / Rapat',
        'icon': Icons.calendar_month_rounded,
        'color': const Color(0xFF2563EB),
      },
      {
        'value': 'policy',
        'label': 'Kebijakan Perusahaan',
        'sub': 'Aturan & SOP',
        'icon': Icons.verified_user_rounded,
        'color': const Color(0xFF10B981),
      },
      {
        'value': 'info',
        'label': 'Informasi Umum',
        'sub': 'Berita & Update',
        'icon': Icons.campaign_rounded,
        'color': const Color(0xFFF59E0B),
      },
    ];

    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        mainAxisSpacing: 10,
        crossAxisSpacing: 10,
        childAspectRatio: 2.2,
      ),
      itemCount: list.length,
      itemBuilder: (context, index) {
        final item = list[index];
        final val = item['value'] as String;
        final isSelected = _category == val;
        final color = item['color'] as Color;

        return InkWell(
          onTap: () => setState(() => _category = val),
          borderRadius: BorderRadius.circular(14),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: isSelected ? color.withOpacity(0.1) : const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                  color: isSelected ? color : const Color(0xFFE2E8F0),
                  width: isSelected ? 2 : 1),
            ),
            child: Row(
              children: [
                Icon(item['icon'] as IconData, size: 22, color: isSelected ? color : Colors.grey),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        item['label'] as String,
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                            color: isSelected ? color : const Color(0xFF1E293B)),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        item['sub'] as String,
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 10, color: const Color(0xFF64748B)),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildTargetTypeSelector() {
    final types = [
      {
        'value': 'broadcast',
        'label': 'Semua Karyawan',
        'sub': 'Broadcast Global',
        'icon': Icons.public_rounded,
      },
      {
        'value': 'role',
        'label': 'Berdasarkan Jabatan',
        'sub': 'Multi-role / Jabatan',
        'icon': Icons.badge_rounded,
      },
      {
        'value': 'individual',
        'label': 'Karyawan Tertentu',
        'sub': 'Direct Panggilan',
        'icon': Icons.person_pin_rounded,
      },
    ];

    return Column(
      children: types.map((t) {
        final val = t['value'] as String;
        final isSelected = _targetType == val;
        return Container(
          margin: const EdgeInsets.only(bottom: 8),
          child: InkWell(
            onTap: () => setState(() => _targetType = val),
            borderRadius: BorderRadius.circular(14),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
              decoration: BoxDecoration(
                color: isSelected
                    ? AppColors.blueHRDark.withOpacity(0.08)
                    : const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(14),
                border: Border.all(
                    color: isSelected ? AppColors.blueHRDark : const Color(0xFFE2E8F0),
                    width: isSelected ? 1.8 : 1),
              ),
              child: Row(
                children: [
                  Icon(t['icon'] as IconData,
                      size: 22,
                      color: isSelected ? AppColors.blueHRDark : const Color(0xFF94A3B8)),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          t['label'] as String,
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 13,
                              fontWeight: FontWeight.w800,
                              color: isSelected
                                  ? AppColors.blueHRDark
                                  : const Color(0xFF1E293B)),
                        ),
                        Text(
                          t['sub'] as String,
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 11, color: const Color(0xFF64748B)),
                        ),
                      ],
                    ),
                  ),
                  Radio<String>(
                    value: val,
                    groupValue: _targetType,
                    activeColor: AppColors.blueHRDark,
                    onChanged: (v) {
                      if (v != null) setState(() => _targetType = v);
                    },
                  ),
                ],
              ),
            ),
          ),
        );
      }).toList(),
    );
  }

  Widget _buildRoleTargetPicker() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                'Pilih Jabatan / Role Target',
                style: GoogleFonts.plusJakartaSans(
                    fontSize: 13, fontWeight: FontWeight.w800, color: const Color(0xFF0F172A)),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: AppColors.blueHRDark.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '± $_selectedRolesAffectedCount Karyawan',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11, fontWeight: FontWeight.w800, color: AppColors.blueHRDark),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Sub tabs: Jabatan vs Hak Akses
          Row(
            children: [
              Expanded(
                child: InkWell(
                  onTap: () => setState(() => _roleSubTabIndex = 0),
                  borderRadius: BorderRadius.circular(10),
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    decoration: BoxDecoration(
                      color: _roleSubTabIndex == 0 ? Colors.white : Colors.transparent,
                      borderRadius: BorderRadius.circular(10),
                      boxShadow: _roleSubTabIndex == 0
                          ? [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 4)]
                          : null,
                    ),
                    child: Center(
                      child: Text(
                        'Jabatan (${_filteredJobPositions.length})',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                            color: _roleSubTabIndex == 0
                                ? AppColors.blueHRDark
                                : const Color(0xFF64748B)),
                      ),
                    ),
                  ),
                ),
              ),
              Expanded(
                child: InkWell(
                  onTap: () => setState(() => _roleSubTabIndex = 1),
                  borderRadius: BorderRadius.circular(10),
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    decoration: BoxDecoration(
                      color: _roleSubTabIndex == 1 ? Colors.white : Colors.transparent,
                      borderRadius: BorderRadius.circular(10),
                      boxShadow: _roleSubTabIndex == 1
                          ? [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 4)]
                          : null,
                    ),
                    child: Center(
                      child: Text(
                        'Hak Akses (${_filteredAccessRoles.length})',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                            color: _roleSubTabIndex == 1
                                ? AppColors.blueHRDark
                                : const Color(0xFF64748B)),
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Role Search
          TextFormField(
            controller: _roleSearchController,
            onChanged: (_) => setState(() {}),
            style: GoogleFonts.plusJakartaSans(fontSize: 12),
            decoration: InputDecoration(
              hintText: 'Cari nama jabatan / role...',
              prefixIcon: const Icon(Icons.search_rounded, size: 18),
              isDense: true,
              filled: true,
              fillColor: Colors.white,
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(10),
                  borderSide: const BorderSide(color: Color(0xFFE2E8F0))),
            ),
          ),
          const SizedBox(height: 10),

          // Select All / Clear
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Terpilih: ${_targetRoles.length} role/jabatan',
                style: GoogleFonts.plusJakartaSans(
                    fontSize: 11, fontWeight: FontWeight.w700, color: const Color(0xFF475569)),
              ),
              Row(
                children: [
                  TextButton(
                    onPressed: () {
                      setState(() {
                        if (_roleSubTabIndex == 0) {
                          for (final pos in _filteredJobPositions) {
                            final t = pos['title']?.toString();
                            if (t != null && !_targetRoles.contains(t)) {
                              _targetRoles.add(t);
                            }
                          }
                        } else {
                          for (final r in _filteredAccessRoles) {
                            final name = (r['name'] ?? '').toString().toLowerCase().replaceAll(' ', '_');
                            if (name.isNotEmpty && !_targetRoles.contains(name)) {
                              _targetRoles.add(name);
                            }
                          }
                        }
                      });
                    },
                    child: Text('Pilih Semua',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 11, fontWeight: FontWeight.w800, color: AppColors.blueHRDark)),
                  ),
                  TextButton(
                    onPressed: () => setState(() => _targetRoles.clear()),
                    child: Text('Hapus',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 11, fontWeight: FontWeight.w800, color: AppColors.error)),
                  ),
                ],
              ),
            ],
          ),

          // List Items
          ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 220),
            child: ListView(
              shrinkWrap: true,
              children: _roleSubTabIndex == 0
                  ? _filteredJobPositions.map((pos) {
                      final title = pos['title']?.toString() ?? '';
                      final isChecked = _targetRoles.contains(title);
                      return CheckboxListTile(
                        value: isChecked,
                        dense: true,
                        title: Text(title,
                            style: GoogleFonts.plusJakartaSans(
                                fontSize: 13, fontWeight: FontWeight.w700)),
                        subtitle: Text(pos['department']?['name'] ?? '-',
                            style: GoogleFonts.plusJakartaSans(fontSize: 11, color: Colors.grey)),
                        activeColor: AppColors.blueHRDark,
                        contentPadding: EdgeInsets.zero,
                        onChanged: (val) {
                          setState(() {
                            if (val == true) {
                              _targetRoles.add(title);
                            } else {
                              _targetRoles.remove(title);
                            }
                          });
                        },
                      );
                    }).toList()
                  : _filteredAccessRoles.map((r) {
                      final roleKey =
                          (r['name'] ?? '').toString().toLowerCase().replaceAll(' ', '_');
                      final roleName = r['name']?.toString() ?? '';
                      final isChecked = _targetRoles.contains(roleKey) ||
                          _targetRoles.contains(roleName);
                      return CheckboxListTile(
                        value: isChecked,
                        dense: true,
                        title: Text(roleName,
                            style: GoogleFonts.plusJakartaSans(
                                fontSize: 13, fontWeight: FontWeight.w700)),
                        activeColor: AppColors.blueHRDark,
                        contentPadding: EdgeInsets.zero,
                        onChanged: (val) {
                          setState(() {
                            if (val == true) {
                              _targetRoles.add(roleKey);
                            } else {
                              _targetRoles.remove(roleKey);
                              _targetRoles.remove(roleName);
                            }
                          });
                        },
                      );
                    }).toList(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildIndividualTargetPicker() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                'Pilih Karyawan Target Direct',
                style: GoogleFonts.plusJakartaSans(
                    fontSize: 13, fontWeight: FontWeight.w800, color: const Color(0xFF0F172A)),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: AppColors.error.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '${_targetUserIds.length} Karyawan Dipilih',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11, fontWeight: FontWeight.w800, color: AppColors.error),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Employee Search
          TextFormField(
            controller: _employeeSearchController,
            onChanged: (_) => setState(() {}),
            style: GoogleFonts.plusJakartaSans(fontSize: 12),
            decoration: InputDecoration(
              hintText: 'Cari Nama, NIK, atau Departemen...',
              prefixIcon: const Icon(Icons.search_rounded, size: 18),
              isDense: true,
              filled: true,
              fillColor: Colors.white,
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(10),
                  borderSide: const BorderSide(color: Color(0xFFE2E8F0))),
            ),
          ),
          const SizedBox(height: 10),

          // Select All / Clear
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Menampilkan ${_filteredEmployees.length} karyawan',
                style: GoogleFonts.plusJakartaSans(
                    fontSize: 11, fontWeight: FontWeight.w600, color: const Color(0xFF64748B)),
              ),
              Row(
                children: [
                  TextButton(
                    onPressed: () {
                      setState(() {
                        for (final emp in _filteredEmployees) {
                          final uid =
                              (emp['user_id'] ?? emp['user']?['id'] ?? emp['id'])?.toString();
                          if (uid != null && !_targetUserIds.contains(uid)) {
                            _targetUserIds.add(uid);
                          }
                        }
                      });
                    },
                    child: Text('Pilih Semua',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 11, fontWeight: FontWeight.w800, color: AppColors.blueHRDark)),
                  ),
                  TextButton(
                    onPressed: () => setState(() => _targetUserIds.clear()),
                    child: Text('Hapus',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 11, fontWeight: FontWeight.w800, color: AppColors.error)),
                  ),
                ],
              ),
            ],
          ),

          // List Employees
          ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 260),
            child: ListView.separated(
              shrinkWrap: true,
              itemCount: _filteredEmployees.length,
              separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFFE2E8F0)),
              itemBuilder: (context, index) {
                final emp = _filteredEmployees[index];
                final uid = (emp['user_id'] ?? emp['user']?['id'] ?? emp['id'])?.toString() ?? '';
                final name = '${emp['first_name'] ?? ''} ${emp['last_name'] ?? ''}'.trim();
                final nik = emp['employee_id_number'] ?? '-';
                final job = emp['job_position']?['title'] ?? '-';
                final dept = emp['department']?['name'] ?? '-';
                final faceUrl = emp['face_reference_url']?.toString();
                final isSelected = _targetUserIds.contains(uid);

                return CheckboxListTile(
                  value: isSelected,
                  activeColor: AppColors.blueHRDark,
                  dense: true,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                  secondary: CircleAvatar(
                    radius: 18,
                    backgroundColor: AppColors.blueHRDark.withOpacity(0.1),
                    backgroundImage: (faceUrl != null && faceUrl.isNotEmpty)
                        ? CachedNetworkImageProvider(AppConstants.buildFaceImageUrl(faceUrl))
                        : null,
                    child: (faceUrl == null || faceUrl.isEmpty)
                        ? Text(
                            name.isNotEmpty ? name[0].toUpperCase() : '?',
                            style: GoogleFonts.plusJakartaSans(
                                fontWeight: FontWeight.w800,
                                fontSize: 13,
                                color: AppColors.blueHRDark),
                          )
                        : null,
                  ),
                  title: Text(
                    name,
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w800,
                        color: const Color(0xFF1E293B)),
                  ),
                  subtitle: Text(
                    '$nik • $job • $dept',
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 11, color: const Color(0xFF64748B)),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  onChanged: (val) {
                    setState(() {
                      if (val == true) {
                        _targetUserIds.add(uid);
                      } else {
                        _targetUserIds.remove(uid);
                      }
                    });
                  },
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _pickPopupExpiryDateTime() async {
    final now = DateTime.now();
    final pickedDate = await showDatePicker(
      context: context,
      initialDate: _popupExpiresAt ?? now.add(const Duration(days: 3)),
      firstDate: now,
      lastDate: now.add(const Duration(days: 365)),
    );

    if (pickedDate == null || !mounted) return;

    final pickedTime = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(_popupExpiresAt ?? now),
    );

    if (pickedTime == null || !mounted) return;

    setState(() {
      _popupExpiresAt = DateTime(
        pickedDate.year,
        pickedDate.month,
        pickedDate.day,
        pickedTime.hour,
        pickedTime.minute,
      );
    });
  }

  InputDecoration _buildInputDecoration({required String hint, required IconData icon}) {
    return InputDecoration(
      hintText: hint,
      hintStyle: GoogleFonts.plusJakartaSans(fontSize: 13, color: const Color(0xFF94A3B8)),
      prefixIcon: Icon(icon, size: 20, color: const Color(0xFF94A3B8)),
      filled: true,
      fillColor: const Color(0xFFF8FAFC),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: Color(0xFFE2E8F0))),
      enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: Color(0xFFE2E8F0))),
      focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: const BorderSide(color: AppColors.blueHRDark, width: 1.5)),
    );
  }
}

// ─────────────────────────────────────────────────────────────────
// ANNOUNCEMENT TRACKING SHEET (CONFIRMATION / ACKNOWLEDGE STATS)
// ─────────────────────────────────────────────────────────────────
class AnnouncementTrackingSheet extends StatefulWidget {
  final String announcementId;
  final String announcementTitle;
  final ApiClient apiClient;

  const AnnouncementTrackingSheet({
    super.key,
    required this.announcementId,
    required this.announcementTitle,
    required this.apiClient,
  });

  @override
  State<AnnouncementTrackingSheet> createState() => _AnnouncementTrackingSheetState();
}

class _AnnouncementTrackingSheetState extends State<AnnouncementTrackingSheet> {
  bool _isLoading = true;
  Map<String, dynamic>? _summary;
  String _filter = 'all'; // 'all' | 'confirmed' | 'pending'
  final TextEditingController _searchController = TextEditingController();
  final Map<String, bool> _isResendingSingle = {};

  @override
  void initState() {
    super.initState();
    _fetchConfirmations();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _fetchConfirmations() async {
    setState(() => _isLoading = true);
    try {
      final res = await widget.apiClient.client
          .get('admin/announcements/${widget.announcementId}/confirmations');
      if (res.statusCode == 200 && mounted) {
        setState(() {
          _summary = res.data['data'];
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        SnackBarUtils.showError(context, 'Gagal mengambil data tracking konfirmasi: $e');
      }
    }
  }

  Future<void> _resendFCM({required String mode, String? userId, String? userName}) async {
    final key = userId ?? mode;
    setState(() => _isResendingSingle[key] = true);
    try {
      final payload = <String, dynamic>{'mode': mode};
      if (userId != null) payload['target_user_id'] = userId;

      final res = await widget.apiClient.client.post(
          'admin/announcements/${widget.announcementId}/resend-fcm',
          data: payload);

      final count = res.data['data']?['recipients_count'] ?? 0;
      if (mounted) {
        if (mode == 'single') {
          SnackBarUtils.showSuccess(
              context, 'Notifikasi FCM berhasil dikirim ke ${userName ?? 'karyawan'}');
        } else if (mode == 'pending_only') {
          if (count > 0) {
            SnackBarUtils.showSuccess(
                context, 'Notifikasi berhasil dikirim ulang ke $count karyawan yang belum konfirmasi');
          } else {
            SnackBarUtils.showInfo(context, 'Semua karyawan target telah mengonfirmasi');
          }
        }
        _fetchConfirmations();
      }
    } catch (e) {
      if (mounted) {
        SnackBarUtils.showError(context, 'Gagal mengirim ulang FCM: $e');
      }
    } finally {
      if (mounted) {
        setState(() => _isResendingSingle[key] = false);
      }
    }
  }

  List<dynamic> get _filteredRecipients {
    final list = (_summary?['recipients'] as List<dynamic>? ?? []);
    var filtered = list;

    if (_filter == 'confirmed') {
      filtered = filtered.where((r) => r['is_confirmed'] == true).toList();
    } else if (_filter == 'pending') {
      filtered = filtered.where((r) => r['is_confirmed'] != true).toList();
    }

    final q = _searchController.text.trim().toLowerCase();
    if (q.isNotEmpty) {
      filtered = filtered.where((r) {
        final name = (r['employee_name'] ?? '').toString().toLowerCase();
        final nik = (r['employee_id_number'] ?? '').toString().toLowerCase();
        final dept = (r['department_name'] ?? '').toString().toLowerCase();
        final job = (r['job_title'] ?? '').toString().toLowerCase();
        return name.contains(q) || nik.contains(q) || dept.contains(q) || job.contains(q);
      }).toList();
    }

    return filtered;
  }

  @override
  Widget build(BuildContext context) {
    final total = _summary?['total_recipients'] ?? 0;
    final confirmed = _summary?['confirmed_count'] ?? 0;
    final pending = _summary?['pending_count'] ?? 0;
    final percent = total > 0 ? ((confirmed / total) * 100).toInt() : 0;

    return Container(
      height: MediaQuery.of(context).size.height * 0.88,
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      child: Column(
        children: [
          // Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: Color(0xFFF1F5F9))),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Tracking Konfirmasi Penerima',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 16, fontWeight: FontWeight.w900, color: const Color(0xFF0F172A)),
                      ),
                      Text(
                        widget.announcementTitle,
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 12, color: const Color(0xFF64748B)),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close_rounded, size: 20, color: Color(0xFF64748B)),
                  onPressed: () => Navigator.pop(context),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
          ),

          // Body
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator(color: AppColors.blueHRDark))
                : Padding(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      children: [
                        // Stat Cards Row
                        Row(
                          children: [
                            _buildStatCard('TOTAL TARGET', '$total', const Color(0xFF3B82F6), Icons.people_outline_rounded),
                            const SizedBox(width: 8),
                            _buildStatCard('SUDAH BACA', '$confirmed ($percent%)', const Color(0xFF10B981), Icons.check_circle_outline_rounded),
                            const SizedBox(width: 8),
                            _buildStatCard('PENDING', '$pending', const Color(0xFFEF4444), Icons.hourglass_top_rounded),
                          ],
                        ),
                        const SizedBox(height: 14),

                        // Resend Pending Button Banner
                        if (pending > 0)
                          SizedBox(
                            width: double.infinity,
                            child: ElevatedButton.icon(
                              onPressed: _isResendingSingle['pending_only'] == true
                                  ? null
                                  : () => _resendFCM(mode: 'pending_only'),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: const Color(0xFFFFFBEB),
                                foregroundColor: const Color(0xFFB45309),
                                elevation: 0,
                                side: const BorderSide(color: Color(0xFFFDE68A)),
                                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                                padding: const EdgeInsets.symmetric(vertical: 12),
                              ),
                              icon: _isResendingSingle['pending_only'] == true
                                  ? const SizedBox(
                                      width: 14,
                                      height: 14,
                                      child: CircularProgressIndicator(strokeWidth: 2, color: Color(0xFFB45309)),
                                    )
                                  : const Icon(Icons.send_rounded, size: 16),
                              label: Text(
                                'Kirim Ulang FCM ke $pending Karyawan Pending',
                                style: GoogleFonts.plusJakartaSans(
                                    fontSize: 12, fontWeight: FontWeight.w800),
                              ),
                            ),
                          ),
                        const SizedBox(height: 14),

                        // Filter Segmented Buttons & Search
                        Row(
                          children: [
                            _buildFilterChip('all', 'Semua ($total)'),
                            const SizedBox(width: 6),
                            _buildFilterChip('confirmed', 'Sudah ($confirmed)'),
                            const SizedBox(width: 6),
                            _buildFilterChip('pending', 'Pending ($pending)'),
                          ],
                        ),
                        const SizedBox(height: 10),

                        TextFormField(
                          controller: _searchController,
                          onChanged: (_) => setState(() {}),
                          style: GoogleFonts.plusJakartaSans(fontSize: 12),
                          decoration: InputDecoration(
                            hintText: 'Cari penerima...',
                            prefixIcon: const Icon(Icons.search_rounded, size: 18),
                            isDense: true,
                            filled: true,
                            fillColor: const Color(0xFFF8FAFC),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                            border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(10),
                                borderSide: const BorderSide(color: Color(0xFFE2E8F0))),
                          ),
                        ),
                        const SizedBox(height: 10),

                        // List of recipients
                        Expanded(
                          child: _filteredRecipients.isEmpty
                              ? Center(
                                  child: Text('Tidak ada data penerima',
                                      style: GoogleFonts.plusJakartaSans(color: Colors.grey)),
                                )
                              : ListView.separated(
                                  itemCount: _filteredRecipients.length,
                                  separatorBuilder: (_, __) =>
                                      const Divider(height: 1, color: Color(0xFFF1F5F9)),
                                  itemBuilder: (context, index) {
                                    final rec = _filteredRecipients[index];
                                    final name = rec['employee_name'] ?? 'Staff';
                                    final nik = rec['employee_id_number'] ?? '-';
                                    final dept = rec['department_name'] ?? '-';
                                    final isConfirmed = rec['is_confirmed'] == true;
                                    final confirmedAtStr = rec['confirmed_at']?.toString();
                                    final userId = rec['user_id']?.toString() ?? '';
                                    final isSendingThis = _isResendingSingle[userId] == true;

                                    DateTime? confirmedAt;
                                    if (confirmedAtStr != null) {
                                      try {
                                        confirmedAt = DateTime.parse(confirmedAtStr).toLocal();
                                      } catch (_) {}
                                    }

                                    return ListTile(
                                      contentPadding: const EdgeInsets.symmetric(
                                          horizontal: 4, vertical: 4),
                                      leading: CircleAvatar(
                                        radius: 18,
                                        backgroundColor: isConfirmed
                                            ? Colors.green.shade50
                                            : Colors.red.shade50,
                                        child: Icon(
                                          isConfirmed
                                              ? Icons.check_circle_rounded
                                              : Icons.schedule_rounded,
                                          color: isConfirmed
                                              ? Colors.green.shade700
                                              : Colors.red.shade700,
                                          size: 18,
                                        ),
                                      ),
                                      title: Text(
                                        name,
                                        style: GoogleFonts.plusJakartaSans(
                                            fontSize: 13,
                                            fontWeight: FontWeight.w800,
                                            color: const Color(0xFF1E293B)),
                                      ),
                                      subtitle: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Text('$nik • $dept',
                                              style: GoogleFonts.plusJakartaSans(
                                                  fontSize: 11, color: const Color(0xFF64748B))),
                                          if (isConfirmed && confirmedAt != null)
                                            Text(
                                              'Dikonfirmasi: ${DateFormat('dd MMM, HH:mm').format(confirmedAt)}',
                                              style: GoogleFonts.plusJakartaSans(
                                                  fontSize: 10,
                                                  fontWeight: FontWeight.w600,
                                                  color: Colors.green.shade800),
                                            ),
                                        ],
                                      ),
                                      trailing: !isConfirmed
                                          ? IconButton(
                                              icon: isSendingThis
                                                  ? const SizedBox(
                                                      width: 14,
                                                      height: 14,
                                                      child: CircularProgressIndicator(strokeWidth: 2),
                                                    )
                                                  : const Icon(Icons.send_rounded,
                                                      size: 16, color: AppColors.blueHRDark),
                                              tooltip: 'Kirim Ulang FCM',
                                              onPressed: isSendingThis
                                                  ? null
                                                  : () => _resendFCM(
                                                      mode: 'single',
                                                      userId: userId,
                                                      userName: name),
                                            )
                                          : null,
                                    );
                                  },
                                ),
                        ),
                      ],
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatCard(String label, String value, Color color, IconData icon) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 12),
        decoration: BoxDecoration(
          color: color.withOpacity(0.08),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: color.withOpacity(0.2)),
        ),
        child: Column(
          children: [
            Icon(icon, size: 18, color: color),
            const SizedBox(height: 6),
            Text(
              value,
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 14, fontWeight: FontWeight.w900, color: color),
            ),
            Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 9, fontWeight: FontWeight.w800, color: const Color(0xFF64748B)),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFilterChip(String key, String label) {
    final isSelected = _filter == key;
    return Expanded(
      child: InkWell(
        onTap: () => setState(() => _filter = key),
        borderRadius: BorderRadius.circular(10),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 8),
          decoration: BoxDecoration(
            color: isSelected ? AppColors.blueHRDark : const Color(0xFFF1F5F9),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Center(
            child: Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  color: isSelected ? Colors.white : const Color(0xFF475569)),
            ),
          ),
        ),
      ),
    );
  }
}

// ─────────────────────────────────────────────────────────────────
// RESEND FCM ACTION SHEET
// ─────────────────────────────────────────────────────────────────
class ResendFcmSheet extends StatefulWidget {
  final Map<String, dynamic> announcement;
  final List<dynamic> allEmployees;
  final ApiClient apiClient;
  final VoidCallback onSuccess;

  const ResendFcmSheet({
    super.key,
    required this.announcement,
    required this.allEmployees,
    required this.apiClient,
    required this.onSuccess,
  });

  @override
  State<ResendFcmSheet> createState() => _ResendFcmSheetState();
}

class _ResendFcmSheetState extends State<ResendFcmSheet> {
  bool _isSending = false;

  Future<void> _send(String mode) async {
    setState(() => _isSending = true);
    try {
      final res = await widget.apiClient.client.post(
        'admin/announcements/${widget.announcement['id']}/resend-fcm',
        data: {'mode': mode},
      );

      final count = res.data['data']?['recipients_count'] ?? 0;
      if (mounted) {
        Navigator.pop(context);
        if (mode == 'all') {
          SnackBarUtils.showSuccess(
              context, 'Notifikasi FCM berhasil disiarkan ulang ke seluruh target ($count penerima)');
        } else {
          if (count > 0) {
            SnackBarUtils.showSuccess(
                context, 'Notifikasi FCM berhasil dikirim ulang ke $count karyawan yang belum konfirmasi');
          } else {
            SnackBarUtils.showInfo(context, 'Semua karyawan target telah mengonfirmasi pengumuman ini');
          }
        }
        widget.onSuccess();
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isSending = false);
        SnackBarUtils.showError(context, 'Gagal mengirim ulang FCM: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final title = widget.announcement['title'] ?? '-';

    return SafeArea(
      child: Container(
        padding: const EdgeInsets.all(24),
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
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
                    color: Colors.grey.shade300, borderRadius: BorderRadius.circular(10)),
              ),
            ),
            const SizedBox(height: 20),
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: AppColors.blueHRDark.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(Icons.send_rounded, color: AppColors.blueHRDark, size: 22),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Kirim Ulang Notifikasi FCM',
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 16, fontWeight: FontWeight.w900, color: const Color(0xFF0F172A)),
                      ),
                      Text(
                        title,
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 12, color: const Color(0xFF64748B)),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 24),
            Text(
              'PILIH MODE PENGIRIMAN ULANG',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                  color: const Color(0xFF64748B),
                  letterSpacing: 0.8),
            ),
            const SizedBox(height: 12),

            // Option 1: All
            _buildOptionTile(
              title: 'Kirim Ulang ke Seluruh Target Penerima',
              subtitle: 'Broadcast ulang notifikasi ke semua staff yang masuk dalam target pengumuman',
              icon: Icons.campaign_rounded,
              color: const Color(0xFF3B82F6),
              onTap: _isSending ? null : () => _send('all'),
            ),
            const SizedBox(height: 10),

            // Option 2: Pending Only
            _buildOptionTile(
              title: 'Kirim Khusus ke yang Belum Konfirmasi',
              subtitle: 'Hanya kirimkan reminder push notification ke staff yang statusnya masih pending',
              icon: Icons.hourglass_top_rounded,
              color: const Color(0xFFF59E0B),
              onTap: _isSending ? null : () => _send('pending_only'),
            ),
            const SizedBox(height: 20),

            if (_isSending)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(8),
                  child: CircularProgressIndicator(color: AppColors.blueHRDark),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildOptionTile({
    required String title,
    required String subtitle,
    required IconData icon,
    required Color color,
    required VoidCallback? onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: color.withOpacity(0.06),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withOpacity(0.25)),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: color.withOpacity(0.12),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: color, size: 20),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 13, fontWeight: FontWeight.w800, color: const Color(0xFF1E293B)),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 11, color: const Color(0xFF64748B), height: 1.3),
                  ),
                ],
              ),
            ),
            Icon(Icons.chevron_right_rounded, color: color, size: 20),
          ],
        ),
      ),
    );
  }
}
