import 'dart:io';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/features/reimbursement/presentation/pages/add_reimbursement_page.dart';
import 'package:hris_app/features/sales/presentation/pages/edit_journey_page.dart';
import 'package:hris_app/injection.dart' as di;

class SalesJourneyHistoryPage extends StatefulWidget {
  const SalesJourneyHistoryPage({super.key});

  @override
  State<SalesJourneyHistoryPage> createState() => _SalesJourneyHistoryPageState();
}

class _SalesJourneyHistoryPageState extends State<SalesJourneyHistoryPage> {
  final _apiService = SalesApiService(apiClient: di.sl<ApiClient>());
  bool _isLoading = true;
  DateTime _selectedMonth = DateTime(DateTime.now().year, DateTime.now().month, 1);
  List<dynamic> _journeys = [];

  @override
  void initState() {
    super.initState();
    _fetchHistory();
  }

  Future<void> _fetchHistory() async {
    setState(() => _isLoading = true);
    try {
      final startOfMonth = DateTime(_selectedMonth.year, _selectedMonth.month, 1);
      final nextMonth = DateTime(_selectedMonth.year, _selectedMonth.month + 1, 1);
      final endOfMonth = nextMonth.subtract(const Duration(days: 1));

      final startStr = DateFormat('yyyy-MM-dd').format(startOfMonth);
      final endStr = DateFormat('yyyy-MM-dd').format(endOfMonth);

      final data = await _apiService.getJourneyHistory(
        startDate: startStr,
        endDate: endStr,
      );

      if (mounted) {
        setState(() {
          _journeys = data;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  void _previousMonth() {
    setState(() {
      _selectedMonth = DateTime(_selectedMonth.year, _selectedMonth.month - 1, 1);
    });
    _fetchHistory();
  }

  void _nextMonth() {
    setState(() {
      _selectedMonth = DateTime(_selectedMonth.year, _selectedMonth.month + 1, 1);
    });
    _fetchHistory();
  }

  double get _totalKM {
    double sum = 0.0;
    for (var j in _journeys) {
      final total = (j['total_km'] as num?)?.toDouble() ?? 0.0;
      sum += total;
    }
    return sum;
  }

  int get _totalDays {
    return _journeys.length;
  }

  double get _totalBBM {
    double sum = 0.0;
    for (var j in _journeys) {
      final bbm = (j['total_bbm_amount'] as num?)?.toDouble() ?? 0.0;
      sum += bbm;
    }
    return sum;
  }

  int get _totalVisits {
    int sum = 0;
    for (var j in _journeys) {
      final visits = (j['total_visits'] as num?)?.toInt() ?? 0;
      sum += visits;
    }
    return sum;
  }

  String _formatCurrency(double amount) {
    final formatter = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    return formatter.format(amount);
  }

  bool _isLocalFile(String path) {
    if (path.isEmpty) return false;
    if (path.startsWith('http')) return false;
    if (path.startsWith('/uploads/') || path.startsWith('/api/v1/uploads/')) return false;
    if (path.startsWith('/data/') ||
        path.startsWith('/var/') ||
        path.startsWith('/storage/') ||
        path.startsWith('/private/')) {
      return true;
    }
    // Guard dengan try-catch agar tidak melempar _Namespace di path invalid / platform web
    try {
      return File(path).existsSync();
    } catch (_) {
      return false;
    }
  }

  Widget _buildSmartImage(String rawPath, {double? height, double? width, BoxFit fit = BoxFit.cover}) {
    if (rawPath.isEmpty) {
      return Container(
        height: height,
        width: width,
        color: Colors.grey.shade100,
        child: const Center(
          child: Icon(Icons.speed_rounded, size: 24, color: Colors.grey),
        ),
      );
    }

    if (_isLocalFile(rawPath)) {
      final file = File(rawPath);
      if (file.existsSync()) {
        return Image.file(
          file,
          height: height,
          width: width,
          fit: fit,
          errorBuilder: (_, __, ___) => _buildImageError(height, width, isLocal: true),
        );
      } else {
        return _buildImageError(height, width, isLocal: true);
      }
    }

    final resolvedUrl = AppConstants.buildImageUrl(rawPath);
    return Image.network(
      resolvedUrl,
      height: height,
      width: width,
      fit: fit,
      loadingBuilder: (context, child, progress) {
        if (progress == null) return child;
        return Container(
          height: height,
          width: width,
          color: Colors.grey.shade100,
          child: const Center(
            child: SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.primaryGreen),
            ),
          ),
        );
      },
      errorBuilder: (_, __, ___) => _buildImageError(height, width),
    );
  }

  Widget _buildImageError(double? height, double? width, {bool isLocal = false}) {
    return Container(
      height: height,
      width: width,
      color: Colors.grey.shade100,
      padding: const EdgeInsets.all(6),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(isLocal ? Icons.save_alt_rounded : Icons.broken_image_rounded, size: 18, color: Colors.blueGrey.shade400),
            const SizedBox(height: 2),
            Text(
              isLocal ? 'Tersimpan lokal' : 'Gagal memuat',
              style: GoogleFonts.outfit(color: Colors.blueGrey.shade600, fontSize: 8.5, fontWeight: FontWeight.w600),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPreviewError({bool isLocal = false}) {
    return Container(
      height: 220,
      color: Colors.grey.shade900,
      padding: const EdgeInsets.all(20),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(isLocal ? Icons.save_alt_rounded : Icons.broken_image_rounded, color: Colors.white54, size: 48),
            const SizedBox(height: 10),
            Text(
              isLocal
                  ? 'Foto tersimpan di penyimpanan lokal perangkat saat input'
                  : 'Gagal memuat foto dari server',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(color: Colors.white70, fontSize: 13, fontWeight: FontWeight.w600),
            ),
          ],
        ),
      ),
    );
  }

  void _showImagePreview(BuildContext context, String rawPath, String title, String subtitle) {
    if (rawPath.isEmpty) return;
    showDialog(
      context: context,
      builder: (ctx) => Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
        child: Container(
          decoration: BoxDecoration(
            color: const Color(0xFF1E293B),
            borderRadius: BorderRadius.circular(24),
          ),
          clipBehavior: Clip.antiAlias,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
                color: const Color(0xFF0F172A),
                child: Row(
                  children: [
                    const Icon(Icons.speed_rounded, color: Color(0xFF38BDF8), size: 20),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            title,
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w700,
                              fontSize: 15,
                              color: Colors.white,
                            ),
                          ),
                          if (subtitle.isNotEmpty)
                            Text(
                              subtitle,
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: Colors.blueGrey.shade200,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                        ],
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close_rounded, color: Colors.white70),
                      onPressed: () => Navigator.pop(ctx),
                      visualDensity: VisualDensity.compact,
                    ),
                  ],
                ),
              ),
              Flexible(
                child: InteractiveViewer(
                  maxScale: 4.0,
                  child: _isLocalFile(rawPath) && File(rawPath).existsSync()
                      ? Image.file(
                          File(rawPath),
                          fit: BoxFit.contain,
                          errorBuilder: (_, __, ___) => _buildPreviewError(isLocal: true),
                        )
                      : Image.network(
                          AppConstants.buildImageUrl(rawPath),
                          fit: BoxFit.contain,
                          loadingBuilder: (context, child, progress) {
                            if (progress == null) return child;
                            return const SizedBox(
                              height: 250,
                              child: Center(
                                child: CircularProgressIndicator(color: AppColors.primaryGreen),
                              ),
                            );
                          },
                          errorBuilder: (_, __, ___) => _buildPreviewError(isLocal: false),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final monthFormat = DateFormat('MMMM yyyy', 'id_ID');

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Riwayat Odometer & KM',
          style: GoogleFonts.outfit(fontWeight: FontWeight.w800, fontSize: 18),
        ),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF0F172A),
        elevation: 0,
        centerTitle: true,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded),
            tooltip: 'Segarkan',
            onPressed: _fetchHistory,
          ),
        ],
      ),
      body: Column(
        children: [
          // Month Selector
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            color: Colors.white,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: const Color(0xFFF1F5F9),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  IconButton(
                    icon: const Icon(Icons.chevron_left_rounded, color: Color(0xFF334155)),
                    onPressed: _previousMonth,
                    tooltip: 'Bulan Sebelumnya',
                  ),
                  Row(
                    children: [
                      const Icon(Icons.calendar_month_rounded, size: 18, color: Color(0xFF0284C7)),
                      const SizedBox(width: 8),
                      Text(
                        monthFormat.format(_selectedMonth),
                        style: GoogleFonts.outfit(
                          fontSize: 15,
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF0F172A),
                        ),
                      ),
                    ],
                  ),
                  IconButton(
                    icon: const Icon(Icons.chevron_right_rounded, color: Color(0xFF334155)),
                    onPressed: _nextMonth,
                    tooltip: 'Bulan Berikutnya',
                  ),
                ],
              ),
            ),
          ),

          // Summary Header Cards
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.03),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: GridView.count(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              crossAxisCount: 2,
              crossAxisSpacing: 10,
              mainAxisSpacing: 10,
              childAspectRatio: 2.0,
              children: [
                _buildSummaryTile(
                  icon: Icons.speed_rounded,
                  iconColor: const Color(0xFF0284C7),
                  bgColor: const Color(0xFFE0F2FE),
                  title: 'Total Jarak',
                  value: '${_totalKM.toStringAsFixed(1)} KM',
                ),
                _buildSummaryTile(
                  icon: Icons.calendar_today_rounded,
                  iconColor: const Color(0xFF10B981),
                  bgColor: const Color(0xFFD1FAE5),
                  title: 'Hari Perjalanan',
                  value: '$_totalDays Hari',
                ),
                _buildSummaryTile(
                  icon: Icons.local_gas_station_rounded,
                  iconColor: const Color(0xFFF59E0B),
                  bgColor: const Color(0xFFFEF3C7),
                  title: 'Klaim BBM',
                  value: _formatCurrency(_totalBBM),
                ),
                _buildSummaryTile(
                  icon: Icons.storefront_rounded,
                  iconColor: const Color(0xFF8B5CF6),
                  bgColor: const Color(0xFFEDE9FE),
                  title: 'Kunjungan Toko',
                  value: '$_totalVisits Outlet',
                ),
              ],
            ),
          ),

          const SizedBox(height: 8),

          // List of Journeys
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(color: AppColors.primaryGreen),
                  )
                : _journeys.isEmpty
                    ? _buildEmptyState()
                    : RefreshIndicator(
                        color: AppColors.primaryGreen,
                        onRefresh: _fetchHistory,
                        child: ListView.separated(
                          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                          itemCount: _journeys.length,
                          separatorBuilder: (_, __) => const SizedBox(height: 12),
                          itemBuilder: (context, index) {
                            final raw = _journeys[index];
                            if (raw == null || raw is! Map<String, dynamic>) {
                              return const SizedBox.shrink();
                            }
                            return _buildJourneyCard(raw);
                          },
                        ),
                      ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryTile({
    required IconData icon,
    required Color iconColor,
    required Color bgColor,
    required String title,
    required String value,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: bgColor.withOpacity(0.4),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: iconColor.withOpacity(0.15)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: bgColor,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(icon, color: iconColor, size: 20),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  title,
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: const Color(0xFF64748B),
                  ),
                ),
                Text(
                  value,
                  style: GoogleFonts.outfit(
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFF0F172A),
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: const BoxDecoration(
                color: Color(0xFFF1F5F9),
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.speed_rounded, size: 56, color: Color(0xFF94A3B8)),
            ),
            const SizedBox(height: 16),
            Text(
              'Belum Ada Catatan Perjalanan',
              style: GoogleFonts.outfit(
                fontSize: 17,
                fontWeight: FontWeight.w800,
                color: const Color(0xFF334155),
              ),
            ),
            const SizedBox(height: 6),
            Text(
              'Belum ada data input KM awal / pulang untuk bulan ini.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 13, color: const Color(0xFF64748B)),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildJourneyCard(Map<String, dynamic> journey) {
    final dateStr = journey['date'] as String? ?? '';
    DateTime? date;
    try {
      date = DateTime.parse(dateStr);
    } catch (_) {}

    final formattedDate = date != null
        ? DateFormat('EEEE, dd MMM yyyy', 'id_ID').format(date)
        : dateStr;

    final status = journey['status'] as String? ?? 'STARTED';
    final isEnded = status == 'ENDED';

    final startKM = (journey['start_km'] as num?)?.toDouble() ?? 0.0;
    final endKM = (journey['end_km'] as num?)?.toDouble() ?? 0.0;
    final totalKM = (journey['total_km'] as num?)?.toDouble() ?? (isEnded ? endKM - startKM : 0.0);

    final startPhoto = journey['start_photo_url'] as String? ?? '';
    final endPhoto = journey['end_photo_url'] as String? ?? '';

    final startLocation = journey['start_location'] as String? ?? '';
    final endLocation = journey['end_location'] as String? ?? '';

    final startTimeStr = journey['start_time'] as String?;
    final endTimeStr = journey['end_time'] as String?;

    String formattedStartTime = '-';
    if (startTimeStr != null) {
      try {
        final st = DateTime.parse(startTimeStr).toLocal();
        formattedStartTime = DateFormat('HH:mm').format(st);
      } catch (_) {}
    }

    String formattedEndTime = '-';
    if (endTimeStr != null) {
      try {
        final et = DateTime.parse(endTimeStr).toLocal();
        formattedEndTime = DateFormat('HH:mm').format(et);
      } catch (_) {}
    }

    final totalVisits = (journey['total_visits'] as num?)?.toInt() ?? 0;
    final bbmList = journey['bbm_reimbursements'] as List<dynamic>? ?? [];
    final totalBBMAmount = (journey['total_bbm_amount'] as num?)?.toDouble() ?? 0.0;

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Card Header: Date & Status Badge
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: const BoxDecoration(
              color: Color(0xFFF8FAFC),
              borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
              border: Border(bottom: BorderSide(color: Color(0xFFE2E8F0))),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    const Icon(Icons.event_note_rounded, size: 16, color: Color(0xFF0284C7)),
                    const SizedBox(width: 8),
                    Text(
                      formattedDate,
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w800,
                        fontSize: 14,
                        color: const Color(0xFF0F172A),
                      ),
                    ),
                  ],
                ),
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: isEnded
                            ? const Color(0xFFDCFCE7)
                            : const Color(0xFFE0F2FE),
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(
                          color: isEnded
                              ? const Color(0xFF86EFAC)
                              : const Color(0xFF7DD3FC),
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 6,
                            height: 6,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              color: isEnded
                                  ? const Color(0xFF16A34A)
                                  : const Color(0xFF0284C7),
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            isEnded ? 'Selesai' : 'Sedang Berjalan',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                              color: isEnded
                                  ? const Color(0xFF15803D)
                                  : const Color(0xFF0369A1),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 6),
                    Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (_) => EditJourneyPage(
                                journeyId: journey['id'].toString(),
                                editMode: isEnded ? 'both' : 'start',
                                initialStartKM: startKM,
                                initialEndKM: isEnded ? endKM : null,
                                currentStartPhotoUrl: startPhoto.isNotEmpty ? startPhoto : null,
                                currentEndPhotoUrl: isEnded && endPhoto.isNotEmpty ? endPhoto : null,
                                currentStartLocation: startLocation,
                                currentEndLocation: endLocation,
                                formattedDate: formattedDate,
                              ),
                            ),
                          ).then((result) {
                            if (result == true) {
                              _fetchHistory();
                            }
                          });
                        },
                        borderRadius: BorderRadius.circular(8),
                        child: Container(
                          padding: const EdgeInsets.all(4),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: const Color(0xFFCBD5E1)),
                          ),
                          child: const Icon(
                            Icons.edit_rounded,
                            size: 15,
                            color: Color(0xFF475569),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Card Body: KM Comparison (Berangkat vs Pulang)
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Start KM Block
                    Expanded(
                      child: _buildKMDetailBlock(
                        title: 'KM BERANGKAT',
                        titleColor: const Color(0xFF0284C7),
                        kmValue: '${startKM.toStringAsFixed(1)} KM',
                        time: formattedStartTime,
                        location: startLocation,
                        photoUrl: startPhoto,
                        onPhotoTap: startPhoto.isNotEmpty
                            ? () => _showImagePreview(
                                  context,
                                  startPhoto,
                                  'Speedometer Berangkat ($formattedStartTime)',
                                  startLocation,
                                )
                            : null,
                      ),
                    ),
                    const SizedBox(width: 12),
                    // End KM Block
                    Expanded(
                      child: isEnded
                          ? _buildKMDetailBlock(
                              title: 'KM PULANG',
                              titleColor: const Color(0xFF16A34A),
                              kmValue: '${endKM.toStringAsFixed(1)} KM',
                              time: formattedEndTime,
                              location: endLocation,
                              photoUrl: endPhoto,
                              onPhotoTap: endPhoto.isNotEmpty
                                  ? () => _showImagePreview(
                                        context,
                                        endPhoto,
                                        'Speedometer Pulang ($formattedEndTime)',
                                        endLocation,
                                      )
                                  : null,
                            )
                          : Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: const Color(0xFFF8FAFC),
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'KM PULANG',
                                    style: GoogleFonts.outfit(
                                      fontSize: 10,
                                      fontWeight: FontWeight.w800,
                                      color: const Color(0xFF94A3B8),
                                      letterSpacing: 0.5,
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    'Belum Dicatat',
                                    style: GoogleFonts.outfit(
                                      fontSize: 14,
                                      fontWeight: FontWeight.w700,
                                      color: const Color(0xFF94A3B8),
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    'Perjalanan masih berlangsung',
                                    style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      color: const Color(0xFFCBD5E1),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                    ),
                  ],
                ),

                const SizedBox(height: 14),

                // Distance & Visit metrics banner
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF1F5F9),
                    borderRadius: BorderRadius.circular(14),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.route_rounded, size: 16, color: Color(0xFF475569)),
                          const SizedBox(width: 6),
                          Text(
                            'Jarak Tempuh:',
                            style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: const Color(0xFF475569),
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            isEnded ? '+${totalKM.toStringAsFixed(1)} KM' : 'Dalam proses...',
                            style: GoogleFonts.outfit(
                              fontSize: 13,
                              fontWeight: FontWeight.w900,
                              color: isEnded ? const Color(0xFF0F172A) : const Color(0xFF0284C7),
                            ),
                          ),
                        ],
                      ),
                      Row(
                        children: [
                          const Icon(Icons.store_mall_directory_rounded, size: 16, color: Color(0xFF475569)),
                          const SizedBox(width: 6),
                          Text(
                            '$totalVisits Outlet',
                            style: GoogleFonts.outfit(
                              fontSize: 12,
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF334155),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                // BBM Section
                const SizedBox(height: 10),
                if (bbmList.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFFBEB),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: const Color(0xFFFDE68A)),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.local_gas_station_rounded, size: 16, color: Color(0xFFD97706)),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            'Klaim BBM: ${_formatCurrency(totalBBMAmount)} (${bbmList.length} pengajuan)',
                            style: GoogleFonts.outfit(
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                              color: const Color(0xFF92400E),
                            ),
                          ),
                        ),
                      ],
                    ),
                  )
                else
                  InkWell(
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => AddReimbursementPage(
                            initialTitle: 'BBM Sales - $formattedDate',
                          ),
                        ),
                      ).then((_) => _fetchHistory());
                    },
                    borderRadius: BorderRadius.circular(12),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: const Color(0xFFCBD5E1), style: BorderStyle.solid),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.add_circle_outline_rounded, size: 16, color: Color(0xFF0284C7)),
                          const SizedBox(width: 6),
                          Text(
                            'Klaim BBM untuk tanggal ini',
                            style: GoogleFonts.outfit(
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                              color: const Color(0xFF0284C7),
                            ),
                          ),
                        ],
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

  Widget _buildKMDetailBlock({
    required String title,
    required Color titleColor,
    required String kmValue,
    required String time,
    required String location,
    required String photoUrl,
    VoidCallback? onPhotoTap,
  }) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                title,
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  color: titleColor,
                  letterSpacing: 0.5,
                ),
              ),
              if (time != '-')
                Row(
                  children: [
                    Icon(Icons.access_time_rounded, size: 11, color: Colors.blueGrey.shade400),
                    const SizedBox(width: 3),
                    Text(
                      time,
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: Colors.blueGrey.shade600,
                      ),
                    ),
                  ],
                ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            kmValue,
            style: GoogleFonts.outfit(
              fontSize: 16,
              fontWeight: FontWeight.w900,
              color: const Color(0xFF0F172A),
            ),
          ),
          if (location.isNotEmpty) ...[
            const SizedBox(height: 4),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.location_on_outlined, size: 12, color: Color(0xFF94A3B8)),
                const SizedBox(width: 4),
                Expanded(
                  child: Text(
                    location,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: GoogleFonts.outfit(
                      fontSize: 10,
                      color: const Color(0xFF64748B),
                    ),
                  ),
                ),
              ],
            ),
          ],
          if (photoUrl.isNotEmpty) ...[
            const SizedBox(height: 8),
            GestureDetector(
              onTap: onPhotoTap,
              child: SizedBox(
                height: 70,
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(10),
                  child: Stack(
                    fit: StackFit.expand,
                    children: [
                      _buildSmartImage(
                        photoUrl,
                        height: 70,
                        width: double.infinity,
                        fit: BoxFit.cover,
                      ),
                      Positioned(
                        bottom: 4,
                        right: 4,
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.65),
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(Icons.zoom_in_rounded, color: Colors.white, size: 12),
                              const SizedBox(width: 2),
                              Text(
                                'Lihat Foto',
                                style: GoogleFonts.outfit(color: Colors.white, fontSize: 9, fontWeight: FontWeight.bold),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}
